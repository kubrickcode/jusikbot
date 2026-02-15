#!/usr/bin/env python3
"""Validate portfolio analysis report structure and content rules.

Exit 0: PASS or FAIL — structured JSON on stdout.
Exit 1: ERROR — file access failure.
"""

import argparse
import json
import re
import sys
from typing import NamedTuple, Optional


class ReportError(NamedTuple):
    rule: str
    detail: str
    line: Optional[int] = None


class ValidationResult(NamedTuple):
    status: str
    errors: list[ReportError]


def extract_sections(content: str) -> dict[str, tuple[int, str]]:
    """Parse markdown into {section_name: (start_line, body)} keyed by H2 headers."""
    sections: dict[str, tuple[int, str]] = {}
    current_name: Optional[str] = None
    current_line = 0
    current_body: list[str] = []

    for line_num, line in enumerate(content.split("\n"), 1):
        if line.startswith("## "):
            if current_name is not None:
                sections[current_name] = (current_line, "\n".join(current_body))
            current_name = line[3:].strip()
            current_line = line_num
            current_body = []
        elif current_name is not None:
            current_body.append(line)

    if current_name is not None:
        sections[current_name] = (current_line, "\n".join(current_body))

    return sections


_REQUIRED_SECTIONS = ["요약", "논제별 현황", "배분 제안", "리스크 요인"]


def validate_required_sections(
    sections: dict[str, tuple[int, str]],
) -> list[ReportError]:
    errors = []
    section_names = list(sections.keys())
    for req in _REQUIRED_SECTIONS:
        if not any(req in name for name in section_names):
            errors.append(
                ReportError(
                    rule="required_sections",
                    detail=f"Missing required section: {req}",
                )
            )
    return errors


def validate_thesis_counterarguments(
    sections: dict[str, tuple[int, str]],
) -> list[ReportError]:
    """Each thesis under '논제별 현황' must have non-empty counterargument content."""
    thesis_section = None
    for name, (line_num, body) in sections.items():
        if "논제별 현황" in name:
            thesis_section = (line_num, body)
            break

    if thesis_section is None:
        return []

    base_line, body = thesis_section
    errors = []

    thesis_blocks = re.split(r"(?m)^### ", body)
    for block in thesis_blocks[1:]:
        first_line = block.split("\n", 1)[0].strip()
        thesis_name = first_line.rstrip(":")

        counter_match = re.search(
            r"(?m)^\*\*반론/리스크\*\*:\s*\n(.*?)(?=\n###|\n##|\Z)",
            block,
            re.DOTALL,
        )

        if counter_match:
            counter_content = counter_match.group(1).strip()
            if not counter_content:
                errors.append(
                    ReportError(
                        rule="thesis_counterarguments",
                        detail=f"{thesis_name}: empty counter-argument section",
                    )
                )
        else:
            has_any_counter = any(
                kw in block for kw in ["반론", "리스크", "위험", "주의"]
            )
            if not has_any_counter:
                errors.append(
                    ReportError(
                        rule="thesis_counterarguments",
                        detail=f"{thesis_name}: missing counter-argument section",
                    )
                )

    return errors


_EVIDENCE_CATEGORIES = {"추세", "모멘텀", "변동성", "상대강도", "외부"}
_EVIDENCE_PATTERN = re.compile(r"\[(추세|모멘텀|변동성|상대강도|외부)\]")


def validate_high_confidence_evidence(content: str) -> list[ReportError]:
    """High confidence theses require 2+ evidence from different categories."""
    errors = []

    thesis_blocks = re.split(r"(?m)^### ", content)
    for block in thesis_blocks[1:]:
        first_line = block.split("\n", 1)[0].strip()
        thesis_name = first_line.rstrip(":")

        confidence_match = re.search(
            r"\*{0,2}(?:신뢰도|confidence)\*{0,2}\s*[:：]\s*(High|높음)",
            block,
            re.IGNORECASE,
        )
        if not confidence_match:
            continue

        evidence_tags = _EVIDENCE_PATTERN.findall(block)
        unique_categories = set(evidence_tags)

        if len(unique_categories) < 2:
            errors.append(
                ReportError(
                    rule="high_confidence_evidence",
                    detail=(
                        f"{thesis_name}: High confidence requires 2+ evidence categories, "
                        f"found {len(unique_categories)}"
                    ),
                )
            )

    return errors


_FORBIDDEN_EXPRESSIONS = ["확실히", "반드시", "무조건", "틀림없이", "100%"]


def validate_forbidden_expressions(content: str) -> list[ReportError]:
    errors = []
    for line_num, line in enumerate(content.split("\n"), 1):
        for expr in _FORBIDDEN_EXPRESSIONS:
            if expr in line:
                errors.append(
                    ReportError(
                        rule="forbidden_expression",
                        detail=f"Forbidden expression '{expr}' found",
                        line=line_num,
                    )
                )
    return errors


_NUMBER_PATTERN = re.compile(
    r"\d[\d,]*\.?\d*\s*[%원KMB달러]"
    r"|\d[\d,]*\.?\d*\s*(?:퍼센트|percent)"
    r"|\d{1,3}(?:,\d{3})+",
    re.IGNORECASE,
)
_SOURCE_TAG_PATTERN = re.compile(r"\[(summary|psql|user|사용자|config)\]")


def validate_source_tags(content: str) -> list[ReportError]:
    """Prose lines containing numbers must have at least one source tag."""
    errors = []
    for line_num, line in enumerate(content.split("\n"), 1):
        stripped = line.strip()
        if not stripped:
            continue
        if "|" in stripped:
            continue
        if stripped.startswith("#"):
            continue
        if stripped.startswith(">"):
            continue

        if _NUMBER_PATTERN.search(stripped):
            if not _SOURCE_TAG_PATTERN.search(stripped):
                errors.append(
                    ReportError(
                        rule="source_tag",
                        detail=f"Number without source tag",
                        line=line_num,
                    )
                )
    return errors


_KR_STOCK_PATTERN = re.compile(r"(069500|KODEX)")
_KR_WARNING_PATTERNS = ["수정주가", "adj_close", "배당.*미반영", "분배금.*미반영"]


def validate_kr_data_warning(content: str) -> list[ReportError]:
    """KR stock mentions require data limitation warning somewhere in the report."""
    if not _KR_STOCK_PATTERN.search(content):
        return []

    for kw in _KR_WARNING_PATTERNS:
        if re.search(kw, content):
            return []

    return [
        ReportError(
            rule="kr_data_warning",
            detail="KR stock mentioned but missing adj_close data limitation warning",
        )
    ]


def validate_report(content: str) -> ValidationResult:
    sections = extract_sections(content)
    errors: list[ReportError] = []

    errors.extend(validate_required_sections(sections))
    errors.extend(validate_thesis_counterarguments(sections))
    errors.extend(validate_high_confidence_evidence(content))
    errors.extend(validate_forbidden_expressions(content))
    errors.extend(validate_source_tags(content))
    errors.extend(validate_kr_data_warning(content))

    status = "PASS" if not errors else "FAIL"
    return ValidationResult(status=status, errors=errors)


def _serialize_result(result: ValidationResult) -> dict:
    if result.status == "PASS":
        return {"status": "PASS"}
    return {
        "status": "FAIL",
        "errors": [
            {"rule": e.rule, "detail": e.detail, "line": e.line}
            for e in result.errors
        ],
    }


def _parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Validate portfolio analysis report"
    )
    parser.add_argument("report_path", help="Path to report markdown file")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = _parse_args(argv)

    try:
        with open(args.report_path) as f:
            content = f.read()
    except (FileNotFoundError, OSError) as exc:
        print(
            json.dumps({"status": "ERROR", "detail": str(exc)}),
            file=sys.stderr,
        )
        return 1

    result = validate_report(content)
    print(json.dumps(_serialize_result(result), ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())

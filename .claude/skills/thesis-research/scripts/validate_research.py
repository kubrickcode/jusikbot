#!/usr/bin/env python3
"""Validate thesis-research output files against schemas and cross-references.

Exit 0: PASS or FAIL — structured JSON on stdout.
Exit 1: ERROR — input parsing or file access failure.
"""

import argparse
import json
import re
import sys
from datetime import date, timedelta
from typing import NamedTuple, Optional


_STALENESS_THRESHOLD_DAYS = 7

_VALID_THESIS_STATUSES = {"valid", "weakening", "invalidated"}
_VALID_CONDITION_TYPES = {"validity", "invalidation"}
_VALID_CONDITION_STATUSES = {"met", "partially_met", "not_yet", "refuted", "unknown"}
_VALID_TRANSITIONS = {"stable", "improving", "degrading", "new"}
_VALID_MARKETS = {"US", "KR", "EU"}
_VALID_ASSET_TYPES = {"stock", "etf"}
_VALID_MARKET_CAP_CATEGORIES = {"large", "mid", "small"}
_SOURCE_TIER_RANGE = (1, 4)


class ResearchError(NamedTuple):
    rule: str
    detail: str


class ValidationResult(NamedTuple):
    status: str
    errors: list[ResearchError]


def parse_thesis_names(theses_content: str) -> list[str]:
    """Extract H2 header names from theses.md, excluding non-thesis sections."""
    names = []
    # Intentionally skip known non-thesis H2 sections
    skip_sections = {"투자 선호도"}
    for match in re.finditer(r"^## (.+)$", theses_content, re.MULTILINE):
        name = match.group(1).strip()
        if name not in skip_sections:
            names.append(name)
    return names


def validate_checked_at(
    checked_at_str: str, label: str, reference: date
) -> list[ResearchError]:
    """Validate checked_at is a valid ISO date within staleness threshold."""
    try:
        checked_at = date.fromisoformat(checked_at_str)
    except (ValueError, TypeError):
        return [
            ResearchError(
                rule="checked_at_format",
                detail=f"{label}: '{checked_at_str}' is not a valid ISO date (YYYY-MM-DD)",
            )
        ]

    days_old = (reference - checked_at).days
    if days_old > _STALENESS_THRESHOLD_DAYS:
        return [
            ResearchError(
                rule="checked_at_stale",
                detail=f"{label}: checked_at ({checked_at_str}) is {days_old} days old (threshold: {_STALENESS_THRESHOLD_DAYS})",
            )
        ]
    if days_old < 0:
        return [
            ResearchError(
                rule="checked_at_future",
                detail=f"{label}: checked_at ({checked_at_str}) is in the future",
            )
        ]
    return []


def validate_source(
    source: dict, thesis_name: str, condition_index: int, source_index: int
) -> list[ResearchError]:
    """Validate a single source entry within a condition."""
    prefix = f"{thesis_name}.conditions[{condition_index}].sources[{source_index}]"
    errors = []

    for field in ("title", "url", "tier", "date"):
        if field not in source:
            errors.append(
                ResearchError(
                    rule="source_missing_field",
                    detail=f"{prefix}: missing required field '{field}'",
                )
            )

    if errors:
        return errors

    if not isinstance(source["title"], str) or not source["title"].strip():
        errors.append(
            ResearchError(
                rule="source_title_empty",
                detail=f"{prefix}: title must be a non-empty string",
            )
        )

    if not isinstance(source["url"], str) or not source["url"].strip():
        errors.append(
            ResearchError(
                rule="source_url_empty",
                detail=f"{prefix}: url must be a non-empty string",
            )
        )

    tier = source["tier"]
    if not isinstance(tier, int) or tier < _SOURCE_TIER_RANGE[0] or tier > _SOURCE_TIER_RANGE[1]:
        errors.append(
            ResearchError(
                rule="source_tier_range",
                detail=f"{prefix}: tier must be integer {_SOURCE_TIER_RANGE[0]}-{_SOURCE_TIER_RANGE[1]}, got {tier!r}",
            )
        )

    source_date = source["date"]
    if isinstance(source_date, str):
        try:
            date.fromisoformat(source_date)
        except ValueError:
            errors.append(
                ResearchError(
                    rule="source_date_format",
                    detail=f"{prefix}: date '{source_date}' is not a valid ISO date",
                )
            )
    else:
        errors.append(
            ResearchError(
                rule="source_date_format",
                detail=f"{prefix}: date must be a string, got {type(source_date).__name__}",
            )
        )

    return errors


def validate_condition(
    condition: dict, thesis_name: str, condition_index: int
) -> list[ResearchError]:
    """Validate a single condition entry within a thesis."""
    prefix = f"{thesis_name}.conditions[{condition_index}]"
    errors = []

    for field in ("text", "type", "status", "evidence", "sources"):
        if field not in condition:
            errors.append(
                ResearchError(
                    rule="condition_missing_field",
                    detail=f"{prefix}: missing required field '{field}'",
                )
            )

    if errors:
        return errors

    if not isinstance(condition["text"], str) or not condition["text"].strip():
        errors.append(
            ResearchError(
                rule="condition_text_empty",
                detail=f"{prefix}: text must be a non-empty string",
            )
        )

    cond_type = condition["type"]
    if cond_type not in _VALID_CONDITION_TYPES:
        errors.append(
            ResearchError(
                rule="condition_type_enum",
                detail=f"{prefix}: type must be one of {sorted(_VALID_CONDITION_TYPES)}, got '{cond_type}'",
            )
        )

    cond_status = condition["status"]
    if cond_status not in _VALID_CONDITION_STATUSES:
        errors.append(
            ResearchError(
                rule="condition_status_enum",
                detail=f"{prefix}: status must be one of {sorted(_VALID_CONDITION_STATUSES)}, got '{cond_status}'",
            )
        )

    if not isinstance(condition["evidence"], str) or not condition["evidence"].strip():
        errors.append(
            ResearchError(
                rule="condition_evidence_empty",
                detail=f"{prefix}: evidence must be a non-empty string",
            )
        )

    sources = condition["sources"]
    if not isinstance(sources, list) or len(sources) == 0:
        errors.append(
            ResearchError(
                rule="condition_sources_empty",
                detail=f"{prefix}: must have at least one source",
            )
        )
    elif isinstance(sources, list):
        for si, source in enumerate(sources):
            if not isinstance(source, dict):
                errors.append(
                    ResearchError(
                        rule="source_type",
                        detail=f"{prefix}.sources[{si}]: source must be an object",
                    )
                )
            else:
                errors.extend(validate_source(source, thesis_name, condition_index, si))

    prev_status = condition.get("previous_status")
    if prev_status is not None and prev_status not in _VALID_CONDITION_STATUSES:
        errors.append(
            ResearchError(
                rule="condition_previous_status_enum",
                detail=f"{prefix}: previous_status must be one of {sorted(_VALID_CONDITION_STATUSES)} or null, got '{prev_status}'",
            )
        )

    transition = condition.get("status_transition")
    if transition is not None and transition not in _VALID_TRANSITIONS:
        errors.append(
            ResearchError(
                rule="condition_transition_enum",
                detail=f"{prefix}: status_transition must be one of {sorted(_VALID_TRANSITIONS)} or null, got '{transition}'",
            )
        )

    if prev_status is None and transition is not None and transition != "new":
        errors.append(
            ResearchError(
                rule="condition_transition_consistency",
                detail=f"{prefix}: previous_status is null but status_transition is '{transition}' (expected 'new' or null)",
            )
        )

    if prev_status is not None and transition == "new":
        errors.append(
            ResearchError(
                rule="condition_transition_consistency",
                detail=f"{prefix}: previous_status is '{prev_status}' but status_transition is 'new' (expected stable/improving/degrading)",
            )
        )

    return errors


def validate_thesis_entry(thesis: dict) -> list[ResearchError]:
    """Validate a single thesis entry in thesis-check.json."""
    errors = []

    for field in ("name", "status", "conditions"):
        if field not in thesis:
            errors.append(
                ResearchError(
                    rule="thesis_missing_field",
                    detail=f"thesis entry missing required field '{field}'",
                )
            )

    if errors:
        return errors

    name = thesis["name"]

    if not isinstance(name, str) or not name.strip():
        errors.append(
            ResearchError(
                rule="thesis_name_empty",
                detail="thesis name must be a non-empty string",
            )
        )
        return errors

    thesis_status = thesis["status"]
    if thesis_status not in _VALID_THESIS_STATUSES:
        errors.append(
            ResearchError(
                rule="thesis_status_enum",
                detail=f"{name}: status must be one of {sorted(_VALID_THESIS_STATUSES)}, got '{thesis_status}'",
            )
        )

    prev_status = thesis.get("previous_status")
    if prev_status is not None and prev_status not in _VALID_THESIS_STATUSES:
        errors.append(
            ResearchError(
                rule="thesis_previous_status_enum",
                detail=f"{name}: previous_status must be one of {sorted(_VALID_THESIS_STATUSES)} or null, got '{prev_status}'",
            )
        )

    transition = thesis.get("status_transition")
    if transition is not None and transition not in _VALID_TRANSITIONS:
        errors.append(
            ResearchError(
                rule="thesis_transition_enum",
                detail=f"{name}: status_transition must be one of {sorted(_VALID_TRANSITIONS)} or null, got '{transition}'",
            )
        )

    if prev_status is None and transition is not None and transition != "new":
        errors.append(
            ResearchError(
                rule="thesis_transition_consistency",
                detail=f"{name}: previous_status is null but status_transition is '{transition}' (expected 'new' or null)",
            )
        )

    if prev_status is not None and transition == "new":
        errors.append(
            ResearchError(
                rule="thesis_transition_consistency",
                detail=f"{name}: previous_status is '{prev_status}' but status_transition is 'new' (expected stable/improving/degrading)",
            )
        )

    conditions = thesis["conditions"]
    if not isinstance(conditions, list) or len(conditions) == 0:
        errors.append(
            ResearchError(
                rule="thesis_conditions_empty",
                detail=f"{name}: must have at least one condition",
            )
        )
    elif isinstance(conditions, list):
        for ci, condition in enumerate(conditions):
            if not isinstance(condition, dict):
                errors.append(
                    ResearchError(
                        rule="condition_type",
                        detail=f"{name}.conditions[{ci}]: condition must be an object",
                    )
                )
            else:
                errors.extend(validate_condition(condition, name, ci))

    return errors


def validate_thesis_check(
    thesis_check: dict, thesis_names: list[str], reference: date
) -> list[ResearchError]:
    """Validate thesis-check.json structure and completeness."""
    errors = []

    if "checked_at" not in thesis_check:
        errors.append(
            ResearchError(
                rule="thesis_check_missing_field",
                detail="thesis-check.json: missing required field 'checked_at'",
            )
        )
    else:
        errors.extend(
            validate_checked_at(thesis_check["checked_at"], "thesis-check.json", reference)
        )

    if "theses" not in thesis_check:
        errors.append(
            ResearchError(
                rule="thesis_check_missing_field",
                detail="thesis-check.json: missing required field 'theses'",
            )
        )
        return errors

    theses = thesis_check["theses"]
    if not isinstance(theses, list):
        errors.append(
            ResearchError(
                rule="thesis_check_theses_type",
                detail="thesis-check.json: 'theses' must be an array",
            )
        )
        return errors

    for thesis in theses:
        if not isinstance(thesis, dict):
            errors.append(
                ResearchError(
                    rule="thesis_entry_type",
                    detail="thesis-check.json: each thesis entry must be an object",
                )
            )
        else:
            errors.extend(validate_thesis_entry(thesis))

    checked_names = {
        t["name"] for t in theses if isinstance(t, dict) and "name" in t
    }
    for expected_name in sorted(thesis_names):
        if expected_name not in checked_names:
            errors.append(
                ResearchError(
                    rule="thesis_completeness",
                    detail=f"thesis '{expected_name}' from theses.md has no entry in thesis-check.json",
                )
            )

    return errors


def validate_candidate(
    candidate: dict, candidate_index: int, watchlist_symbols: set[str], thesis_names: list[str]
) -> list[ResearchError]:
    """Validate a single candidate entry in candidates.json."""
    prefix = f"candidates[{candidate_index}]"
    errors = []

    required_fields = (
        "symbol", "name", "market", "sector", "type",
        "related_theses", "rationale", "risks",
        "market_cap_category", "already_in_watchlist",
    )
    for field in required_fields:
        if field not in candidate:
            errors.append(
                ResearchError(
                    rule="candidate_missing_field",
                    detail=f"{prefix}: missing required field '{field}'",
                )
            )

    if errors:
        return errors

    symbol = candidate["symbol"]
    label = f"{prefix} ({symbol})"

    if not isinstance(symbol, str) or not symbol.strip():
        errors.append(
            ResearchError(
                rule="candidate_symbol_empty",
                detail=f"{prefix}: symbol must be a non-empty string",
            )
        )

    if not isinstance(candidate["name"], str) or not candidate["name"].strip():
        errors.append(
            ResearchError(
                rule="candidate_name_empty",
                detail=f"{label}: name must be a non-empty string",
            )
        )

    market = candidate["market"]
    if market not in _VALID_MARKETS:
        errors.append(
            ResearchError(
                rule="candidate_market_enum",
                detail=f"{label}: market must be one of {sorted(_VALID_MARKETS)}, got '{market}'",
            )
        )

    if not isinstance(candidate["sector"], str) or not candidate["sector"].strip():
        errors.append(
            ResearchError(
                rule="candidate_sector_empty",
                detail=f"{label}: sector must be a non-empty string",
            )
        )

    asset_type = candidate["type"]
    if asset_type not in _VALID_ASSET_TYPES:
        errors.append(
            ResearchError(
                rule="candidate_type_enum",
                detail=f"{label}: type must be one of {sorted(_VALID_ASSET_TYPES)}, got '{asset_type}'",
            )
        )

    related = candidate["related_theses"]
    if not isinstance(related, list) or len(related) == 0:
        errors.append(
            ResearchError(
                rule="candidate_related_theses_empty",
                detail=f"{label}: related_theses must be a non-empty array",
            )
        )
    elif isinstance(related, list):
        thesis_name_set = set(thesis_names)
        for thesis_ref in related:
            if not isinstance(thesis_ref, str):
                errors.append(
                    ResearchError(
                        rule="candidate_related_thesis_type",
                        detail=f"{label}: related_theses entries must be strings",
                    )
                )
            elif thesis_ref not in thesis_name_set:
                errors.append(
                    ResearchError(
                        rule="candidate_related_thesis_unknown",
                        detail=f"{label}: related thesis '{thesis_ref}' not found in theses.md",
                    )
                )

    if not isinstance(candidate["rationale"], str) or not candidate["rationale"].strip():
        errors.append(
            ResearchError(
                rule="candidate_rationale_empty",
                detail=f"{label}: rationale must be a non-empty string",
            )
        )

    if not isinstance(candidate["risks"], str) or not candidate["risks"].strip():
        errors.append(
            ResearchError(
                rule="candidate_risks_empty",
                detail=f"{label}: risks must be a non-empty string",
            )
        )

    cap_category = candidate["market_cap_category"]
    if cap_category not in _VALID_MARKET_CAP_CATEGORIES:
        errors.append(
            ResearchError(
                rule="candidate_market_cap_enum",
                detail=f"{label}: market_cap_category must be one of {sorted(_VALID_MARKET_CAP_CATEGORIES)}, got '{cap_category}'",
            )
        )

    if candidate["already_in_watchlist"] is not False:
        errors.append(
            ResearchError(
                rule="candidate_already_in_watchlist_false",
                detail=f"{label}: already_in_watchlist must be false",
            )
        )

    if isinstance(symbol, str) and symbol in watchlist_symbols:
        errors.append(
            ResearchError(
                rule="candidate_watchlist_dedup",
                detail=f"{label}: symbol already exists in watchlist.json",
            )
        )

    return errors


def validate_candidates(
    candidates_data: dict,
    watchlist_symbols: set[str],
    thesis_names: list[str],
    reference: date,
) -> list[ResearchError]:
    """Validate candidates.json structure and cross-references."""
    errors = []

    if "checked_at" not in candidates_data:
        errors.append(
            ResearchError(
                rule="candidates_missing_field",
                detail="candidates.json: missing required field 'checked_at'",
            )
        )
    else:
        errors.extend(
            validate_checked_at(candidates_data["checked_at"], "candidates.json", reference)
        )

    if "candidates" not in candidates_data:
        errors.append(
            ResearchError(
                rule="candidates_missing_field",
                detail="candidates.json: missing required field 'candidates'",
            )
        )
        return errors

    candidates_list = candidates_data["candidates"]
    if not isinstance(candidates_list, list):
        errors.append(
            ResearchError(
                rule="candidates_list_type",
                detail="candidates.json: 'candidates' must be an array",
            )
        )
        return errors

    seen_symbols: set[str] = set()
    for ci, candidate in enumerate(candidates_list):
        if not isinstance(candidate, dict):
            errors.append(
                ResearchError(
                    rule="candidate_entry_type",
                    detail=f"candidates[{ci}]: candidate entry must be an object",
                )
            )
        else:
            errors.extend(
                validate_candidate(candidate, ci, watchlist_symbols, thesis_names)
            )
            symbol = candidate.get("symbol")
            if isinstance(symbol, str) and symbol:
                if symbol in seen_symbols:
                    errors.append(
                        ResearchError(
                            rule="candidate_duplicate_symbol",
                            detail=f"candidates[{ci}] ({symbol}): duplicate symbol in candidates list",
                        )
                    )
                seen_symbols.add(symbol)

    return errors


def validate_research(
    thesis_check: dict,
    candidates_data: dict,
    thesis_names: list[str],
    watchlist_symbols: set[str],
    reference: Optional[date] = None,
) -> ValidationResult:
    """Run all validations on thesis-research output files."""
    if reference is None:
        reference = date.today()

    errors: list[ResearchError] = []

    errors.extend(validate_thesis_check(thesis_check, thesis_names, reference))
    errors.extend(validate_candidates(candidates_data, watchlist_symbols, thesis_names, reference))

    status = "PASS" if not errors else "FAIL"
    return ValidationResult(status=status, errors=errors)


def _serialize_result(result: ValidationResult) -> dict:
    if result.status == "PASS":
        return {"status": "PASS"}
    return {
        "status": "FAIL",
        "errors": [
            {"rule": e.rule, "detail": e.detail}
            for e in result.errors
        ],
    }


def _parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Validate thesis-research output files"
    )
    parser.add_argument(
        "--thesis-check", required=True,
        help="Path to thesis-check.json",
    )
    parser.add_argument(
        "--candidates", required=True,
        help="Path to candidates.json",
    )
    parser.add_argument(
        "--theses", required=True,
        help="Path to config/theses.md",
    )
    parser.add_argument(
        "--watchlist", required=True,
        help="Path to config/watchlist.json",
    )
    parser.add_argument(
        "--reference-date",
        help="Reference date for staleness check (YYYY-MM-DD). Defaults to today.",
    )
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = _parse_args(argv)

    try:
        with open(args.thesis_check) as f:
            thesis_check = json.load(f)
        with open(args.candidates) as f:
            candidates_data = json.load(f)
        with open(args.theses) as f:
            theses_content = f.read()
        with open(args.watchlist) as f:
            watchlist = json.load(f)
    except (json.JSONDecodeError, FileNotFoundError, OSError) as exc:
        print(
            json.dumps({"status": "ERROR", "detail": str(exc)}),
            file=sys.stderr,
        )
        return 1

    thesis_names = parse_thesis_names(theses_content)
    watchlist_symbols = {item["symbol"] for item in watchlist}

    reference = None
    if args.reference_date:
        try:
            reference = date.fromisoformat(args.reference_date)
        except ValueError as exc:
            print(
                json.dumps({"status": "ERROR", "detail": f"Invalid reference date: {exc}"}),
                file=sys.stderr,
            )
            return 1
    else:
        reference = date.today()

    result = validate_research(
        thesis_check, candidates_data, thesis_names, watchlist_symbols, reference
    )
    print(json.dumps(_serialize_result(result), ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())

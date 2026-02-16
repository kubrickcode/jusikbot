"""Unit tests for validate_report module."""

import unittest
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from validate_report import (
    extract_sections,
    validate_required_sections,
    validate_thesis_counterarguments,
    validate_high_confidence_evidence,
    validate_forbidden_expressions,
    validate_source_tags,
    validate_kr_data_warning,
    validate_action_directives,
    validate_report,
)

VALID_REPORT = """\
# 포트폴리오 분석 리포트

## 요약

포트폴리오 현재 가치는 5,200,000원입니다 [summary]. AI 반도체 논제는 지속 상승 모멘텀을 보이고 있으나, 빅테크 실적 둔화 리스크가 존재합니다.

## 논제별 현황

### 논제 1: AI 반도체 인프라 확장

**신뢰도**: High

**찬성 근거**:
- NVDA 20일 이평선 상향 돌파 [추세]
- 거래량 전일 대비 150% 증가 [모멘텀] [psql]
- 업계 전문가 긍정 전망 발표 [외부]

**반론/리스크**:
공급 과잉 우려가 대두되고 있으며, 중국 수출 규제 강화 가능성이 있습니다.

### 논제 2: 한국 시장 저평가

**신뢰도**: Medium

**찬성 근거**:
- 069500(KODEX 200) 52주 대비 저점 근접 [추세] [summary]

**반론/리스크**:
환율 변동성 증가로 외국인 자금 유출 우려. 수정주가(adj_close) 미반영으로 기술 지표 과대평가 가능성 있음.

## 배분 제안

| 종목 | 배분 | 역할 |
|------|------|------|
| QQQ | 2,000,000원 | core |
| NVDA | 1,300,000원 | satellite |

QQQ 비중을 40%로 유지합니다 [user].

## 리스크 요인

- 금리 인상 지속 시 밸류에이션 부담
- 지정학적 리스크 확대 가능성
"""


class TestExtractSections(unittest.TestCase):
    """extract_sections"""

    def test_extracts_h2_and_h3_sections(self):
        sections = extract_sections(VALID_REPORT)
        self.assertIn("요약", sections)
        self.assertIn("논제별 현황", sections)
        self.assertIn("배분 제안", sections)
        self.assertIn("리스크 요인", sections)

    def test_returns_empty_for_blank_content(self):
        sections = extract_sections("")
        self.assertEqual(sections, {})


class TestRequiredSections(unittest.TestCase):
    """validate_required_sections"""

    def test_passes_with_all_sections(self):
        sections = extract_sections(VALID_REPORT)
        errors = validate_required_sections(sections)
        self.assertEqual(errors, [])

    def test_rejects_missing_section(self):
        report = """\
## 요약

분석 결과 [summary]

## 논제별 현황

현황 내용

## 배분 제안

제안 내용
"""
        sections = extract_sections(report)
        errors = validate_required_sections(sections)
        self.assertEqual(len(errors), 1)
        self.assertIn("리스크 요인", errors[0].detail)


class TestThesisCounterarguments(unittest.TestCase):
    """validate_thesis_counterarguments"""

    def test_passes_with_counterarguments(self):
        sections = extract_sections(VALID_REPORT)
        errors = validate_thesis_counterarguments(sections)
        self.assertEqual(errors, [])

    def test_rejects_empty_counterargument(self):
        report = """\
## 논제별 현황

### 논제 1: AI 반도체

**찬성 근거**:
- 강세 지속 [추세]

**반론/리스크**:

### 논제 2: 한국 시장

**찬성 근거**:
- 저평가 [추세]

**반론/리스크**:
환율 변동성 증가
"""
        sections = extract_sections(report)
        errors = validate_thesis_counterarguments(sections)
        self.assertEqual(len(errors), 1)
        self.assertIn("논제 1", errors[0].detail)


class TestHighConfidenceEvidence(unittest.TestCase):
    """validate_high_confidence_evidence"""

    def test_passes_with_diverse_evidence(self):
        errors = validate_high_confidence_evidence(VALID_REPORT)
        self.assertEqual(errors, [])

    def test_rejects_single_evidence_category(self):
        report = """\
## 논제별 현황

### 논제 1: AI 반도체

**신뢰도**: High

**찬성 근거**:
- NVDA 20일 이평선 돌파 [추세]
- 상승 채널 유지 [추세]
- 저항선 상향 돌파 [추세]

**반론/리스크**:
경쟁 심화 우려
"""
        errors = validate_high_confidence_evidence(report)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "high_confidence_evidence")

    def test_passes_without_high_confidence(self):
        report = """\
## 논제별 현황

### 논제 1: AI 반도체

**신뢰도**: Medium

**찬성 근거**:
- 강세 지속 [추세]

**반론/리스크**:
리스크 존재
"""
        errors = validate_high_confidence_evidence(report)
        self.assertEqual(errors, [])


class TestForbiddenExpressions(unittest.TestCase):
    """validate_forbidden_expressions"""

    def test_passes_clean_report(self):
        errors = validate_forbidden_expressions(VALID_REPORT)
        self.assertEqual(errors, [])

    def test_rejects_forbidden_words(self):
        report = "NVDA는 확실히 상승할 것이며, 반드시 매수해야 합니다."
        errors = validate_forbidden_expressions(report)
        self.assertEqual(len(errors), 2)
        rules = {e.detail for e in errors}
        self.assertTrue(any("확실히" in d for d in rules))
        self.assertTrue(any("반드시" in d for d in rules))

    def test_rejects_100_percent(self):
        report = "상승 확률 100% 확신합니다."
        errors = validate_forbidden_expressions(report)
        self.assertEqual(len(errors), 1)
        self.assertIn("100%", errors[0].detail)


class TestSourceTags(unittest.TestCase):
    """validate_source_tags"""

    def test_passes_with_tagged_numbers(self):
        errors = validate_source_tags(VALID_REPORT)
        self.assertEqual(errors, [])

    def test_rejects_untagged_number(self):
        report = """\
## 요약

포트폴리오 가치는 5,200,000원이며, 전월 대비 3.5% 상승했습니다.
"""
        errors = validate_source_tags(report)
        self.assertGreaterEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "source_tag")

    def test_ignores_table_rows(self):
        report = """\
| 종목 | 가격 |
|------|------|
| NVDA | 1,300,000원 |
"""
        errors = validate_source_tags(report)
        self.assertEqual(errors, [])

    def test_ignores_headings(self):
        report = "## 논제 1: AI 반도체 50% 성장"
        errors = validate_source_tags(report)
        self.assertEqual(errors, [])

    def test_passes_config_tag(self):
        report = "총 예산 5,000,000원 중 70%를 코어에 배분합니다 [config]."
        errors = validate_source_tags(report)
        self.assertEqual(errors, [])


class TestKrDataWarning(unittest.TestCase):
    """validate_kr_data_warning"""

    def test_passes_with_warning(self):
        errors = validate_kr_data_warning(VALID_REPORT)
        self.assertEqual(errors, [])

    def test_rejects_without_warning(self):
        report = """\
## 논제별 현황

069500은 양호한 흐름을 보이고 있습니다 [psql].
"""
        errors = validate_kr_data_warning(report)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "kr_data_warning")

    def test_passes_without_kr_stock(self):
        report = """\
## 논제별 현황

NVDA는 상승 추세입니다 [summary].
"""
        errors = validate_kr_data_warning(report)
        self.assertEqual(errors, [])


class TestActionDirectives(unittest.TestCase):
    """validate_action_directives"""

    def test_skips_when_no_holdings(self):
        sections = extract_sections(VALID_REPORT)
        errors = validate_action_directives(sections, has_holdings=False)
        self.assertEqual(errors, [])

    def test_rejects_missing_section_when_holdings_exist(self):
        sections = extract_sections(VALID_REPORT)
        errors = validate_action_directives(sections, has_holdings=True)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "action_directives_required")
        self.assertIn("행동 지시", errors[0].detail)

    def test_passes_with_section_when_holdings_exist(self):
        report_with_action = VALID_REPORT.rstrip() + """

## 행동 지시

| 종목 | 행동 | 금액 | 비고 |
|------|------|------|------|
| NVDA | 매수 | 200,000원 | 토스증권 |
| QQQ | 유지 | 0원 | — |
"""
        sections = extract_sections(report_with_action)
        errors = validate_action_directives(sections, has_holdings=True)
        self.assertEqual(errors, [])


class TestSourceTagsWithHoldings(unittest.TestCase):
    """validate_source_tags — [holdings] tag support"""

    def test_passes_holdings_tag(self):
        report = "현재 NVDA 보유 평가금액은 900,000원입니다 [holdings]."
        errors = validate_source_tags(report)
        self.assertEqual(errors, [])


class TestSourceTagsWithResearch(unittest.TestCase):
    """validate_source_tags — [research] tag support"""

    def test_passes_research_tag(self):
        report = "NVDA 분기 매출 YoY +78% 성장 지속 [research]."
        errors = validate_source_tags(report)
        self.assertEqual(errors, [])


class TestValidateReport(unittest.TestCase):
    """validate_report — integration of all rules"""

    def test_passes_valid_report(self):
        result = validate_report(VALID_REPORT)
        self.assertEqual(result.status, "PASS")
        self.assertEqual(result.errors, [])

    def test_collects_multiple_errors(self):
        report = """\
## 요약

NVDA는 확실히 5,200,000원까지 상승합니다.

## 논제별 현황

### 논제 1: AI 반도체

**찬성 근거**:
- 강세 [추세]

**반론/리스크**:
리스크 존재

## 배분 제안

QQQ 비중을 유지합니다 [user].
"""
        result = validate_report(report)
        self.assertEqual(result.status, "FAIL")
        rules = {e.rule for e in result.errors}
        self.assertIn("required_sections", rules)
        self.assertIn("forbidden_expression", rules)

    def test_fails_when_holdings_but_no_action_section(self):
        result = validate_report(VALID_REPORT, has_holdings=True)
        self.assertEqual(result.status, "FAIL")
        rules = {e.rule for e in result.errors}
        self.assertIn("action_directives_required", rules)

    def test_passes_without_holdings_flag(self):
        """Backward compatibility: has_holdings=False by default."""
        result = validate_report(VALID_REPORT, has_holdings=False)
        self.assertEqual(result.status, "PASS")


if __name__ == "__main__":
    unittest.main()

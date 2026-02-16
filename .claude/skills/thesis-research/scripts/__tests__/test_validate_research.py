"""Unit tests for validate_research module."""

import unittest
import sys
from datetime import date
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from validate_research import (
    ResearchError,
    parse_thesis_names,
    validate_checked_at,
    validate_source,
    validate_condition,
    validate_thesis_entry,
    validate_thesis_check,
    validate_candidate,
    validate_candidates,
    validate_research,
)

REFERENCE_DATE = date(2026, 2, 16)

THESES_MD = """# 투자 논제

## 투자 선호도

- 안전 투자(코어)는 나스닥 ETF에 집중한다

## AI 인프라 확장

**역할**: satellite
**유효 조건**:
- NVDA 분기 매출 전년 대비 30%+ 성장 지속

**무효화 조건**:
- GPU 공급 과잉 신호

**관련 종목**: NVDA, ASML

## 나스닥 장기 성장

**역할**: core
**유효 조건**:
- QQQ 52주 추세 우상향

**무효화 조건**:
- 금리 5%+ 장기 지속

**관련 종목**: QQQ
"""

THESIS_NAMES = ["AI 인프라 확장", "나스닥 장기 성장"]

WATCHLIST_SYMBOLS = {"NVDA", "ASML", "META", "MSFT", "AAPL", "133690"}


def _make_source(**overrides):
    base = {
        "title": "NVIDIA Q4 FY2026 Earnings",
        "url": "https://investor.nvidia.com/earnings",
        "tier": 1,
        "date": "2026-02-10",
    }
    base.update(overrides)
    return base


def _make_condition(**overrides):
    base = {
        "text": "NVDA 분기 매출 전년 대비 30%+ 성장 지속",
        "type": "validity",
        "status": "met",
        "evidence": "NVDA Q4 매출 $44B, YoY +78%",
        "sources": [_make_source()],
    }
    base.update(overrides)
    return base


def _make_thesis(**overrides):
    base = {
        "name": "AI 인프라 확장",
        "status": "valid",
        "conditions": [
            _make_condition(),
            _make_condition(
                text="GPU 공급 과잉 신호",
                type="invalidation",
                status="not_yet",
                evidence="GPU 리드타임 여전히 6개월+",
            ),
        ],
    }
    base.update(overrides)
    return base


def _make_thesis_check(**overrides):
    base = {
        "checked_at": "2026-02-15",
        "theses": [
            _make_thesis(),
            _make_thesis(
                name="나스닥 장기 성장",
                conditions=[
                    _make_condition(
                        text="QQQ 52주 추세 우상향",
                        status="met",
                        evidence="QQQ 52주 최고가 경신",
                    ),
                    _make_condition(
                        text="금리 5%+ 장기 지속",
                        type="invalidation",
                        status="not_yet",
                        evidence="Fed 금리 4.25%",
                    ),
                ],
            ),
        ],
    }
    base.update(overrides)
    return base


def _make_candidate(**overrides):
    base = {
        "symbol": "MRVL",
        "name": "Marvell Technology",
        "market": "US",
        "sector": "semiconductor",
        "type": "stock",
        "related_theses": ["AI 인프라 확장"],
        "rationale": "Data center custom ASIC growth",
        "risks": "Cyclical semiconductor demand",
        "market_cap_category": "large",
        "already_in_watchlist": False,
    }
    base.update(overrides)
    return base


def _make_candidates(**overrides):
    base = {
        "checked_at": "2026-02-15",
        "candidates": [_make_candidate()],
    }
    base.update(overrides)
    return base


class TestParseThesisNames(unittest.TestCase):
    """parse_thesis_names"""

    def test_extracts_thesis_headers(self):
        names = parse_thesis_names(THESES_MD)
        self.assertEqual(names, ["AI 인프라 확장", "나스닥 장기 성장"])

    def test_skips_preference_section(self):
        names = parse_thesis_names(THESES_MD)
        self.assertNotIn("투자 선호도", names)

    def test_returns_empty_for_no_h2(self):
        names = parse_thesis_names("# Title\nSome text")
        self.assertEqual(names, [])


class TestValidateCheckedAt(unittest.TestCase):
    """validate_checked_at"""

    def test_passes_recent_date(self):
        errors = validate_checked_at("2026-02-15", "test", REFERENCE_DATE)
        self.assertEqual(errors, [])

    def test_passes_same_day(self):
        errors = validate_checked_at("2026-02-16", "test", REFERENCE_DATE)
        self.assertEqual(errors, [])

    def test_passes_at_threshold_boundary(self):
        errors = validate_checked_at("2026-02-09", "test", REFERENCE_DATE)
        self.assertEqual(errors, [])

    def test_rejects_stale_date(self):
        errors = validate_checked_at("2026-02-01", "test", REFERENCE_DATE)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "checked_at_stale")

    def test_rejects_invalid_format(self):
        errors = validate_checked_at("02/15/2026", "test", REFERENCE_DATE)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "checked_at_format")

    def test_rejects_future_date(self):
        errors = validate_checked_at("2026-02-20", "test", REFERENCE_DATE)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "checked_at_future")


class TestValidateSource(unittest.TestCase):
    """validate_source"""

    def test_passes_valid_source(self):
        errors = validate_source(_make_source(), "thesis", 0, 0)
        self.assertEqual(errors, [])

    def test_rejects_missing_title(self):
        source = _make_source()
        del source["title"]
        errors = validate_source(source, "thesis", 0, 0)
        rules = {e.rule for e in errors}
        self.assertIn("source_missing_field", rules)

    def test_rejects_tier_zero(self):
        errors = validate_source(_make_source(tier=0), "thesis", 0, 0)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "source_tier_range")

    def test_rejects_tier_five(self):
        errors = validate_source(_make_source(tier=5), "thesis", 0, 0)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "source_tier_range")

    def test_passes_tier_boundaries(self):
        for tier in (1, 2, 3, 4):
            errors = validate_source(_make_source(tier=tier), "thesis", 0, 0)
            self.assertEqual(errors, [], f"tier={tier} should pass")

    def test_rejects_float_tier(self):
        errors = validate_source(_make_source(tier=2.5), "thesis", 0, 0)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "source_tier_range")

    def test_rejects_invalid_date(self):
        errors = validate_source(_make_source(date="not-a-date"), "thesis", 0, 0)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "source_date_format")


class TestValidateCondition(unittest.TestCase):
    """validate_condition"""

    def test_passes_valid_condition(self):
        errors = validate_condition(_make_condition(), "thesis", 0)
        self.assertEqual(errors, [])

    def test_rejects_invalid_type_enum(self):
        errors = validate_condition(_make_condition(type="prerequisite"), "thesis", 0)
        rules = {e.rule for e in errors}
        self.assertIn("condition_type_enum", rules)

    def test_rejects_invalid_status_enum(self):
        errors = validate_condition(_make_condition(status="maybe"), "thesis", 0)
        rules = {e.rule for e in errors}
        self.assertIn("condition_status_enum", rules)

    def test_rejects_empty_sources(self):
        errors = validate_condition(_make_condition(sources=[]), "thesis", 0)
        rules = {e.rule for e in errors}
        self.assertIn("condition_sources_empty", rules)

    def test_rejects_empty_evidence(self):
        errors = validate_condition(_make_condition(evidence=""), "thesis", 0)
        rules = {e.rule for e in errors}
        self.assertIn("condition_evidence_empty", rules)

    def test_accepts_all_valid_statuses(self):
        for status in ("met", "partially_met", "not_yet", "refuted", "unknown"):
            errors = validate_condition(_make_condition(status=status), "thesis", 0)
            self.assertEqual(errors, [], f"status={status} should pass")

    def test_accepts_optional_quantitative_distance(self):
        cond = _make_condition(quantitative_distance="+18pp margin")
        errors = validate_condition(cond, "thesis", 0)
        self.assertEqual(errors, [])

    def test_accepts_null_previous_status(self):
        cond = _make_condition(previous_status=None, status_transition="new")
        errors = validate_condition(cond, "thesis", 0)
        self.assertEqual(errors, [])

    def test_accepts_valid_transition_with_previous(self):
        cond = _make_condition(previous_status="not_yet", status_transition="improving")
        errors = validate_condition(cond, "thesis", 0)
        self.assertEqual(errors, [])

    def test_accepts_stable_transition(self):
        cond = _make_condition(previous_status="met", status_transition="stable")
        errors = validate_condition(cond, "thesis", 0)
        self.assertEqual(errors, [])

    def test_accepts_degrading_transition(self):
        cond = _make_condition(previous_status="met", status_transition="degrading")
        errors = validate_condition(cond, "thesis", 0)
        self.assertEqual(errors, [])

    def test_rejects_invalid_previous_status_enum(self):
        cond = _make_condition(previous_status="strong")
        errors = validate_condition(cond, "thesis", 0)
        rules = {e.rule for e in errors}
        self.assertIn("condition_previous_status_enum", rules)

    def test_rejects_invalid_transition_enum(self):
        cond = _make_condition(status_transition="worsening")
        errors = validate_condition(cond, "thesis", 0)
        rules = {e.rule for e in errors}
        self.assertIn("condition_transition_enum", rules)

    def test_rejects_null_previous_with_non_new_transition(self):
        """previous_status=null but transition='stable' is inconsistent."""
        cond = _make_condition(previous_status=None, status_transition="stable")
        errors = validate_condition(cond, "thesis", 0)
        rules = {e.rule for e in errors}
        self.assertIn("condition_transition_consistency", rules)

    def test_rejects_previous_with_new_transition(self):
        """previous_status='met' but transition='new' is inconsistent."""
        cond = _make_condition(previous_status="met", status_transition="new")
        errors = validate_condition(cond, "thesis", 0)
        rules = {e.rule for e in errors}
        self.assertIn("condition_transition_consistency", rules)

    def test_accepts_no_transition_fields(self):
        """Backward compat: omitting both fields is valid."""
        cond = _make_condition()
        errors = validate_condition(cond, "thesis", 0)
        self.assertEqual(errors, [])


class TestValidateThesisEntry(unittest.TestCase):
    """validate_thesis_entry"""

    def test_passes_valid_thesis(self):
        errors = validate_thesis_entry(_make_thesis())
        self.assertEqual(errors, [])

    def test_rejects_invalid_status(self):
        errors = validate_thesis_entry(_make_thesis(status="uncertain"))
        rules = {e.rule for e in errors}
        self.assertIn("thesis_status_enum", rules)

    def test_rejects_empty_conditions(self):
        errors = validate_thesis_entry(_make_thesis(conditions=[]))
        rules = {e.rule for e in errors}
        self.assertIn("thesis_conditions_empty", rules)

    def test_rejects_missing_name(self):
        thesis = _make_thesis()
        del thesis["name"]
        errors = validate_thesis_entry(thesis)
        rules = {e.rule for e in errors}
        self.assertIn("thesis_missing_field", rules)

    def test_accepts_optional_upstream_dependency(self):
        errors = validate_thesis_entry(
            _make_thesis(upstream_dependency="나스닥 장기 성장")
        )
        self.assertEqual(errors, [])

    def test_accepts_optional_chain_impact(self):
        errors = validate_thesis_entry(
            _make_thesis(chain_impact="AI 인프라 논제 약화 시 META 위성 논제도 영향")
        )
        self.assertEqual(errors, [])

    def test_accepts_thesis_transition_fields(self):
        errors = validate_thesis_entry(
            _make_thesis(previous_status="valid", status_transition="stable")
        )
        self.assertEqual(errors, [])

    def test_accepts_thesis_new_transition(self):
        errors = validate_thesis_entry(
            _make_thesis(previous_status=None, status_transition="new")
        )
        self.assertEqual(errors, [])

    def test_rejects_invalid_thesis_previous_status(self):
        errors = validate_thesis_entry(
            _make_thesis(previous_status="strong")
        )
        rules = {e.rule for e in errors}
        self.assertIn("thesis_previous_status_enum", rules)

    def test_rejects_invalid_thesis_transition(self):
        errors = validate_thesis_entry(
            _make_thesis(status_transition="worsening")
        )
        rules = {e.rule for e in errors}
        self.assertIn("thesis_transition_enum", rules)

    def test_rejects_thesis_null_previous_with_stable(self):
        errors = validate_thesis_entry(
            _make_thesis(previous_status=None, status_transition="stable")
        )
        rules = {e.rule for e in errors}
        self.assertIn("thesis_transition_consistency", rules)

    def test_rejects_thesis_previous_with_new(self):
        errors = validate_thesis_entry(
            _make_thesis(previous_status="valid", status_transition="new")
        )
        rules = {e.rule for e in errors}
        self.assertIn("thesis_transition_consistency", rules)


class TestValidateThesisCheck(unittest.TestCase):
    """validate_thesis_check — completeness against theses.md"""

    def test_passes_complete_thesis_check(self):
        errors = validate_thesis_check(
            _make_thesis_check(), THESIS_NAMES, REFERENCE_DATE
        )
        self.assertEqual(errors, [])

    def test_rejects_missing_thesis(self):
        """Every thesis in theses.md must have an entry."""
        tc = _make_thesis_check()
        tc["theses"] = [_make_thesis()]  # Only AI infra, missing nasdaq
        errors = validate_thesis_check(tc, THESIS_NAMES, REFERENCE_DATE)
        rules = {e.rule for e in errors}
        self.assertIn("thesis_completeness", rules)
        detail = next(e.detail for e in errors if e.rule == "thesis_completeness")
        self.assertIn("나스닥 장기 성장", detail)

    def test_rejects_missing_checked_at(self):
        tc = _make_thesis_check()
        del tc["checked_at"]
        errors = validate_thesis_check(tc, THESIS_NAMES, REFERENCE_DATE)
        rules = {e.rule for e in errors}
        self.assertIn("thesis_check_missing_field", rules)

    def test_rejects_missing_theses_array(self):
        tc = {"checked_at": "2026-02-15"}
        errors = validate_thesis_check(tc, THESIS_NAMES, REFERENCE_DATE)
        rules = {e.rule for e in errors}
        self.assertIn("thesis_check_missing_field", rules)


class TestValidateCandidate(unittest.TestCase):
    """validate_candidate"""

    def test_passes_valid_candidate(self):
        errors = validate_candidate(_make_candidate(), 0, WATCHLIST_SYMBOLS, THESIS_NAMES)
        self.assertEqual(errors, [])

    def test_rejects_invalid_market(self):
        errors = validate_candidate(
            _make_candidate(market="JP"), 0, WATCHLIST_SYMBOLS, THESIS_NAMES
        )
        rules = {e.rule for e in errors}
        self.assertIn("candidate_market_enum", rules)

    def test_rejects_invalid_type(self):
        errors = validate_candidate(
            _make_candidate(type="bond"), 0, WATCHLIST_SYMBOLS, THESIS_NAMES
        )
        rules = {e.rule for e in errors}
        self.assertIn("candidate_type_enum", rules)

    def test_rejects_invalid_market_cap(self):
        errors = validate_candidate(
            _make_candidate(market_cap_category="mega"), 0, WATCHLIST_SYMBOLS, THESIS_NAMES
        )
        rules = {e.rule for e in errors}
        self.assertIn("candidate_market_cap_enum", rules)

    def test_rejects_already_in_watchlist_true(self):
        errors = validate_candidate(
            _make_candidate(already_in_watchlist=True), 0, WATCHLIST_SYMBOLS, THESIS_NAMES
        )
        rules = {e.rule for e in errors}
        self.assertIn("candidate_already_in_watchlist_false", rules)

    def test_rejects_symbol_in_watchlist(self):
        errors = validate_candidate(
            _make_candidate(symbol="NVDA"), 0, WATCHLIST_SYMBOLS, THESIS_NAMES
        )
        rules = {e.rule for e in errors}
        self.assertIn("candidate_watchlist_dedup", rules)

    def test_rejects_unknown_related_thesis(self):
        errors = validate_candidate(
            _make_candidate(related_theses=["존재하지 않는 논제"]),
            0, WATCHLIST_SYMBOLS, THESIS_NAMES,
        )
        rules = {e.rule for e in errors}
        self.assertIn("candidate_related_thesis_unknown", rules)

    def test_rejects_empty_related_theses(self):
        errors = validate_candidate(
            _make_candidate(related_theses=[]), 0, WATCHLIST_SYMBOLS, THESIS_NAMES
        )
        rules = {e.rule for e in errors}
        self.assertIn("candidate_related_theses_empty", rules)

    def test_rejects_empty_rationale(self):
        errors = validate_candidate(
            _make_candidate(rationale=""), 0, WATCHLIST_SYMBOLS, THESIS_NAMES
        )
        rules = {e.rule for e in errors}
        self.assertIn("candidate_rationale_empty", rules)

    def test_rejects_missing_field(self):
        candidate = _make_candidate()
        del candidate["symbol"]
        errors = validate_candidate(candidate, 0, WATCHLIST_SYMBOLS, THESIS_NAMES)
        rules = {e.rule for e in errors}
        self.assertIn("candidate_missing_field", rules)


class TestValidateCandidates(unittest.TestCase):
    """validate_candidates — dedup and cross-reference"""

    def test_passes_valid_candidates(self):
        errors = validate_candidates(
            _make_candidates(), WATCHLIST_SYMBOLS, THESIS_NAMES, REFERENCE_DATE
        )
        self.assertEqual(errors, [])

    def test_passes_empty_candidates_list(self):
        """No candidates is valid — research may find nothing new."""
        errors = validate_candidates(
            {"checked_at": "2026-02-15", "candidates": []},
            WATCHLIST_SYMBOLS, THESIS_NAMES, REFERENCE_DATE,
        )
        self.assertEqual(errors, [])

    def test_rejects_duplicate_symbols_in_candidates(self):
        data = _make_candidates()
        data["candidates"] = [_make_candidate(), _make_candidate()]
        errors = validate_candidates(data, WATCHLIST_SYMBOLS, THESIS_NAMES, REFERENCE_DATE)
        rules = {e.rule for e in errors}
        self.assertIn("candidate_duplicate_symbol", rules)

    def test_rejects_stale_checked_at(self):
        data = _make_candidates(checked_at="2026-01-01")
        errors = validate_candidates(data, WATCHLIST_SYMBOLS, THESIS_NAMES, REFERENCE_DATE)
        rules = {e.rule for e in errors}
        self.assertIn("checked_at_stale", rules)

    def test_rejects_missing_candidates_field(self):
        errors = validate_candidates(
            {"checked_at": "2026-02-15"},
            WATCHLIST_SYMBOLS, THESIS_NAMES, REFERENCE_DATE,
        )
        rules = {e.rule for e in errors}
        self.assertIn("candidates_missing_field", rules)


class TestValidateResearch(unittest.TestCase):
    """validate_research — integration of all rules"""

    def test_passes_valid_research(self):
        result = validate_research(
            _make_thesis_check(),
            _make_candidates(),
            THESIS_NAMES,
            WATCHLIST_SYMBOLS,
            REFERENCE_DATE,
        )
        self.assertEqual(result.status, "PASS")
        self.assertEqual(result.errors, [])

    def test_fails_on_thesis_check_errors(self):
        tc = _make_thesis_check()
        tc["theses"] = []
        result = validate_research(
            tc, _make_candidates(), THESIS_NAMES, WATCHLIST_SYMBOLS, REFERENCE_DATE,
        )
        self.assertEqual(result.status, "FAIL")
        self.assertGreater(len(result.errors), 0)

    def test_fails_on_candidate_errors(self):
        cands = _make_candidates()
        cands["candidates"] = [_make_candidate(symbol="NVDA")]
        result = validate_research(
            _make_thesis_check(), cands, THESIS_NAMES, WATCHLIST_SYMBOLS, REFERENCE_DATE,
        )
        self.assertEqual(result.status, "FAIL")
        rules = {e.rule for e in result.errors}
        self.assertIn("candidate_watchlist_dedup", rules)

    def test_aggregates_errors_from_both_files(self):
        tc = _make_thesis_check()
        tc["theses"] = [_make_thesis()]  # Missing nasdaq thesis
        cands = _make_candidates()
        cands["candidates"] = [_make_candidate(market="JP")]
        result = validate_research(
            tc, cands, THESIS_NAMES, WATCHLIST_SYMBOLS, REFERENCE_DATE,
        )
        self.assertEqual(result.status, "FAIL")
        rules = {e.rule for e in result.errors}
        self.assertIn("thesis_completeness", rules)
        self.assertIn("candidate_market_enum", rules)


if __name__ == "__main__":
    unittest.main()

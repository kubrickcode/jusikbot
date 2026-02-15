"""Unit tests for validate_allocation module."""

import unittest
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from validate_allocation import (
    AllocationError,
    validate_budget_total,
    validate_adjustment_unit,
    validate_position_limits,
    validate_sector_concentration,
    validate_min_position,
    validate_core_satellite_ratio,
    validate_core_internal_ratio,
    validate_anchoring,
    validate_confidence_pool,
    validate_watchlist_membership,
    validate_allocation,
)

SETTINGS = {
    "budget_krw": 5_000_000,
    "adjustment_unit_krw": 100_000,
    "strategy": {
        "core_pct": 70,
        "satellite_pct": 30,
        "core_internal_ratio": {"QQQ": 2, "069500": 1},
    },
    "sizing": {
        "high_confidence_pool_pct": 50,
        "medium_confidence_pool_pct": 35,
        "low_confidence_pool_pct": 15,
    },
    "risk_tolerance": {
        "max_single_stock_pct": 30,
        "max_single_etf_pct": 50,
        "max_sector_concentration_pct": 60,
        "max_drawdown_warning_pct": 15,
        "max_drawdown_action_pct": 25,
        "min_position_size_krw": 100_000,
    },
    "anchoring": {
        "monthly_max_change_per_position_krw": 200_000,
        "monthly_max_total_change_krw": 500_000,
        "quarterly_max_change_per_position_krw": 400_000,
        "quarterly_max_total_change_krw": 1_500_000,
    },
}

WATCHLIST = [
    {"symbol": "ASML", "type": "stock", "sector": "semiconductor", "market": "US"},
    {"symbol": "META", "type": "stock", "sector": "big-tech", "market": "US"},
    {"symbol": "NVDA", "type": "stock", "sector": "semiconductor", "market": "US"},
    {"symbol": "QQQ", "type": "etf", "sector": "us-broad-market", "market": "US"},
    {"symbol": "069500", "type": "etf", "sector": "kr-broad-market", "market": "KR"},
]

VALID_ALLOCATION = {
    "QQQ": {"amount": 2_300_000, "role": "core", "confidence": "high"},
    "069500": {"amount": 1_200_000, "role": "core", "confidence": "medium"},
    "NVDA": {"amount": 600_000, "role": "satellite", "confidence": "medium"},
    "ASML": {"amount": 500_000, "role": "satellite", "confidence": "medium"},
    "META": {"amount": 400_000, "role": "satellite", "confidence": "low"},
}


class TestBudgetTotal(unittest.TestCase):
    """validate_budget_total"""

    def test_passes_when_total_matches_budget(self):
        errors = validate_budget_total(VALID_ALLOCATION, SETTINGS)
        self.assertEqual(errors, [])

    def test_rejects_total_exceeding_budget(self):
        allocation = {
            "QQQ": {"amount": 2_500_000, "role": "core", "confidence": "high"},
            "NVDA": {"amount": 1_500_000, "role": "satellite", "confidence": "high"},
            "META": {"amount": 1_100_000, "role": "satellite", "confidence": "medium"},
        }
        errors = validate_budget_total(allocation, SETTINGS)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "budget_total")
        self.assertEqual(errors[0].actual, 5_100_000)

    def test_rejects_total_under_budget(self):
        allocation = {
            "QQQ": {"amount": 2_500_000, "role": "core", "confidence": "high"},
            "NVDA": {"amount": 1_500_000, "role": "satellite", "confidence": "high"},
            "META": {"amount": 900_000, "role": "satellite", "confidence": "medium"},
        }
        errors = validate_budget_total(allocation, SETTINGS)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "budget_total")
        self.assertEqual(errors[0].actual, 4_900_000)


class TestAdjustmentUnit(unittest.TestCase):
    """validate_adjustment_unit"""

    def test_passes_when_all_multiples(self):
        errors = validate_adjustment_unit(VALID_ALLOCATION, SETTINGS)
        self.assertEqual(errors, [])

    def test_rejects_non_multiple(self):
        allocation = {
            "QQQ": {"amount": 2_000_000, "role": "core", "confidence": "high"},
            "NVDA": {"amount": 1_350_000, "role": "satellite", "confidence": "high"},
            "META": {"amount": 1_700_000, "role": "satellite", "confidence": "medium"},
        }
        errors = validate_adjustment_unit(allocation, SETTINGS)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "adjustment_unit")
        self.assertIn("NVDA", errors[0].detail)


class TestPositionLimits(unittest.TestCase):
    """validate_position_limits"""

    def test_passes_stock_at_30_percent_boundary(self):
        allocation = {
            "QQQ": {"amount": 2_000_000, "role": "core", "confidence": "high"},
            "NVDA": {"amount": 1_500_000, "role": "satellite", "confidence": "high"},
            "069500": {"amount": 1_000_000, "role": "core", "confidence": "medium"},
            "META": {"amount": 500_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_position_limits(allocation, WATCHLIST, SETTINGS)
        self.assertEqual(errors, [])

    def test_rejects_stock_exceeding_30_percent(self):
        allocation = {
            "QQQ": {"amount": 2_000_000, "role": "core", "confidence": "high"},
            "NVDA": {"amount": 1_600_000, "role": "satellite", "confidence": "high"},
            "069500": {"amount": 1_000_000, "role": "core", "confidence": "medium"},
            "META": {"amount": 400_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_position_limits(allocation, WATCHLIST, SETTINGS)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "single_stock_limit")
        self.assertIn("NVDA", errors[0].detail)

    def test_passes_etf_at_50_percent_boundary(self):
        allocation = {
            "QQQ": {"amount": 2_500_000, "role": "core", "confidence": "high"},
            "NVDA": {"amount": 1_000_000, "role": "satellite", "confidence": "high"},
            "069500": {"amount": 1_000_000, "role": "core", "confidence": "medium"},
            "META": {"amount": 500_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_position_limits(allocation, WATCHLIST, SETTINGS)
        self.assertEqual(errors, [])

    def test_rejects_etf_exceeding_50_percent(self):
        allocation = {
            "QQQ": {"amount": 2_600_000, "role": "core", "confidence": "high"},
            "NVDA": {"amount": 1_000_000, "role": "satellite", "confidence": "high"},
            "069500": {"amount": 900_000, "role": "core", "confidence": "medium"},
            "META": {"amount": 500_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_position_limits(allocation, WATCHLIST, SETTINGS)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "single_etf_limit")
        self.assertIn("QQQ", errors[0].detail)


class TestSectorConcentration(unittest.TestCase):
    """validate_sector_concentration"""

    def test_passes_at_60_percent_boundary(self):
        allocation = {
            "QQQ": {"amount": 2_000_000, "role": "core", "confidence": "high"},
            "069500": {"amount": 1_000_000, "role": "core", "confidence": "medium"},
            "NVDA": {"amount": 1_000_000, "role": "satellite", "confidence": "high"},
            "META": {"amount": 500_000, "role": "satellite", "confidence": "medium"},
            "ASML": {"amount": 500_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_sector_concentration(allocation, WATCHLIST, SETTINGS)
        self.assertEqual(errors, [])

    def test_rejects_sector_exceeding_60_percent(self):
        allocation = {
            "QQQ": {"amount": 1_300_000, "role": "core", "confidence": "high"},
            "069500": {"amount": 500_000, "role": "core", "confidence": "medium"},
            "NVDA": {"amount": 1_600_000, "role": "satellite", "confidence": "high"},
            "ASML": {"amount": 1_500_000, "role": "satellite", "confidence": "high"},
            "META": {"amount": 100_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_sector_concentration(allocation, WATCHLIST, SETTINGS)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "sector_concentration")
        self.assertIn("semiconductor", errors[0].detail)


class TestMinPosition(unittest.TestCase):
    """validate_min_position"""

    def test_passes_all_above_minimum(self):
        errors = validate_min_position(VALID_ALLOCATION, SETTINGS)
        self.assertEqual(errors, [])

    def test_rejects_position_below_minimum(self):
        allocation = {
            "QQQ": {"amount": 2_500_000, "role": "core", "confidence": "high"},
            "069500": {"amount": 1_000_000, "role": "core", "confidence": "medium"},
            "NVDA": {"amount": 1_400_000, "role": "satellite", "confidence": "high"},
            "ASML": {"amount": 50_000, "role": "satellite", "confidence": "low"},
            "META": {"amount": 50_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_min_position(allocation, SETTINGS)
        self.assertEqual(len(errors), 2)
        symbols = {e.detail.split(":")[0].strip() for e in errors}
        self.assertEqual(symbols, {"ASML", "META"})


class TestCoreInternalRatio(unittest.TestCase):
    """validate_core_internal_ratio"""

    def test_passes_at_2_to_1_ratio(self):
        allocation = {
            "QQQ": {"amount": 2_400_000, "role": "core", "confidence": "high"},
            "069500": {"amount": 1_100_000, "role": "core", "confidence": "medium"},
            "NVDA": {"amount": 600_000, "role": "satellite", "confidence": "high"},
            "ASML": {"amount": 500_000, "role": "satellite", "confidence": "medium"},
            "META": {"amount": 400_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_core_internal_ratio(allocation, SETTINGS)
        self.assertEqual(errors, [])

    def test_passes_within_5pp_tolerance(self):
        allocation = {
            "QQQ": {"amount": 2_200_000, "role": "core", "confidence": "high"},
            "069500": {"amount": 1_300_000, "role": "core", "confidence": "medium"},
            "NVDA": {"amount": 600_000, "role": "satellite", "confidence": "high"},
            "ASML": {"amount": 500_000, "role": "satellite", "confidence": "medium"},
            "META": {"amount": 400_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_core_internal_ratio(allocation, SETTINGS)
        self.assertEqual(errors, [])

    def test_rejects_inverted_ratio(self):
        allocation = {
            "QQQ": {"amount": 1_000_000, "role": "core", "confidence": "high"},
            "069500": {"amount": 2_500_000, "role": "core", "confidence": "medium"},
            "NVDA": {"amount": 600_000, "role": "satellite", "confidence": "high"},
            "ASML": {"amount": 500_000, "role": "satellite", "confidence": "medium"},
            "META": {"amount": 400_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_core_internal_ratio(allocation, SETTINGS)
        self.assertGreaterEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "core_internal_ratio")

    def test_skips_when_only_one_core_symbol_in_ratio(self):
        allocation = {
            "QQQ": {"amount": 3_500_000, "role": "core", "confidence": "high"},
            "NVDA": {"amount": 600_000, "role": "satellite", "confidence": "high"},
            "ASML": {"amount": 500_000, "role": "satellite", "confidence": "medium"},
            "META": {"amount": 400_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_core_internal_ratio(allocation, SETTINGS)
        self.assertEqual(errors, [])

    def test_skips_when_no_ratio_configured(self):
        settings_no_ratio = {
            **SETTINGS,
            "strategy": {"core_pct": 70, "satellite_pct": 30},
        }
        allocation = {
            "QQQ": {"amount": 2_000_000, "role": "core", "confidence": "high"},
            "069500": {"amount": 1_500_000, "role": "core", "confidence": "medium"},
            "NVDA": {"amount": 600_000, "role": "satellite", "confidence": "high"},
            "ASML": {"amount": 500_000, "role": "satellite", "confidence": "medium"},
            "META": {"amount": 400_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_core_internal_ratio(allocation, settings_no_ratio)
        self.assertEqual(errors, [])


class TestCoreSatelliteRatio(unittest.TestCase):
    """validate_core_satellite_ratio"""

    def test_passes_at_70_percent(self):
        allocation = {
            "QQQ": {"amount": 2_300_000, "role": "core", "confidence": "high"},
            "069500": {"amount": 1_200_000, "role": "core", "confidence": "medium"},
            "NVDA": {"amount": 600_000, "role": "satellite", "confidence": "medium"},
            "ASML": {"amount": 500_000, "role": "satellite", "confidence": "medium"},
            "META": {"amount": 400_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_core_satellite_ratio(allocation, SETTINGS)
        self.assertEqual(errors, [])

    def test_passes_at_65_percent_lower_boundary(self):
        allocation = {
            "QQQ": {"amount": 2_250_000, "role": "core", "confidence": "high"},
            "069500": {"amount": 1_000_000, "role": "core", "confidence": "medium"},
            "NVDA": {"amount": 1_200_000, "role": "satellite", "confidence": "high"},
            "META": {"amount": 550_000, "role": "satellite", "confidence": "medium"},
        }
        errors = validate_core_satellite_ratio(allocation, SETTINGS)
        self.assertEqual(errors, [])

    def test_passes_at_75_percent_upper_boundary(self):
        allocation = {
            "QQQ": {"amount": 2_500_000, "role": "core", "confidence": "high"},
            "069500": {"amount": 1_250_000, "role": "core", "confidence": "medium"},
            "NVDA": {"amount": 800_000, "role": "satellite", "confidence": "high"},
            "META": {"amount": 450_000, "role": "satellite", "confidence": "medium"},
        }
        errors = validate_core_satellite_ratio(allocation, SETTINGS)
        self.assertEqual(errors, [])

    def test_rejects_below_65_percent(self):
        allocation = {
            "QQQ": {"amount": 1_500_000, "role": "core", "confidence": "high"},
            "069500": {"amount": 1_600_000, "role": "satellite", "confidence": "medium"},
            "NVDA": {"amount": 1_200_000, "role": "satellite", "confidence": "high"},
            "META": {"amount": 700_000, "role": "satellite", "confidence": "medium"},
        }
        errors = validate_core_satellite_ratio(allocation, SETTINGS)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "core_satellite_ratio")

    def test_rejects_above_75_percent(self):
        allocation = {
            "QQQ": {"amount": 2_500_000, "role": "core", "confidence": "high"},
            "069500": {"amount": 1_400_000, "role": "core", "confidence": "medium"},
            "NVDA": {"amount": 700_000, "role": "satellite", "confidence": "high"},
            "META": {"amount": 400_000, "role": "satellite", "confidence": "medium"},
        }
        errors = validate_core_satellite_ratio(allocation, SETTINGS)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "core_satellite_ratio")


class TestAnchoring(unittest.TestCase):
    """validate_anchoring"""

    def test_skips_without_previous(self):
        errors = validate_anchoring(VALID_ALLOCATION, None, "monthly", SETTINGS)
        self.assertEqual(errors, [])

    def test_passes_within_monthly_per_position_limit(self):
        previous = {
            "QQQ": {"amount": 2_000_000, "role": "core"},
            "NVDA": {"amount": 1_500_000, "role": "core"},
            "069500": {"amount": 500_000, "role": "satellite"},
            "ASML": {"amount": 500_000, "role": "satellite"},
            "META": {"amount": 500_000, "role": "satellite"},
        }
        current = {
            "QQQ": {"amount": 2_200_000, "role": "core", "confidence": "high"},
            "NVDA": {"amount": 1_300_000, "role": "core", "confidence": "high"},
            "069500": {"amount": 500_000, "role": "satellite", "confidence": "medium"},
            "ASML": {"amount": 500_000, "role": "satellite", "confidence": "medium"},
            "META": {"amount": 500_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_anchoring(current, previous, "monthly", SETTINGS)
        self.assertEqual(errors, [])

    def test_rejects_monthly_per_position_exceeding_limit(self):
        previous = {
            "QQQ": {"amount": 2_000_000, "role": "core"},
            "NVDA": {"amount": 1_500_000, "role": "core"},
            "069500": {"amount": 500_000, "role": "satellite"},
            "ASML": {"amount": 500_000, "role": "satellite"},
            "META": {"amount": 500_000, "role": "satellite"},
        }
        current = {
            "QQQ": {"amount": 2_300_000, "role": "core", "confidence": "high"},
            "NVDA": {"amount": 1_300_000, "role": "core", "confidence": "high"},
            "069500": {"amount": 400_000, "role": "satellite", "confidence": "medium"},
            "ASML": {"amount": 500_000, "role": "satellite", "confidence": "medium"},
            "META": {"amount": 500_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_anchoring(current, previous, "monthly", SETTINGS)
        per_position_errors = [e for e in errors if "per_position" in e.rule]
        self.assertEqual(len(per_position_errors), 1)
        self.assertIn("QQQ", per_position_errors[0].detail)

    def test_rejects_monthly_total_exceeding_limit(self):
        previous = {
            "QQQ": {"amount": 2_000_000, "role": "core"},
            "NVDA": {"amount": 1_300_000, "role": "core"},
            "069500": {"amount": 700_000, "role": "satellite"},
            "ASML": {"amount": 500_000, "role": "satellite"},
            "META": {"amount": 500_000, "role": "satellite"},
        }
        current = {
            "QQQ": {"amount": 2_200_000, "role": "core", "confidence": "high"},
            "NVDA": {"amount": 1_100_000, "role": "core", "confidence": "high"},
            "069500": {"amount": 900_000, "role": "satellite", "confidence": "medium"},
            "ASML": {"amount": 500_000, "role": "satellite", "confidence": "medium"},
            "META": {"amount": 300_000, "role": "satellite", "confidence": "low"},
        }
        errors = validate_anchoring(current, previous, "monthly", SETTINGS)
        total_errors = [e for e in errors if "total" in e.rule]
        self.assertGreaterEqual(len(total_errors), 1)


class TestConfidencePool(unittest.TestCase):
    """validate_confidence_pool"""

    def test_passes_when_pools_sum_to_100(self):
        errors = validate_confidence_pool(SETTINGS)
        self.assertEqual(errors, [])

    def test_rejects_pools_not_summing_to_100(self):
        bad_settings = {
            **SETTINGS,
            "sizing": {
                "high_confidence_pool_pct": 50,
                "medium_confidence_pool_pct": 30,
                "low_confidence_pool_pct": 15,
            },
        }
        errors = validate_confidence_pool(bad_settings)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "confidence_pool_total")
        self.assertEqual(errors[0].actual, 95)


class TestWatchlistMembership(unittest.TestCase):
    """validate_watchlist_membership"""

    def test_passes_all_in_watchlist(self):
        errors = validate_watchlist_membership(VALID_ALLOCATION, WATCHLIST)
        self.assertEqual(errors, [])

    def test_rejects_unknown_symbol(self):
        allocation = {
            "QQQ": {"amount": 2_500_000, "role": "core", "confidence": "high"},
            "TSLA": {"amount": 1_000_000, "role": "satellite", "confidence": "high"},
            "069500": {"amount": 1_000_000, "role": "core", "confidence": "medium"},
            "META": {"amount": 500_000, "role": "satellite", "confidence": "medium"},
        }
        errors = validate_watchlist_membership(allocation, WATCHLIST)
        self.assertEqual(len(errors), 1)
        self.assertIn("TSLA", errors[0].detail)


class TestValidateAllocation(unittest.TestCase):
    """validate_allocation — integration of all rules"""

    def test_passes_valid_allocation(self):
        result = validate_allocation(VALID_ALLOCATION, WATCHLIST, SETTINGS)
        self.assertEqual(result.status, "PASS")
        self.assertEqual(result.errors, [])

    def test_collects_multiple_errors(self):
        allocation = {
            "QQQ": {"amount": 2_800_000, "role": "core", "confidence": "high"},
            "TSLA": {"amount": 1_650_000, "role": "satellite", "confidence": "high"},
            "META": {"amount": 550_050, "role": "satellite", "confidence": "medium"},
        }
        result = validate_allocation(allocation, WATCHLIST, SETTINGS)
        self.assertEqual(result.status, "FAIL")
        rules = {e.rule for e in result.errors}
        self.assertIn("budget_total", rules)
        self.assertIn("adjustment_unit", rules)
        self.assertIn("watchlist_membership", rules)


if __name__ == "__main__":
    unittest.main()

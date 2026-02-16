"""Unit tests for validate_holdings module."""

import unittest
import sys
from datetime import date
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from validate_holdings import (
    HoldingsError,
    HoldingsWarning,
    validate_positions,
    validate_watchlist_cross_ref,
    validate_currency_market_match,
    validate_staleness,
    validate_holdings,
)

WATCHLIST = [
    {"symbol": "ASML", "type": "stock", "sector": "semiconductor", "market": "US"},
    {"symbol": "META", "type": "stock", "sector": "big-tech", "market": "US"},
    {"symbol": "NVDA", "type": "stock", "sector": "semiconductor", "market": "US"},
    {"symbol": "QQQ", "type": "etf", "sector": "us-broad-market", "market": "US"},
    {"symbol": "069500", "type": "etf", "sector": "kr-broad-market", "market": "KR"},
]

VALID_HOLDINGS = {
    "as_of": "2026-02-10",
    "positions": {
        "NVDA": {"quantity": 3, "avg_cost": 150.50, "currency": "USD"},
        "QQQ": {"quantity": 2.5, "avg_cost": 580.00, "currency": "USD"},
        "069500": {"quantity": 10, "avg_cost": 78000, "currency": "KRW"},
    },
}


class TestValidatePositions(unittest.TestCase):
    """validate_positions"""

    def test_passes_valid_positions(self):
        errors = validate_positions(VALID_HOLDINGS)
        self.assertEqual(errors, [])

    def test_rejects_negative_quantity(self):
        holdings = {
            "as_of": "2026-02-10",
            "positions": {
                "NVDA": {"quantity": -1, "avg_cost": 150.00, "currency": "USD"},
            },
        }
        errors = validate_positions(holdings)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "quantity_non_negative")
        self.assertIn("NVDA", errors[0].detail)

    def test_rejects_zero_avg_cost(self):
        holdings = {
            "as_of": "2026-02-10",
            "positions": {
                "NVDA": {"quantity": 1, "avg_cost": 0, "currency": "USD"},
            },
        }
        errors = validate_positions(holdings)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "avg_cost_positive")

    def test_rejects_negative_avg_cost(self):
        holdings = {
            "as_of": "2026-02-10",
            "positions": {
                "NVDA": {"quantity": 1, "avg_cost": -100, "currency": "USD"},
            },
        }
        errors = validate_positions(holdings)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "avg_cost_positive")

    def test_passes_fractional_quantity(self):
        """Fractional shares (e.g. 0.5) from amount-based purchases."""
        holdings = {
            "as_of": "2026-02-10",
            "positions": {
                "NVDA": {"quantity": 0.5, "avg_cost": 150.00, "currency": "USD"},
            },
        }
        errors = validate_positions(holdings)
        self.assertEqual(errors, [])

    def test_passes_zero_quantity(self):
        """Zero quantity is valid (sold all shares but keeping record)."""
        holdings = {
            "as_of": "2026-02-10",
            "positions": {
                "NVDA": {"quantity": 0, "avg_cost": 150.00, "currency": "USD"},
            },
        }
        errors = validate_positions(holdings)
        self.assertEqual(errors, [])

    def test_rejects_invalid_currency(self):
        holdings = {
            "as_of": "2026-02-10",
            "positions": {
                "NVDA": {"quantity": 1, "avg_cost": 150.00, "currency": "EUR"},
            },
        }
        errors = validate_positions(holdings)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "currency_enum")

    def test_rejects_missing_required_fields(self):
        holdings = {
            "as_of": "2026-02-10",
            "positions": {
                "NVDA": {"quantity": 1},
            },
        }
        errors = validate_positions(holdings)
        rules = {e.rule for e in errors}
        self.assertIn("missing_field", rules)

    def test_rejects_non_numeric_quantity(self):
        holdings = {
            "as_of": "2026-02-10",
            "positions": {
                "NVDA": {"quantity": "three", "avg_cost": 150.00, "currency": "USD"},
            },
        }
        errors = validate_positions(holdings)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "quantity_numeric")

    def test_passes_empty_positions(self):
        """Empty positions is valid for first-time setup."""
        holdings = {
            "as_of": "2026-02-10",
            "positions": {},
        }
        errors = validate_positions(holdings)
        self.assertEqual(errors, [])


class TestWatchlistCrossRef(unittest.TestCase):
    """validate_watchlist_cross_ref"""

    def test_passes_all_in_watchlist(self):
        warnings = validate_watchlist_cross_ref(VALID_HOLDINGS, WATCHLIST)
        self.assertEqual(warnings, [])

    def test_warns_symbol_not_in_watchlist(self):
        holdings = {
            "as_of": "2026-02-10",
            "positions": {
                "TSLA": {"quantity": 1, "avg_cost": 300.00, "currency": "USD"},
                "NVDA": {"quantity": 2, "avg_cost": 150.00, "currency": "USD"},
            },
        }
        warnings = validate_watchlist_cross_ref(holdings, WATCHLIST)
        self.assertEqual(len(warnings), 1)
        self.assertEqual(warnings[0].rule, "watchlist_cross_ref")
        self.assertIn("TSLA", warnings[0].detail)

    def test_no_warning_for_empty_positions(self):
        holdings = {"as_of": "2026-02-10", "positions": {}}
        warnings = validate_watchlist_cross_ref(holdings, WATCHLIST)
        self.assertEqual(warnings, [])


class TestCurrencyMarketMatch(unittest.TestCase):
    """validate_currency_market_match"""

    def test_passes_correct_currency_market_pairs(self):
        errors = validate_currency_market_match(VALID_HOLDINGS, WATCHLIST)
        self.assertEqual(errors, [])

    def test_rejects_us_stock_with_krw(self):
        holdings = {
            "as_of": "2026-02-10",
            "positions": {
                "NVDA": {"quantity": 1, "avg_cost": 200000, "currency": "KRW"},
            },
        }
        errors = validate_currency_market_match(holdings, WATCHLIST)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "currency_market_mismatch")
        self.assertIn("NVDA", errors[0].detail)

    def test_rejects_kr_stock_with_usd(self):
        holdings = {
            "as_of": "2026-02-10",
            "positions": {
                "069500": {"quantity": 10, "avg_cost": 55.00, "currency": "USD"},
            },
        }
        errors = validate_currency_market_match(holdings, WATCHLIST)
        self.assertEqual(len(errors), 1)
        self.assertEqual(errors[0].rule, "currency_market_mismatch")
        self.assertIn("069500", errors[0].detail)

    def test_skips_symbol_not_in_watchlist(self):
        """Symbols not in watchlist can't be cross-validated — handled by cross-ref warning."""
        holdings = {
            "as_of": "2026-02-10",
            "positions": {
                "TSLA": {"quantity": 1, "avg_cost": 300.00, "currency": "USD"},
            },
        }
        errors = validate_currency_market_match(holdings, WATCHLIST)
        self.assertEqual(errors, [])


class TestStaleness(unittest.TestCase):
    """validate_staleness"""

    def test_passes_recent_date(self):
        holdings = {"as_of": "2026-02-10", "positions": {}}
        reference = date(2026, 2, 16)
        warnings = validate_staleness(holdings, reference, 30)
        self.assertEqual(warnings, [])

    def test_warns_stale_date_over_30_days(self):
        holdings = {"as_of": "2026-01-01", "positions": {}}
        reference = date(2026, 2, 16)
        warnings = validate_staleness(holdings, reference, 30)
        self.assertEqual(len(warnings), 1)
        self.assertEqual(warnings[0].rule, "stale_holdings")
        self.assertIn("46", warnings[0].detail)

    def test_passes_at_exactly_30_days(self):
        holdings = {"as_of": "2026-01-17", "positions": {}}
        reference = date(2026, 2, 16)
        warnings = validate_staleness(holdings, reference, 30)
        self.assertEqual(warnings, [])

    def test_warns_at_31_days(self):
        holdings = {"as_of": "2026-01-16", "positions": {}}
        reference = date(2026, 2, 16)
        warnings = validate_staleness(holdings, reference, 30)
        self.assertEqual(len(warnings), 1)
        self.assertEqual(warnings[0].rule, "stale_holdings")

    def test_rejects_invalid_date_format(self):
        holdings = {"as_of": "02/16/2026", "positions": {}}
        reference = date(2026, 2, 16)
        warnings = validate_staleness(holdings, reference, 30)
        self.assertEqual(len(warnings), 1)
        self.assertEqual(warnings[0].rule, "invalid_as_of_date")

    def test_respects_custom_threshold(self):
        holdings = {"as_of": "2026-02-01", "positions": {}}
        reference = date(2026, 2, 16)
        warnings_default = validate_staleness(holdings, reference, 30)
        self.assertEqual(warnings_default, [])
        warnings_strict = validate_staleness(holdings, reference, 7)
        self.assertEqual(len(warnings_strict), 1)
        self.assertEqual(warnings_strict[0].rule, "stale_holdings")


class TestValidateHoldings(unittest.TestCase):
    """validate_holdings — integration of all rules"""

    def test_passes_valid_holdings(self):
        reference = date(2026, 2, 16)
        result = validate_holdings(VALID_HOLDINGS, WATCHLIST, reference)
        self.assertEqual(result.status, "PASS")
        self.assertEqual(result.errors, [])
        self.assertEqual(result.warnings, [])

    def test_returns_fail_on_negative_quantity(self):
        holdings = {
            "as_of": "2026-02-10",
            "positions": {
                "NVDA": {"quantity": -1, "avg_cost": 150.00, "currency": "USD"},
            },
        }
        reference = date(2026, 2, 16)
        result = validate_holdings(holdings, WATCHLIST, reference)
        self.assertEqual(result.status, "FAIL")
        error_rules = {e.rule for e in result.errors}
        self.assertIn("quantity_non_negative", error_rules)

    def test_returns_warn_on_stale_and_cross_ref(self):
        holdings = {
            "as_of": "2025-12-01",
            "positions": {
                "TSLA": {"quantity": 1, "avg_cost": 300.00, "currency": "USD"},
            },
        }
        reference = date(2026, 2, 16)
        result = validate_holdings(holdings, WATCHLIST, reference)
        self.assertEqual(result.status, "WARN")
        self.assertEqual(result.errors, [])
        warning_rules = {w.rule for w in result.warnings}
        self.assertIn("stale_holdings", warning_rules)
        self.assertIn("watchlist_cross_ref", warning_rules)

    def test_passes_empty_positions(self):
        holdings = {"as_of": "2026-02-10", "positions": {}}
        reference = date(2026, 2, 16)
        result = validate_holdings(holdings, WATCHLIST, reference)
        self.assertEqual(result.status, "PASS")

    def test_fail_takes_precedence_over_warn(self):
        """When both errors and warnings exist, status is FAIL."""
        holdings = {
            "as_of": "2025-12-01",
            "positions": {
                "NVDA": {"quantity": -1, "avg_cost": 150.00, "currency": "USD"},
            },
        }
        reference = date(2026, 2, 16)
        result = validate_holdings(holdings, WATCHLIST, reference)
        self.assertEqual(result.status, "FAIL")
        self.assertGreater(len(result.errors), 0)
        self.assertGreater(len(result.warnings), 0)

    def test_passes_without_reference_date(self):
        """When no reference date, skip staleness check."""
        result = validate_holdings(VALID_HOLDINGS, WATCHLIST)
        self.assertEqual(result.status, "PASS")

    def test_rejects_missing_as_of(self):
        holdings = {"positions": {}}
        result = validate_holdings(holdings, WATCHLIST)
        self.assertEqual(result.status, "FAIL")
        error_rules = {e.rule for e in result.errors}
        self.assertIn("missing_as_of", error_rules)

    def test_rejects_missing_positions(self):
        holdings = {"as_of": "2026-02-10"}
        result = validate_holdings(holdings, WATCHLIST)
        self.assertEqual(result.status, "FAIL")
        error_rules = {e.rule for e in result.errors}
        self.assertIn("missing_positions", error_rules)


if __name__ == "__main__":
    unittest.main()

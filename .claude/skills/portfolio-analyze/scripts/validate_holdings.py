#!/usr/bin/env python3
"""Validate holdings data against schema and watchlist constraints.

Exit 0: PASS, WARN, or FAIL — structured JSON on stdout.
Exit 1: ERROR — input parsing or file access failure.

Severity levels:
  PASS — all validations passed
  WARN — non-critical issues (stale date, watchlist mismatch)
  FAIL — critical issues (negative quantity, missing fields, currency mismatch)
"""

import argparse
import json
import sys
from datetime import date, timedelta
from typing import NamedTuple, Optional


_VALID_CURRENCIES = {"USD", "KRW"}
_MARKET_CURRENCY = {"US": "USD", "KR": "KRW"}
_REQUIRED_POSITION_FIELDS = {"quantity", "avg_cost", "currency"}
_DEFAULT_STALENESS_THRESHOLD_DAYS = 30


class HoldingsError(NamedTuple):
    rule: str
    detail: str


class HoldingsWarning(NamedTuple):
    rule: str
    detail: str


class ValidationResult(NamedTuple):
    status: str
    errors: list[HoldingsError]
    warnings: list[HoldingsWarning]


def validate_positions(holdings: dict) -> list[HoldingsError]:
    positions = holdings.get("positions", {})
    errors = []

    for symbol in sorted(positions):
        pos = positions[symbol]

        missing = _REQUIRED_POSITION_FIELDS - set(pos.keys())
        if missing:
            errors.append(
                HoldingsError(
                    rule="missing_field",
                    detail=f"{symbol}: missing required fields: {sorted(missing)}",
                )
            )
            continue

        if not isinstance(pos["quantity"], (int, float)):
            errors.append(
                HoldingsError(
                    rule="quantity_numeric",
                    detail=f"{symbol}: quantity must be numeric, got {type(pos['quantity']).__name__}",
                )
            )
        elif pos["quantity"] < 0:
            errors.append(
                HoldingsError(
                    rule="quantity_non_negative",
                    detail=f"{symbol}: quantity must be >= 0, got {pos['quantity']}",
                )
            )

        if not isinstance(pos["avg_cost"], (int, float)):
            errors.append(
                HoldingsError(
                    rule="avg_cost_numeric",
                    detail=f"{symbol}: avg_cost must be numeric, got {type(pos['avg_cost']).__name__}",
                )
            )
        elif pos["avg_cost"] <= 0:
            errors.append(
                HoldingsError(
                    rule="avg_cost_positive",
                    detail=f"{symbol}: avg_cost must be > 0, got {pos['avg_cost']}",
                )
            )

        currency = pos.get("currency")
        if currency not in _VALID_CURRENCIES:
            errors.append(
                HoldingsError(
                    rule="currency_enum",
                    detail=f"{symbol}: currency must be one of {sorted(_VALID_CURRENCIES)}, got '{currency}'",
                )
            )

    return errors


def validate_watchlist_cross_ref(
    holdings: dict, watchlist: list[dict]
) -> list[HoldingsWarning]:
    valid_symbols = {item["symbol"] for item in watchlist}
    warnings = []

    for symbol in sorted(holdings.get("positions", {})):
        if symbol not in valid_symbols:
            warnings.append(
                HoldingsWarning(
                    rule="watchlist_cross_ref",
                    detail=f"{symbol}: not found in watchlist",
                )
            )

    return warnings


def validate_currency_market_match(
    holdings: dict, watchlist: list[dict]
) -> list[HoldingsError]:
    symbol_market = {item["symbol"]: item["market"] for item in watchlist}
    errors = []

    for symbol in sorted(holdings.get("positions", {})):
        pos = holdings["positions"][symbol]
        market = symbol_market.get(symbol)
        if market is None:
            continue

        expected_currency = _MARKET_CURRENCY.get(market)
        actual_currency = pos.get("currency")
        if expected_currency and actual_currency != expected_currency:
            errors.append(
                HoldingsError(
                    rule="currency_market_mismatch",
                    detail=f"{symbol}: market {market} requires {expected_currency}, got {actual_currency}",
                )
            )

    return errors


def validate_staleness(
    holdings: dict, reference: date, threshold_days: int
) -> list[HoldingsWarning]:
    as_of_str = holdings.get("as_of", "")
    warnings = []

    try:
        as_of = date.fromisoformat(as_of_str)
    except (ValueError, TypeError):
        warnings.append(
            HoldingsWarning(
                rule="invalid_as_of_date",
                detail=f"as_of '{as_of_str}' is not a valid ISO date (YYYY-MM-DD)",
            )
        )
        return warnings

    days_old = (reference - as_of).days
    if days_old > threshold_days:
        warnings.append(
            HoldingsWarning(
                rule="stale_holdings",
                detail=f"as_of ({as_of_str}) is {days_old} days old (threshold: {threshold_days})",
            )
        )

    return warnings


def validate_holdings(
    holdings: dict,
    watchlist: list[dict],
    reference: Optional[date] = None,
    staleness_threshold_days: int = _DEFAULT_STALENESS_THRESHOLD_DAYS,
) -> ValidationResult:
    errors: list[HoldingsError] = []
    warnings: list[HoldingsWarning] = []

    if "as_of" not in holdings:
        errors.append(
            HoldingsError(rule="missing_as_of", detail="holdings must contain 'as_of' field")
        )
    if "positions" not in holdings:
        errors.append(
            HoldingsError(rule="missing_positions", detail="holdings must contain 'positions' field")
        )

    if errors:
        return ValidationResult(status="FAIL", errors=errors, warnings=warnings)

    errors.extend(validate_positions(holdings))
    errors.extend(validate_currency_market_match(holdings, watchlist))
    warnings.extend(validate_watchlist_cross_ref(holdings, watchlist))

    if reference is not None:
        warnings.extend(validate_staleness(holdings, reference, staleness_threshold_days))

    if errors:
        status = "FAIL"
    elif warnings:
        status = "WARN"
    else:
        status = "PASS"

    return ValidationResult(status=status, errors=errors, warnings=warnings)


def _serialize_result(result: ValidationResult) -> dict:
    output: dict = {"status": result.status}
    if result.errors:
        output["errors"] = [
            {"rule": e.rule, "detail": e.detail} for e in result.errors
        ]
    if result.warnings:
        output["warnings"] = [
            {"rule": w.rule, "detail": w.detail} for w in result.warnings
        ]
    return output


def _parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Validate holdings data against schema and watchlist constraints"
    )
    parser.add_argument("holdings_path", help="Path to holdings.json")
    parser.add_argument("watchlist_path", help="Path to watchlist.json")
    parser.add_argument(
        "--settings",
        help="Path to settings.json (reads holdings_staleness_threshold_days)",
    )
    parser.add_argument(
        "--reference-date",
        help="Reference date for staleness check (YYYY-MM-DD). Defaults to today.",
    )
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = _parse_args(argv)

    try:
        with open(args.holdings_path) as f:
            holdings = json.load(f)
        with open(args.watchlist_path) as f:
            watchlist = json.load(f)
    except (json.JSONDecodeError, FileNotFoundError, OSError) as exc:
        print(
            json.dumps({"status": "ERROR", "detail": str(exc)}),
            file=sys.stderr,
        )
        return 1

    staleness_threshold = _DEFAULT_STALENESS_THRESHOLD_DAYS
    if args.settings:
        try:
            with open(args.settings) as f:
                settings = json.load(f)
            staleness_threshold = settings.get(
                "holdings_staleness_threshold_days", _DEFAULT_STALENESS_THRESHOLD_DAYS
            )
        except (json.JSONDecodeError, FileNotFoundError, OSError) as exc:
            print(
                json.dumps({"status": "ERROR", "detail": str(exc)}),
                file=sys.stderr,
            )
            return 1

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

    result = validate_holdings(holdings, watchlist, reference, staleness_threshold)
    print(json.dumps(_serialize_result(result), ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())

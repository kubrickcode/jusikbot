#!/usr/bin/env python3
"""Validate portfolio allocation against settings and watchlist constraints.

Exit 0: PASS or FAIL — structured JSON on stdout.
Exit 1: ERROR — input parsing or file access failure.
"""

import argparse
import json
import sys
from typing import NamedTuple, Optional


_CORE_SATELLITE_TOLERANCE = 0.05  # ±5 percentage points
_FLOAT_EPSILON = 1e-9


class AllocationError(NamedTuple):
    rule: str
    expected: str | int | float
    actual: str | int | float
    detail: str


class ValidationResult(NamedTuple):
    status: str
    errors: list[AllocationError]


def validate_budget_total(
    allocations: dict, settings: dict
) -> list[AllocationError]:
    budget = settings["budget_krw"]
    total = sum(pos["amount"] for pos in allocations.values())
    if total != budget:
        direction = "exceeds" if total > budget else "below"
        return [
            AllocationError(
                rule="budget_total",
                expected=budget,
                actual=total,
                detail=f"Total {direction} budget by {abs(total - budget)}",
            )
        ]
    return []


def validate_adjustment_unit(
    allocations: dict, settings: dict
) -> list[AllocationError]:
    unit = settings["adjustment_unit_krw"]
    errors = []
    for symbol, pos in sorted(allocations.items()):
        if pos["amount"] % unit != 0:
            errors.append(
                AllocationError(
                    rule="adjustment_unit",
                    expected=f"multiple of {unit}",
                    actual=pos["amount"],
                    detail=f"{symbol}: {pos['amount']} is not a multiple of {unit}",
                )
            )
    return errors


def validate_position_limits(
    allocations: dict, watchlist: list[dict], settings: dict
) -> list[AllocationError]:
    budget = settings["budget_krw"]
    max_stock_pct = settings["risk_tolerance"]["max_single_stock_pct"] / 100
    max_etf_pct = settings["risk_tolerance"]["max_single_etf_pct"] / 100
    symbol_type = {item["symbol"]: item["type"] for item in watchlist}

    errors = []
    for symbol, pos in sorted(allocations.items()):
        pct = pos["amount"] / budget
        asset_type = symbol_type.get(symbol)

        if asset_type == "stock" and pct > max_stock_pct:
            errors.append(
                AllocationError(
                    rule="single_stock_limit",
                    expected=f"<= {max_stock_pct * 100:.0f}%",
                    actual=f"{pct * 100:.1f}%",
                    detail=f"{symbol}: stock position {pct * 100:.1f}% exceeds {max_stock_pct * 100:.0f}% limit",
                )
            )
        elif asset_type == "etf" and pct > max_etf_pct:
            errors.append(
                AllocationError(
                    rule="single_etf_limit",
                    expected=f"<= {max_etf_pct * 100:.0f}%",
                    actual=f"{pct * 100:.1f}%",
                    detail=f"{symbol}: ETF position {pct * 100:.1f}% exceeds {max_etf_pct * 100:.0f}% limit",
                )
            )
    return errors


def validate_sector_concentration(
    allocations: dict, watchlist: list[dict], settings: dict
) -> list[AllocationError]:
    budget = settings["budget_krw"]
    max_pct = settings["risk_tolerance"]["max_sector_concentration_pct"] / 100
    symbol_sector = {item["symbol"]: item["sector"] for item in watchlist}

    sector_totals: dict[str, int] = {}
    for symbol, pos in allocations.items():
        sector = symbol_sector.get(symbol, "unknown")
        sector_totals[sector] = sector_totals.get(sector, 0) + pos["amount"]

    errors = []
    for sector in sorted(sector_totals):
        total = sector_totals[sector]
        pct = total / budget
        if pct > max_pct:
            errors.append(
                AllocationError(
                    rule="sector_concentration",
                    expected=f"<= {max_pct * 100:.0f}%",
                    actual=f"{pct * 100:.1f}%",
                    detail=f"{sector}: sector concentration {pct * 100:.1f}% exceeds {max_pct * 100:.0f}% limit",
                )
            )
    return errors


def validate_min_position(
    allocations: dict, settings: dict
) -> list[AllocationError]:
    min_size = settings["risk_tolerance"]["min_position_size_krw"]
    errors = []
    for symbol, pos in sorted(allocations.items()):
        if pos["amount"] < min_size:
            errors.append(
                AllocationError(
                    rule="min_position",
                    expected=f">= {min_size}",
                    actual=pos["amount"],
                    detail=f"{symbol}: position {pos['amount']} below minimum {min_size}",
                )
            )
    return errors


def validate_core_satellite_ratio(
    allocations: dict, settings: dict
) -> list[AllocationError]:
    budget = settings["budget_krw"]
    target_core_pct = settings["strategy"]["core_pct"] / 100

    core_total = sum(
        pos["amount"] for pos in allocations.values() if pos["role"] == "core"
    )
    core_pct = core_total / budget

    # Constraint: ±5%p inclusive boundary — IEEE 754 epsilon for exact boundary values
    if abs(core_pct - target_core_pct) > _CORE_SATELLITE_TOLERANCE + _FLOAT_EPSILON:
        lower = (target_core_pct - _CORE_SATELLITE_TOLERANCE) * 100
        upper = (target_core_pct + _CORE_SATELLITE_TOLERANCE) * 100
        return [
            AllocationError(
                rule="core_satellite_ratio",
                expected=f"{lower:.0f}–{upper:.0f}%",
                actual=f"{core_pct * 100:.1f}%",
                detail=f"Core ratio {core_pct * 100:.1f}% outside {lower:.0f}–{upper:.0f}% range",
            )
        ]
    return []


def validate_core_internal_ratio(
    allocations: dict, settings: dict
) -> list[AllocationError]:
    ratio_map = settings.get("strategy", {}).get("core_internal_ratio")
    if not ratio_map:
        return []

    core_positions = {
        sym: pos for sym, pos in allocations.items() if pos["role"] == "core"
    }
    if not core_positions:
        return []

    ratio_symbols = set(ratio_map.keys())
    core_symbols = set(core_positions.keys())
    relevant = ratio_symbols & core_symbols
    if len(relevant) < 2:
        return []

    total_weight = sum(ratio_map[s] for s in relevant)
    errors = []
    for symbol in sorted(relevant):
        expected_pct = ratio_map[symbol] / total_weight
        core_total = sum(pos["amount"] for pos in core_positions.values())
        if core_total == 0:
            continue
        actual_pct = core_positions[symbol]["amount"] / core_total

        if abs(actual_pct - expected_pct) > _CORE_SATELLITE_TOLERANCE + _FLOAT_EPSILON:
            errors.append(
                AllocationError(
                    rule="core_internal_ratio",
                    expected=f"{expected_pct * 100:.0f}%",
                    actual=f"{actual_pct * 100:.1f}%",
                    detail=f"{symbol}: core internal ratio {actual_pct * 100:.1f}% vs expected {expected_pct * 100:.0f}% (±5%p tolerance)",
                )
            )

    return errors


def validate_anchoring(
    allocations: dict,
    previous: Optional[dict],
    review_type: str,
    settings: dict,
    current_holdings: Optional[dict] = None,
) -> list[AllocationError]:
    # Holdings reflect actual ownership; previous allocations are unexecuted proposals
    reference = current_holdings if current_holdings is not None else previous
    if reference is None:
        return []

    anchoring = settings["anchoring"]
    max_per = anchoring[f"{review_type}_max_change_per_position_krw"]
    max_total = anchoring[f"{review_type}_max_total_change_krw"]

    all_symbols = sorted(set(allocations.keys()) | set(reference.keys()))
    errors = []
    total_change = 0

    for symbol in all_symbols:
        old_amount = reference.get(symbol, {}).get("amount", 0)
        new_amount = allocations.get(symbol, {}).get("amount", 0)
        change = abs(new_amount - old_amount)
        total_change += change

        if change > max_per:
            errors.append(
                AllocationError(
                    rule=f"anchoring_per_position_{review_type}",
                    expected=f"<= {max_per}",
                    actual=change,
                    detail=f"{symbol}: change {change} exceeds {review_type} per-position limit {max_per}",
                )
            )

    if total_change > max_total:
        errors.append(
            AllocationError(
                rule=f"anchoring_total_{review_type}",
                expected=f"<= {max_total}",
                actual=total_change,
                detail=f"Total change {total_change} exceeds {review_type} limit {max_total}",
            )
        )

    return errors


def validate_confidence_pool(settings: dict) -> list[AllocationError]:
    sizing = settings["sizing"]
    total = (
        sizing["high_confidence_pool_pct"]
        + sizing["medium_confidence_pool_pct"]
        + sizing["low_confidence_pool_pct"]
    )
    if total != 100:
        return [
            AllocationError(
                rule="confidence_pool_total",
                expected=100,
                actual=total,
                detail=f"Confidence pools sum to {total}%, must equal 100%",
            )
        ]
    return []


def validate_watchlist_membership(
    allocations: dict, watchlist: list[dict]
) -> list[AllocationError]:
    valid_symbols = {item["symbol"] for item in watchlist}
    errors = []
    for symbol in sorted(allocations.keys()):
        if symbol not in valid_symbols:
            errors.append(
                AllocationError(
                    rule="watchlist_membership",
                    expected="symbol in watchlist",
                    actual=symbol,
                    detail=f"{symbol}: not found in watchlist",
                )
            )
    return errors


def validate_allocation(
    allocations: dict,
    watchlist: list[dict],
    settings: dict,
    previous: Optional[dict] = None,
    review_type: str = "monthly",
    current_holdings: Optional[dict] = None,
) -> ValidationResult:
    budget = settings.get("budget_krw", 0)
    if not isinstance(budget, (int, float)) or budget <= 0:
        return ValidationResult(
            status="FAIL",
            errors=[
                AllocationError(
                    rule="budget_positive",
                    expected="> 0",
                    actual=budget,
                    detail=f"budget_krw must be a positive number, got {budget}",
                )
            ],
        )

    errors: list[AllocationError] = []

    errors.extend(validate_confidence_pool(settings))
    errors.extend(validate_watchlist_membership(allocations, watchlist))
    errors.extend(validate_budget_total(allocations, settings))
    errors.extend(validate_adjustment_unit(allocations, settings))
    errors.extend(validate_position_limits(allocations, watchlist, settings))
    errors.extend(validate_sector_concentration(
        allocations, watchlist, settings))
    errors.extend(validate_min_position(allocations, settings))
    errors.extend(validate_core_satellite_ratio(allocations, settings))
    errors.extend(validate_core_internal_ratio(allocations, settings))
    errors.extend(validate_anchoring(
        allocations, previous, review_type, settings,
        current_holdings=current_holdings))

    status = "PASS" if not errors else "FAIL"
    return ValidationResult(status=status, errors=errors)


def _serialize_result(result: ValidationResult) -> dict:
    if result.status == "PASS":
        return {"status": "PASS"}
    return {
        "status": "FAIL",
        "errors": [
            {
                "rule": e.rule,
                "expected": e.expected,
                "actual": e.actual,
                "detail": e.detail,
            }
            for e in result.errors
        ],
    }


def _parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Validate portfolio allocation against constraints"
    )
    parser.add_argument("--settings", required=True,
                        help="Path to settings.json")
    parser.add_argument("--watchlist", required=True,
                        help="Path to watchlist.json")
    parser.add_argument("--allocations", required=True,
                        help="Allocation JSON string")
    parser.add_argument(
        "--previous-allocations", help="Previous allocation JSON string"
    )
    parser.add_argument(
        "--current-holdings",
        help="Current holdings KRW amounts JSON string (takes priority over --previous-allocations for anchoring)",
    )
    parser.add_argument(
        "--review-type",
        choices=["monthly", "quarterly"],
        default="monthly",
    )
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = _parse_args(argv)

    try:
        with open(args.settings) as f:
            settings = json.load(f)
        with open(args.watchlist) as f:
            watchlist = json.load(f)
        allocations = json.loads(args.allocations)
    except (json.JSONDecodeError, FileNotFoundError, OSError, KeyError, TypeError) as exc:
        print(
            json.dumps({"status": "ERROR", "detail": str(exc)}),
            file=sys.stderr,
        )
        return 1

    previous = None
    if args.previous_allocations:
        try:
            previous = json.loads(args.previous_allocations)
        except json.JSONDecodeError as exc:
            print(
                json.dumps({"status": "ERROR", "detail": str(exc)}),
                file=sys.stderr,
            )
            return 1

    current_holdings = None
    if args.current_holdings:
        try:
            current_holdings = json.loads(args.current_holdings)
        except json.JSONDecodeError as exc:
            print(
                json.dumps({"status": "ERROR", "detail": str(exc)}),
                file=sys.stderr,
            )
            return 1

    result = validate_allocation(
        allocations, watchlist, settings, previous, args.review_type,
        current_holdings=current_holdings,
    )
    print(json.dumps(_serialize_result(result), ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())

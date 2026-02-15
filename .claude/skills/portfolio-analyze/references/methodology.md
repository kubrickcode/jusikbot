# Investment Methodology — Reference Card

## Thesis Types

| Type       | Key Indicators                                 | Data Source                                           | Fallback When Insufficient           |
| ---------- | ---------------------------------------------- | ----------------------------------------------------- | ------------------------------------ |
| Macro      | FX rate trend, broad index vs 200D             | summary.md `vs 200D`, `Exchange Rate`; psql `fx_rate` | Cap confidence at Medium             |
| Sector     | Sector ETF relative performance, HV comparison | summary.md `vs Bench`, `HV 20D/60D`                   | Query psql for 60D sector comparison |
| Individual | Earnings-driven momentum, 52W position, volume | summary.md `52W Pos`, `Vol Ratio`, `5D/20D`           | Maintain current allocation          |
| Thematic   | Cross-asset correlation within theme           | summary.md multiple symbols in same theme             | Cap confidence at Medium             |

## Confidence Framework

| Level      | Role        | Rules                                                                                                                                                                                                                                            |
| ---------- | ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Medium** | **Default** | Starting point for ALL theses. No evidence requirement beyond data availability.                                                                                                                                                                 |
| **High**   | Upgrade     | Requires ALL: (1) 2+ evidence from different categories (Section 3), (2) at least 1 from Volatility or External, (3) no invalidation condition triggered. KR data quality warning targets: CANNOT upgrade to High on technical indicators alone. |
| **Low**    | Downgrade   | Triggered when: invalidation condition partially met OR data insufficient. Rules: no new entry permitted; existing positions reduce to 50% of current allocation.                                                                                |

**Overextension Qualifier**: If confidence = High AND `vs 200D` > +20% → append "진입 타이밍 주의" to thesis assessment.

**Pool Allocation**: Core positions use fixed `strategy.core_internal_ratio` (confidence pools do not apply). Satellite positions distribute by confidence pools from `settings.json`:

- High: `sizing.high_confidence_pool_pct`
- Medium: `sizing.medium_confidence_pool_pct`
- Low: `sizing.low_confidence_pool_pct`

## Evidence Independence

| Category          | Measures                  | Example Indicators (summary.md columns)                        |
| ----------------- | ------------------------- | -------------------------------------------------------------- |
| Trend             | Price direction over time | `52W Pos`, `vs 200D`, `Cross` (MA crossover)                   |
| Momentum          | Rate of change            | `5D`, `20D` (short/medium-term returns)                        |
| Volatility        | Price dispersion          | `HV 20D`, `HV 60D`, `Vol Ratio`                                |
| Relative Strength | Performance vs benchmark  | `vs Bench`                                                     |
| External          | Non-price data            | `$ARGUMENTS` (user-provided: earnings, guidance, macro events) |

**Correlation Rule**: Trend + Momentum + Relative Strength all track price direction — they confirm simultaneously in bull markets. High confidence MUST include at least 1 evidence from Volatility or External to ensure independence.

## Sizing

**Stage 1 — Core Fixed Allocation**:

- Core total: `settings.json` → `strategy.core_pct` of `budget_krw`
- Satellite total: `settings.json` → `strategy.satellite_pct` of `budget_krw`
- Core positions use `strategy.core_internal_ratio` (symbol-to-weight map) — fixed proportional split regardless of confidence level. Core internal ratio takes priority over confidence pools.

**Stage 2 — Satellite Confidence Pool Allocation**:
Within the satellite bucket ONLY, distribute by confidence pool percentages from `settings.json` (`sizing.*_confidence_pool_pct`).

- **Empty Pool Redistribution**: If a confidence pool has zero positions, redistribute that pool's percentage proportionally to non-empty pools in the satellite bucket. Example: if no High positions exist (50% pool empty) and Medium (35%) and Low (15%) have positions, redistribute as Medium 70% and Low 30% of satellite budget.

**Stage 3 — Equal Weight Within Pool**:
Within each satellite confidence pool, distribute equally among positions. Round each to nearest `adjustment_unit_krw` multiple.

**Remainder Handling**: After rounding, if total != `budget_krw`, absorb remainder into the largest core position.

**Minimum Position**: Every position MUST >= `risk_tolerance.min_position_size_krw`. If a position falls below minimum after sizing, either increase to minimum (borrowing from the same pool) or exclude entirely.

**Low Confidence Sizing**:

- New positions: NOT permitted.
- Existing positions: apply BOTH constraints, take the SMALLER result: (a) 50% of previous allocation amount, (b) the Low confidence pool share. First run (no previous report): only constraint (b) applies.

**Constraint Priority** (when rules conflict during retry): (1) budget_total, (2) min_position, (3) anchoring, (4) core_satellite_ratio, (5) core_internal_ratio, (6) position_limits, (7) sector_concentration.

**Drawdown Reference Point**: 52-week high from summary.md `52W H` column. Drawdown % = (52W H - Close) / 52W H \* 100.

- Warning threshold: `risk_tolerance.max_drawdown_warning_pct`
- Action threshold: `risk_tolerance.max_drawdown_action_pct` → reduce position by 50%

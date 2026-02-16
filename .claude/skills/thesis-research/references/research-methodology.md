# Research Methodology — Reference Card

## 1. Condition Evaluation Framework

### Status Definitions

| Status          | Definition                                           | Example                                                    |
| --------------- | ---------------------------------------------------- | ---------------------------------------------------------- |
| `met`           | Condition clearly satisfied with current data        | NVDA rev growth 48% > 30% threshold                        |
| `partially_met` | Some but not all aspects satisfied                   | 2 of 3 hyperscalers raised capex, 1 flat                   |
| `not_yet`       | Not satisfied but thesis timeline has not expired    | Smart glasses at 10M units, threshold is 50M               |
| `refuted`       | Clear evidence contradicting the condition           | Hyperscalers cut capex 2 consecutive quarters              |
| `unknown`       | Insufficient data to evaluate after reasonable search | No reliable source found after checking Tier 1-2 sources   |

### Source Count Requirements

- `met` or `refuted`: minimum 2 independent sources, at least 1 from Tier 1-2.
- `partially_met`: minimum 1 source from Tier 1-3.
- `not_yet`: minimum 1 source. Acceptable from any tier since this status merely confirms timeline is still open.
- `unknown`: declare after checking at least 2 Tier 1-2 sources and finding no relevant data.

### Conflicting Sources

When sources disagree on the same data point:

1. **Tier trumps tier.** Tier 1 source overrides Tier 3 on the same fact. Example: 10-Q filing revenue figure overrides a news article estimate.
2. **Same tier conflict.** If two Tier 1-2 sources conflict, report `partially_met` and note the discrepancy explicitly. Do not resolve the conflict by choosing one.
3. **Recency matters within same tier.** A Q4 2025 earnings transcript supersedes a Q3 2025 transcript for the same metric.

### Boundary Between Statuses

- `partially_met` upgrades to `met` when ALL sub-components of the condition are satisfied. If a condition has quantitative thresholds (e.g., "30%+ growth"), partial means the metric is positive but below threshold, or some sub-conditions are met but not all.
- `partially_met` downgrades to `not_yet` if the partial evidence turns out to be outdated (older than 2 quarters for a 12-month thesis) and no newer confirming data exists.
- `not_yet` becomes `refuted` only with explicit counter-evidence, not mere absence of positive data.

## 2. Source Quality Assessment

### Tier Definitions

| Tier | Sources                                                                                      | Reliability |
| ---- | -------------------------------------------------------------------------------------------- | ----------- |
| 1    | Company filings (10-K, 10-Q, 8-K, earnings transcripts), government data, exchange filings   | Highest     |
| 2    | Industry research (Gartner, IDC, Counterpoint, Semiconductor Industry Association), Bloomberg, Reuters, FT | High        |
| 3    | Major news outlets (WSJ, CNBC, TechCrunch), sell-side analyst notes, conference presentations | Moderate    |
| 4    | Blogs, social media, Reddit, opinion pieces, YouTube analysis                                 | Low         |

### Korean Market Sources

| Tier | Sources                                                                        |
| ---- | ------------------------------------------------------------------------------ |
| 1    | DART filings (dart.fss.or.kr), KRX exchange data, Bank of Korea statistics     |
| 2    | Korea Institute of Finance, KOSIS, major securities firm research (삼성증권, 미래에셋 리서치) |
| 3    | 한경, 매경, 이데일리 articles                                                    |
| 4    | Naver finance comments, individual blogs, YouTube                              |

### Status Change Rules

| Transition                                   | Minimum Source Requirement                    |
| -------------------------------------------- | --------------------------------------------- |
| Any status to `met`                          | 2+ sources, at least 1 Tier 1-2              |
| Any status to `refuted`                      | 2+ sources, at least 1 Tier 1-2              |
| Any status to `partially_met`                | 1+ source from Tier 1-3                       |
| `met`/`refuted` to `partially_met` (reversal) | 1 Tier 1-2 source showing changed conditions |

Tier 3-4 sources alone can flag a condition for review but cannot finalize `met` or `refuted`. When only Tier 3-4 evidence exists, cap at `partially_met` and note the evidence quality gap.

## 3. Thesis Overall Status Derivation

### Status Computation

Compute thesis status from its validity conditions (V) and invalidation conditions (I):

| Thesis Status   | Rule                                                                            |
| --------------- | ------------------------------------------------------------------------------- |
| `valid`         | Majority of V conditions are `met`/`partially_met` AND zero I conditions `met`/`refuted` in thesis-adverse direction |
| `weakening`     | Any I condition is `partially_met` OR majority of V conditions are `not_yet`/`unknown` |
| `invalidated`   | Any I condition is `refuted` (confirmed adverse) OR majority of V conditions are `refuted` |
| `insufficient`  | Majority of both V and I conditions are `unknown`                               |

### Asymmetric Weighting

Invalidation conditions carry more weight than validity conditions. Rationale: protecting capital matters more than capturing upside for a small, growing portfolio.

- A single I condition at `refuted` triggers `invalidated` regardless of V condition status.
- All V conditions at `met` with one I condition at `partially_met` yields `weakening`, not `valid`.
- This asymmetry is intentional: it is the thesis-level equivalent of the bias guardrails' loss-aversion countermeasure.

### Edge Case: Mixed Signals

When V conditions are strongly `met` but an I condition is `partially_met`:

1. Report thesis as `weakening` (not `valid`).
2. In the thesis assessment, explicitly note the V-I tension.
3. Recommend increased monitoring frequency (next review in 2 weeks instead of monthly).

## 4. Candidate Company Discovery

### Hard Filters (Exclude Immediately)

| Filter              | Threshold                                                                 |
| ------------------- | ------------------------------------------------------------------------- |
| Market cap          | Minimum $5B USD (or KRW 6.5T equivalent). Avoids micro/small-cap volatility inappropriate for a small portfolio. |
| Average daily volume | Minimum $10M USD daily traded value (20-day average). Ensures executable position sizes. |
| Market access       | US (NYSE, NASDAQ) or KR (KRX: KOSPI, KOSDAQ). EU stocks acceptable only via US-listed ADR. Other markets excluded. |
| Listing history     | Minimum 2 years public trading history. No recent IPOs/SPACs.            |

### Soft Filters (Evaluate Case-by-Case)

| Filter                  | Guidance                                                                                              |
| ----------------------- | ----------------------------------------------------------------------------------------------------- |
| Sector concentration    | Check against `config/watchlist.json` sectors. If adding would push any sector above `max_sector_concentration_pct` in settings.json, flag but do not auto-reject. |
| Thesis linkage strength | Must demonstrate direct revenue exposure to thesis driver, not tangential. "Benefits from AI" is insufficient; specify which revenue segment and what percentage. |
| Existing coverage       | Prefer companies with 5+ sell-side analysts covering. Low coverage means higher `unknown` rate in condition evaluation. |
| Profitability           | Prefer profitable companies (positive trailing 12M operating income). Pre-profit companies acceptable only for High conviction theses with 24M+ horizon. |

### Thesis Linkage Test

A candidate passes the linkage test if it satisfies ALL of:

1. **Revenue specificity**: At least one identifiable business segment derives 20%+ revenue from the thesis driver.
2. **Causal mechanism**: A concrete, articulable chain from thesis condition to company revenue exists (not just sector co-movement).
3. **Differentiation**: The company has a specific competitive advantage in the thesis area beyond "participates in the industry."

Candidates that fail the linkage test are excluded even if they pass all hard filters.

## 5. Chain Dependency Rules

### Dependency Declaration

Theses in `config/theses.md` may depend on other theses. Example: "Meta Platform AI Monetization" depends partly on "AI Infra Expansion" (Meta needs functioning AI infrastructure to monetize it).

Dependencies are directional: upstream thesis affects downstream, not vice versa.

### Propagation Rules

| Upstream Status | Effect on Downstream                                                          |
| --------------- | ----------------------------------------------------------------------------- |
| `valid`         | No constraint. Evaluate downstream independently.                             |
| `weakening`     | Downstream thesis status capped at `weakening`. Cannot be `valid`.            |
| `invalidated`   | Downstream thesis status capped at `weakening`. Flag for accelerated review.  |
| `insufficient`  | Downstream evaluation proceeds but note upstream data gap in assessment.      |

### Why Not Auto-Invalidate Downstream?

Upstream `invalidated` caps downstream at `weakening` rather than auto-invalidating because:

- Downstream theses may have independent validity conditions that remain intact.
- The dependency relationship may weaken over time (e.g., Meta builds its own chips, reducing NVDA dependency).
- Auto-invalidation cascades amplify single-point-of-failure risk in thesis evaluation.

If upstream remains `invalidated` for 2 consecutive review cycles, escalate downstream to `invalidated` unless the researcher documents a specific reason the dependency has weakened.

### Evaluation Order

When running a full research cycle, evaluate upstream theses first. Use upstream results when assessing downstream theses. This prevents circular evaluation.

## 6. Time Horizon Awareness

### Horizon Categories

| Category    | Thesis Horizon | Review Cadence | "Too Early" Tolerance       |
| ----------- | -------------- | -------------- | --------------------------- |
| Short-term  | < 12 months    | Monthly        | 1 quarter of `not_yet`      |
| Medium-term | 12-36 months   | Quarterly      | 2 quarters of `not_yet`     |
| Long-term   | 36+ months     | Semi-annually  | 4 quarters of `not_yet`     |

### Timeline Expiry Rules

A thesis is flagged as "timeline expiring" when:

- **Short-term**: Less than 3 months remain and majority of V conditions are `not_yet`.
- **Medium-term**: Less than 6 months remain and majority of V conditions are `not_yet` or `unknown`.
- **Long-term**: No expiry flag. Instead, conduct a structural re-evaluation every 12 months asking: "Has the thesis timeframe itself shifted?"

### "Too Early to Evaluate" Handling

For conditions that cannot yet be measured (e.g., a product not yet launched, a regulation not yet enacted):

1. Mark the condition as `not_yet` with a note: "evaluation deferred — [specific trigger event] required."
2. Do not let deferred conditions count against thesis status within the tolerance window above.
3. After the tolerance window expires, `not_yet` conditions with no progress shift the thesis toward `weakening`.

### Horizon Mismatch Warning

If a short-term thesis (< 12 months) has conditions that structurally require 24+ months to resolve, flag the mismatch. The thesis horizon or the condition needs revision — this is a formulation error, not a research finding.

## 7. Status Transition Tracking

### Purpose

Tracking how condition and thesis statuses change across research cycles reveals trends that a single snapshot cannot. A condition that has been `met` for 3 consecutive cycles is stronger evidence than one that just flipped to `met`. Conversely, `met→partially_met→not_yet` across 3 cycles signals structural deterioration.

### Condition-Level Transition Classification

The status space has a natural ordering for transition classification:

```
refuted < not_yet < unknown < partially_met < met
```

| Transition | Definition | Example |
| --- | --- | --- |
| `new` | No previous_status exists (first run or condition added to theses.md) | — |
| `stable` | Current status == previous status | met → met |
| `improving` | Current status is higher in ordering than previous | not_yet → partially_met, partially_met → met |
| `degrading` | Current status is lower in ordering than previous | met → partially_met, partially_met → refuted |

**Special cases**:
- `unknown → met`: classified as `improving` (gained information)
- `unknown → refuted`: classified as `degrading` (gained adverse information)
- `met → unknown`: classified as `degrading` (lost confirming evidence)
- `unknown → unknown`: classified as `stable`

### Thesis-Level Transition Classification

Same ordering applies at thesis level:

```
invalidated < weakening < valid
```

| Transition | Example |
| --- | --- |
| `new` | First research cycle for this thesis |
| `stable` | valid → valid |
| `improving` | weakening → valid |
| `degrading` | valid → weakening, weakening → invalidated |

### Previous Source Reuse Rules

When `previous_check` contains sources for a condition:

1. **Current sources** (date within 90 days of today): May be carried forward into the new record. Still counts toward the source count requirements.
2. **Aging sources** (91-180 days): Carry forward only as supplementary context. Do NOT use as the sole basis for any status assessment.
3. **Stale sources** (>180 days): Drop entirely. These are preserved in `data/research-history/` for historical reference.

Carrying forward a source does NOT exempt the researcher from performing new searches. Every condition still requires minimum 1 confirming + 1 disconfirming search to detect changes.

### Consecutive Degradation Warning

If a condition shows `degrading` for 2 consecutive research cycles (checkable via `data/research-history/`), flag it in the summary as "연속 악화 — 논제 재검토 필요". This signal is stronger than a single-cycle degradation and warrants thesis-level attention.

### Archive Structure

```
data/research-history/
  thesis-check-2026-01-15.json
  thesis-check-2026-02-16.json
  ...
```

Files are named by the `checked_at` date of the archived research. The archive is append-only — never modify or delete historical files. If a same-date archive already exists (re-run on same day), overwrite it.

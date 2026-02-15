# Cognitive Bias Guardrails

## Bias Catalog

| Bias                      | Detection Signal                                                                                   | Mitigation Action                                                                              |
| ------------------------- | -------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| 최신편향 (Recency)        | 5D change cited with higher decisional weight than 52W trend                                       | When citing 5D data, MUST co-cite 52W range context in the same paragraph                      |
| 서사편향 (Narrative)      | Thesis section has zero counter-arguments or all evidence points in the same direction             | Each thesis MUST include at least one data-backed counter-argument                             |
| 확증편향 (Confirmation)   | All evidence categories are price-direction-correlated (trend + momentum + relative strength only) | High confidence requires 1+ evidence from volatility or external category (see methodology.md) |
| 손실회피 (Loss Aversion)  | Drawdown exceeds `max_drawdown_warning_pct` while thesis conditions remain valid                   | Flag "논제 건전 — 손실회피 편향 점검" in asset note; do NOT auto-reduce allocation             |
| 과잉확신 (Overconfidence) | More than 60% of theses rated High confidence                                                      | Apply High upgrade rules in methodology.md; downgrade excess to Medium                         |
| 처분효과 (Disposition)    | Reducing allocation for thesis-valid gainers OR increasing for thesis-weakening losers             | Any allocation change MUST cite a specific thesis status change as rationale                   |
| 앵커링 (Anchoring)        | Decision rationale references purchase price or a specific past price level                        | Use 52W range and MA levels as reference points; never reference absolute price targets        |

## Bias Audit Output

Output this table at the END of Phase 2, BEFORE proceeding to Phase 3. All 7 rows required.

| Bias                      | Status          | Evidence                   |
| ------------------------- | --------------- | -------------------------- |
| 최신편향 (Recency)        | clear / flagged | _1-sentence justification_ |
| 서사편향 (Narrative)      | clear / flagged | _1-sentence justification_ |
| 확증편향 (Confirmation)   | clear / flagged | _1-sentence justification_ |
| 손실회피 (Loss Aversion)  | clear / flagged | _1-sentence justification_ |
| 과잉확신 (Overconfidence) | clear / flagged | _1-sentence justification_ |
| 처분효과 (Disposition)    | clear / flagged | _1-sentence justification_ |
| 앵커링 (Anchoring)        | clear / flagged | _1-sentence justification_ |

Status MUST be exactly `clear` or `flagged`. If flagged, evidence MUST state which asset triggered it.

## Self-Check Questions

Answer each before finalizing Phase 2. If any answer is **No**, revise before proceeding.

- Did I cite 52W data whenever I cited 5D data? → If No, add 52W context.
- Does every thesis contain at least one counter-argument with supporting data? → If No, add counter-arguments.
- For every High confidence rating, did I use evidence from 2+ categories including volatility or external? → If No, downgrade to Medium.
- Is every allocation change justified by a thesis status change, not price movement alone? → If No, revise rationale.
- Are fewer than 60% of theses rated High? → If No, re-evaluate against methodology.md rules.

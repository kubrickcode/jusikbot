---
name: portfolio-analyze
description: 논제 기반 포트폴리오 분석. 시장 데이터 요약과 투자 논제를 대조하여 확신도별 배분 금액을 제안하고, 인지 편향 가드레일과 직전 배분 앵커링을 적용한다.
allowed-tools: Bash(python *), Bash(.claude/skills/portfolio-analyze/scripts/*), Bash(psql *), Bash(ls *), Read, Write, AskUserQuestion
disable-model-invocation: true
---

# Portfolio Analysis Workflow

All output MUST be in **Korean**. Instructions below are in English for accuracy.

## Hard Constraints

These rules apply to ALL phases. Violation causes report rejection.

**Data Sources — Whitelist (MUST cite ONLY from these)**:

| Tag       | Source                         | Contains                       |
| --------- | ------------------------------ | ------------------------------ |
| [summary] | `data/summary.md`              | 14-column technical indicators |
| [psql]    | PostgreSQL via `$DATABASE_URL` | price_history, fx_rate tables  |
| [user]    | `$ARGUMENTS`                   | User-provided context          |
| [config]  | `config/settings.json`         | Budget, ratios, limits         |

- You MUST NOT cite, infer, or reference any data not from the above sources.
- Every numeric claim in the report MUST carry a source tag: `[summary]`, `[psql]`, `[user]`, or `[config]`.

**Confidence**: Default is **Medium**. Upgrade/downgrade rules in `references/methodology.md`.

**Forbidden Patterns**:

- NEVER use certainty language: "확실히", "반드시", "무조건", "틀림없이", "100%"
- NEVER provide price targets or fair value estimates
- NEVER cite external data sources (news, earnings reports, analyst ratings) unless provided in `$ARGUMENTS`

**KR Data Warning**: When mentioning 069500 or KODEX 200, MUST include a warning that KIS adj_close does not reflect dividends/distributions, so MA divergence and 52W Position may be overstated.

**Anchoring**: When a previous report exists, allocation changes MUST respect `config/settings.json` `anchoring.*` limits for the current review type.

---

## Phase 1: Load + Validate (deterministic)

**Round 1 — Parallel Read**:

| #   | Action                                                                |
| --- | --------------------------------------------------------------------- |
| 1   | Read `config/settings.json`                                           |
| 2   | Read `config/watchlist.json`                                          |
| 3   | Read `config/theses.md`                                               |
| 4   | Read `data/summary.md`                                                |
| 5   | Read `.claude/skills/portfolio-analyze/references/methodology.md`     |
| 6   | Read `.claude/skills/portfolio-analyze/references/bias-guardrails.md` |
| 7   | Bash: `ls output/reports/`                                            |

**Round 2 — Latest Report + Freshness**:

- If `ls` found report files: Read the most recent `output/reports/YYYY-MM-DD.md` and parse its allocation table into `previous_allocations` JSON.
- Bash: `.claude/skills/portfolio-analyze/scripts/validate-freshness.sh data/summary.md output/reports`
- If status is `STALE`: AskUserQuestion — "summary.md is N days old. Proceed with stale data or run `just collect` first?"
- Store `review_type` (monthly/quarterly) from freshness output.

**Overtrading Check** (deterministic):

- If `latest_report` date is < `min_review_interval_days` (from settings.json) AND `config/theses.md` content is unchanged from the previous report's thesis snapshot: produce a brief "현 배분 유지" report with rationale. Skip Phase 2-3.
- "Thesis change" is defined as: (1) `config/theses.md` was modified since the latest report date, OR (2) `$ARGUMENTS` explicitly requests a review.

---

## Phase 2: Analyze (LLM)

**Input**: $ARGUMENTS (user request/context), loaded files from Phase 1. If `$ARGUMENTS` is empty, analyze all theses with current data — no special user context to address.

**Conditional Load**: If `data/summary.md` is missing a watchlist symbol or a thesis evaluation requires data beyond the 14-column summary (e.g., exact 52W high date, intra-period volume spike), Read `.claude/skills/portfolio-analyze/references/query-templates.md` and execute relevant psql queries.

**For each thesis in `config/theses.md`**:

1. Evaluate validity conditions against current data (cite sources)
2. Evaluate invalidation conditions
3. Assign confidence: Medium (default), upgrade to High or downgrade to Low per `references/methodology.md` rules
4. Overextension qualifier: if confidence = High AND `vs 200D` > +20% [summary] → append "진입 타이밍 주의"
5. KR data quality: if symbol has KIS adj_close warning in summary.md, MUST NOT upgrade to High based on technical indicators alone
6. Include at least one counter-argument with data

**Inter-thesis conflict detection**: Flag when two theses imply opposing positions.

**Loss aversion check**: If drawdown from 52W High > `risk_tolerance.max_drawdown_warning_pct` (settings.json) but thesis conditions still valid → flag "논제 건전 — 손실회피 편향 점검".

**Bias Audit**: Complete the 7-row audit table from `references/bias-guardrails.md`. All rows MUST show `clear` or `flagged` with 1-sentence evidence.

**Self-Check**: Answer all questions from `references/bias-guardrails.md` Self-Check section. Revise analysis if any answer is No.

---

## Phase 3: Allocate + Validate (LLM + deterministic)

**Generate allocation JSON** with this exact structure:

```json
{
  "SYMBOL": {"amount": N, "role": "core|satellite", "confidence": "high|medium|low"},
  ...
}
```

- Use symbol names exactly as in `config/watchlist.json`
- Follow sizing rules in `references/methodology.md` Section 4
- Round all amounts to `adjustment_unit_krw` multiples
- Total MUST equal `budget_krw`

**Validate**: Run:

```bash
python .claude/skills/portfolio-analyze/scripts/validate_allocation.py \
  --settings config/settings.json \
  --watchlist config/watchlist.json \
  --allocations '<ALLOCATION_JSON>' \
  [--previous-allocations '<PREV_JSON>'] \
  --review-type <monthly|quarterly>
```

- If PASS: proceed to Phase 4.
- If FAIL: read the `errors` array. Fix the violated items first, then adjust related positions to maintain budget_total and other constraints. Do NOT regenerate all allocations from scratch — preserve unaffected positions. If a fix creates a new violation, follow the constraint priority in `references/methodology.md`. Retry validation. Maximum 2 retries.
- If 3 failures AND previous report exists: use `previous_allocations` from Phase 1 as fallback. Add warning: "검증 3회 실패 — 직전 배분 유지".
- If 3 failures AND no previous report (first run): generate a conservative equal-weight allocation using only core positions at Medium confidence, then re-validate once.

---

## Phase 4: Report + Validate (LLM + deterministic)

**Write** `output/reports/YYYY-MM-DD.md` (today's date) with these H2 sections:

- `## 요약` — 1-paragraph executive summary
- `## 논제별 현황` — H3 per thesis, each with: status, confidence, evidence, and counter-argument under `**반론/리스크**:` heading
- `## 배분 제안` — allocation table + change rationale per position
- `## 리스크 요인` — portfolio-level risks, drawdown status, bias audit summary

**Validate**: Run:

```bash
python .claude/skills/portfolio-analyze/scripts/validate_report.py output/reports/YYYY-MM-DD.md
```

- If FAIL: fix violated rules and rewrite. Maximum 2 retries.

**Self-Check (2 items)**:

1. Did every allocation change cite a specific rationale (thesis change, rebalancing, or data-driven)?
2. Did the report address the user's `$ARGUMENTS` request? (Skip if `$ARGUMENTS` was empty.)

If any applicable answer is No, revise the report before completing.

---
name: thesis-research
description: 투자 논제의 유효/무효화 조건을 웹 리서치로 팩트체크하고, 워치리스트 후보 종목을 탐색한다. /portfolio-analyze 실행 전에 사용하여 data/thesis-check.json과 data/candidates.json을 생성한다. 논제 조건의 실제 충족 여부를 확인하고 싶을 때, 새로운 투자 종목을 발굴하고 싶을 때, 포트폴리오 분석 전 최신 팩트체크가 필요할 때 사용한다.
allowed-tools: WebFetch, WebSearch, Read, Write, Bash(python .claude/skills/thesis-research/scripts/*), Bash(ls *), Bash(mkdir *), Bash(cp *), AskUserQuestion
---

# Thesis Research Workflow

All output MUST be in **Korean**. Instructions below are in English for accuracy.

## Hard Constraints

These rules apply to ALL phases. Violation causes output rejection.

**Anti-Confirmation Bias — Mandatory Disconfirmation Search**:

For every condition evaluated, you MUST perform at least one web search specifically formulated to find COUNTER-EVIDENCE. If you search "NVDA revenue growth 2025", you MUST also search "NVDA revenue slowdown risk" or "GPU demand decline." Skipping the disconfirmation search is a hard failure.

**Source Quality Hierarchy**:

| Tier | Source Type                                      | Trust Level | Examples                                    |
| ---- | ------------------------------------------------ | ----------- | ------------------------------------------- |
| T1   | Primary — official filings, government data, IRs | Highest     | 10-K/10-Q, earnings transcripts, DART, FRED |
| T2   | Industry — research firms, institutional reports | High        | IDC, Gartner, Bloomberg, Reuters            |
| T3   | News — major financial media                     | Medium      | WSJ, CNBC, FT, 한경, 매경                   |
| T4   | Opinion — blogs, social media, analysts          | Low         | Seeking Alpha, X threads, YouTube           |

- Every evidence item MUST carry a tier tag.
- Conditions assessed solely on T4 sources CANNOT be marked `met` or `refuted` — cap at `partially_met` or `unknown`.

**Recency Requirement**:

- Evidence ≤90 days old: `current`
- Evidence 91–180 days old: `aging` — usable but flag in output
- Evidence >180 days old: `stale` — historical context only, NEVER basis for status assessment

**No Hallucination — Zero Tolerance**:

- If web search returns no useful results, mark status as `unknown`.
- NEVER fabricate URLs, dates, numbers, or quotes.
- NEVER infer specific numbers from general trends.

**Forbidden Patterns**:

- NEVER use certainty language: "확실히", "반드시", "무조건", "틀림없이", "100%"
- NEVER provide price targets or fair value estimates
- NEVER state investment recommendations (buy/sell/hold)

---

## Phase 1: Load + Parse

**Read these files in parallel**:

| #   | Action                                    |
| --- | ----------------------------------------- |
| 1   | Read `config/theses.md`                   |
| 2   | Read `config/watchlist.json`              |
| 3   | Read `config/settings.json`               |
| 4   | Read `data/thesis-check.json` (if exists) |

**Load previous research** (if `data/thesis-check.json` exists):

Store as `previous_check`. This provides:

- Per-condition `previous_status` for transition tracking
- Previous T1/T2 sources still within recency window — reusable without re-searching
- Previous thesis-level status for trend detection

If `data/thesis-check.json` does not exist, this is a first run — set `previous_check = null`.

**Parse theses into structured checklist**:

For each thesis in `config/theses.md`, extract: thesis name, role, horizon, validity conditions, invalidation conditions, related symbols. Build a flat research checklist — every individual condition becomes a research task.

If total conditions > 30: AskUserQuestion — "조건 {N}개 확인 필요. 전체 진행 or 특정 논제 선택?"

---

## Phase 2: Per-Thesis Condition Research

Process theses in document order. For each thesis, research ALL its conditions before moving to the next.

Detailed methodology: See `references/research-methodology.md` for condition evaluation framework, source quality rules, chain dependency propagation, and status transition classification.

### 2.0 Previous Result Lookup

Before researching each condition, check `previous_check` for the matching condition (match by `text` field):

- **If found**: note `previous_status`. Previous T1/T2 sources with `date` still within 90 days (current) may be carried forward — cite them alongside new evidence instead of re-searching. Still perform at least 1 new confirming + 1 new disconfirming search to check for changes.
- **If not found**: this is a new condition (`status_transition = "new"`).

### 2.1 Per-Condition Research Protocol

For each condition:

**Step A — Formulate Search Queries** (minimum 2 per condition):

- At least 1 CONFIRMING query
- At least 1 DISCONFIRMING query
- Use specific terms: company names, metric names, time periods
- English for US/global data, Korean for Korean market data
- If previous research exists with `current` sources, focus new queries on CHANGES since the previous check date

**Step B — Execute Searches**:

- WebSearch for each query
- WebFetch on most promising results (prioritize T1/T2 sources)
- Cap at 5 fetches per condition
- If initial searches yield nothing useful, try 1 reformulated query before marking `unknown`

**Step C — Assess Source Quality**:

Per evidence item: assign tier (T1–T4), check recency (current/aging/stale), extract specific data point.

**Step D — Classify Condition Status**:

| Status          | Criteria                                                       |
| --------------- | -------------------------------------------------------------- |
| `met`           | Clear T1/T2 evidence supports condition. Current data.         |
| `partially_met` | Some evidence supports but incomplete, or only T3/T4 sources.  |
| `not_yet`       | Condition not yet testable (e.g., next earnings not released). |
| `refuted`       | Clear T1/T2 evidence contradicts condition. Current data.      |
| `unknown`       | No useful evidence found after good-faith search.              |

**Step E — Quantitative Distance** (when applicable):

If the condition involves a numeric threshold, compute distance:
`"quantitative_distance": "+18pp margin"` (actual 48% vs threshold 30%)

If actual value unavailable, omit entirely. Do NOT estimate.

**Step F — Record Evidence** per condition:

```json
{
  "text": "condition text from theses.md",
  "type": "validity|invalidation",
  "status": "met|partially_met|not_yet|refuted|unknown",
  "previous_status": "met|null",
  "status_transition": "stable|improving|degrading|new|null",
  "evidence": "1-2 sentence Korean summary",
  "sources": [
    {"title": "...", "url": "https://...", "tier": 1, "date": "YYYY-MM-DD"}
  ],
  "quantitative_distance": "+18pp margin" | null
}
```

**Status transition classification** (see `references/research-methodology.md` Section 7):

| Transition  | Rule                                                                          |
| ----------- | ----------------------------------------------------------------------------- |
| `new`       | No previous_status (first run or new condition)                               |
| `stable`    | status == previous_status                                                     |
| `improving` | Status moved toward `met` (e.g. not_yet→partially_met, partially_met→met)     |
| `degrading` | Status moved toward `refuted` (e.g. met→partially_met, partially_met→refuted) |

````

### 2.2 Thesis-Level Assessment

After all conditions for a thesis are researched:

| Thesis Status | Rule                                                                              |
| ------------- | --------------------------------------------------------------------------------- |
| `valid`       | Majority V conditions `met`/`partially_met` AND zero I conditions `met`/`refuted` |
| `weakening`   | Any I condition `partially_met` OR majority V conditions `not_yet`/`unknown`      |
| `invalidated` | Any I condition `met`/`refuted` OR majority V conditions `refuted`                |

Asymmetric weighting: a single invalidation condition `refuted` triggers `invalidated` regardless of validity status. See `references/research-methodology.md` Section 3.

After determining thesis status, compute thesis-level transition from `previous_check`:
- `previous_status`: thesis status from previous cycle (null if first run)
- `status_transition`: same classification as condition-level (stable/improving/degrading/new)

### 2.3 Cross-Thesis Dependency Check

After ALL theses are assessed:

- If upstream thesis weakens → downstream capped at `weakening`
- If upstream thesis invalidated → downstream capped at `weakening`, flag for accelerated review
- Record as `upstream_dependency` and `chain_impact` fields in output

See `references/research-methodology.md` Section 5 for full propagation rules.

---

## Phase 3: Candidate Company Discovery

During Phase 2 research, collect companies relevant to each thesis that are NOT in `config/watchlist.json`. Do NOT run separate discovery searches unless fewer than 2 candidates found across all theses.

### 3.1 Hard Filters

| Filter           | Threshold                                                        |
| ---------------- | ---------------------------------------------------------------- |
| Market cap       | ≥ $5B USD (or KRW 6.5T)                                          |
| Daily volume     | ≥ $10M USD (20-day average)                                      |
| Market access    | US (NYSE/NASDAQ) or KR (KOSPI/KOSDAQ). EU only via US-listed ADR |
| Listing history  | ≥ 2 years                                                        |
| Not in watchlist | Check `config/watchlist.json`                                    |

### 3.2 Thesis Linkage Test (ALL must pass)

1. **Revenue specificity**: ≥20% revenue from thesis driver
2. **Causal mechanism**: Concrete chain from thesis condition to company revenue
3. **Differentiation**: Specific competitive advantage beyond sector participation

See `references/research-methodology.md` Section 4 for soft filters and detailed criteria.

### 3.3 Per-Candidate Record

```json
{
  "symbol": "TICKER",
  "name": "Company Name",
  "market": "US|KR",
  "sector": "sector-slug",
  "type": "stock|etf",
  "related_theses": ["thesis name"],
  "rationale": "Korean: why this fits",
  "risks": "Korean: primary risk",
  "market_cap_category": "large|mid|small",
  "already_in_watchlist": false
}
````

### 3.4 Limits

- Maximum 10 candidates total, 5 per thesis
- Priority: T1/T2 source mentions > direct thesis alignment > market cap

---

## Phase 4: Output + Validate

### 4.0 Archive Previous Research

If `data/thesis-check.json` exists:

```bash
mkdir -p data/research-history
cp data/thesis-check.json data/research-history/thesis-check-$(cat data/thesis-check.json | python -c "import sys,json; print(json.load(sys.stdin)['checked_at'])").json
```

This preserves the full history. The `data/research-history/` directory accumulates one file per research cycle.

### 4.1 Write `data/thesis-check.json`

```json
{
  "checked_at": "YYYY-MM-DD",
  "theses": [
    {
      "name": "thesis name from theses.md",
      "status": "valid|weakening|invalidated",
      "previous_status": "valid|null",
      "status_transition": "stable|improving|degrading|new|null",
      "upstream_dependency": null,
      "chain_impact": null,
      "conditions": [
        /* Step F records (including previous_status, status_transition) */
      ]
    }
  ]
}
```

Schema: `data/thesis-check.schema.json`

### 4.2 Write `data/candidates.json`

```json
{
  "checked_at": "YYYY-MM-DD",
  "candidates": [
    /* Per-Candidate Records */
  ]
}
```

If no candidates found: `{"checked_at": "...", "candidates": []}`.
Schema: `data/candidates.schema.json`

### 4.3 Validate

```bash
python .claude/skills/thesis-research/scripts/validate_research.py \
  --thesis-check data/thesis-check.json \
  --candidates data/candidates.json \
  --theses config/theses.md \
  --watchlist config/watchlist.json
```

- PASS → proceed to summary
- FAIL → fix issues, rewrite, re-validate. Maximum 2 retries.

### 4.4 Present Summary

```markdown
## 논제 리서치 결과 요약

| 논제      | 상태      | 전이              | 유효 조건 | 무효화 조건     | 주요 발견          |
| --------- | --------- | ----------------- | --------- | --------------- | ------------------ |
| AI 인프라 | weakening | valid→weakening ↓ | 2/3 met   | 1 partially_met | GPU 과잉 공급 신호 |

### 상태 전이 요약

- ↑ improving: [list]
- → stable: [list]
- ↓ degrading: [list]
- 🆕 new: [list]

### 신규 후보 종목

- SYMBOL (Name): rationale (관련 논제)

### 주의 사항

- [unknown 조건, stale 증거, cross-thesis impact, degrading trends 등]
```

---

## Error Recovery

| Situation                 | Action                                                                     |
| ------------------------- | -------------------------------------------------------------------------- |
| WebSearch zero results    | Reformulate query once. If still nothing, mark `unknown`.                  |
| WebFetch fails (403, etc) | Try alternate URL. If all fail, use search snippets + `fetch_failed` note. |
| Ambiguous evidence        | Mark `partially_met`, include both interpretations.                        |
| Contradictory sources     | Include both. Higher-tier source determines status. Note conflict.         |
| Future data required      | Mark `not_yet` with explanation (e.g., "다음 분기 실적 미발표").           |

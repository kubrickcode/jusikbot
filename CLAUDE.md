# jusikbot

Personal investment portfolio management tool. Collects market data, stores in PostgreSQL, and provides thesis-based portfolio analysis via Claude Code skill.

## Architecture

```
Go collector → PostgreSQL → Claude Code skills
     ↓              ↓              ↓
Tiingo/KIS/     price_history   /thesis-research (web search → fact-check)
Frankfurter     fx_rate              ↓
                                /portfolio-analyze (price data + research → allocation)
```

## Key Commands

| Command               | Description                             |
| --------------------- | --------------------------------------- |
| `just collect`        | Run Go collector (all sources)          |
| `just collect us`     | Collect US stocks only                  |
| `just collect kr`     | Collect KR stocks only                  |
| `just collect fx`     | Collect FX rates only                   |
| `just migrate`        | Run DB migrations                       |
| `just test-collector` | Run Go tests                            |
| `just lint`           | Format config + justfile                |
| `/thesis-research`    | Fact-check theses + discover candidates |
| `/portfolio-analyze`  | Run portfolio analysis skill            |

## File Structure

```
config/
  settings.json          # Budget, risk limits, sizing, anchoring
  watchlist.json         # 5 symbols with market/type/sector/themes
  theses.md              # Investment theses with validity conditions
collector/               # Go module — data collection
  cmd/collect/           # CLI entry point
  cmd/migrate/           # DB migration
  internal/              # collector, store, indicator packages
data/
  summary.md             # 14-column technical indicator summary (auto-generated)
  thesis-check.json      # Thesis condition fact-check results (from /thesis-research)
  candidates.json        # Candidate companies discovered (from /thesis-research)
  research-history/      # Archived thesis-check snapshots (thesis-check-YYYY-MM-DD.json)
output/reports/          # Portfolio analysis reports (YYYY-MM-DD.md)
.claude/skills/thesis-research/
  SKILL.md               # Thesis research workflow (web search enabled)
  scripts/               # validate_research.py
  references/            # research-methodology.md
.claude/skills/portfolio-analyze/
  SKILL.md               # Portfolio analysis workflow (price data + research)
  scripts/               # Validation scripts (Python/Bash)
  references/            # methodology, bias-guardrails, query-templates
docs/decisions/          # Architecture Decision Records
```

## Data Sources

| Source      | Data                       | API Key Env                     |
| ----------- | -------------------------- | ------------------------------- |
| Tiingo      | US stock OHLCV + adj_close | `TIINGO_TOKEN`                  |
| KIS OpenAPI | KR stock OHLCV             | `KIS_APP_KEY`, `KIS_APP_SECRET` |
| Frankfurter | FX rates (USD/KRW)         | None (public)                   |

## Database

- **Connection**: `$DATABASE_URL` (PostgreSQL)
- **Tables**: `price_history` (symbol+date PK), `fx_rate` (pair+date PK)
- **Query**: `psql "$DATABASE_URL" -c "SELECT ..."`

## Rules

- Config files use exact versions, no ranges
- Python scripts: stdlib only, no external dependencies
- Go: standard project layout, tests alongside source
- All settings in `config/settings.json` — single source of truth for numeric parameters
- Reports validated by deterministic scripts before completion

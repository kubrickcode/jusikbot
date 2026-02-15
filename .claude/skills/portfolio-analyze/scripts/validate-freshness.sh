#!/usr/bin/env bash
set -euo pipefail

# Side effect: reads summary.md and scans output/reports/ to determine data freshness and review cycle type.
# Exit 0 = fresh, Exit 1 = stale or error

SUMMARY_PATH="${1:-}"
REPORTS_DIR="${2:-output/reports}"

if [[ -z "$SUMMARY_PATH" ]]; then
  echo '{"status":"ERROR","detail":"Usage: validate-freshness.sh <summary.md> [reports-dir]"}' >&2
  exit 1
fi

if [[ ! -f "$SUMMARY_PATH" ]]; then
  echo '{"status":"ERROR","detail":"File not found at provided path"}' >&2
  exit 1
fi

generated_line=$(grep -m1 '^> Generated:' "$SUMMARY_PATH" || true)
if [[ -z "$generated_line" ]]; then
  echo '{"status":"ERROR","detail":"No Generated timestamp found in summary"}' >&2
  exit 1
fi

summary_date=$(echo "$generated_line" | sed -E 's/^> Generated: ([0-9]{4}-[0-9]{2}-[0-9]{2}).*/\1/')
if [[ -z "$summary_date" ]] || ! date -d "$summary_date" +%s >/dev/null 2>&1; then
  echo '{"status":"ERROR","detail":"Invalid date format in summary"}' >&2
  exit 1
fi

today=$(date -u +%Y-%m-%d)
summary_epoch=$(date -d "$summary_date" +%s)
today_epoch=$(date -d "$today" +%s)
days_old=$(( (today_epoch - summary_epoch) / 86400 ))

latest_report=""
latest_report_date=""
review_type="quarterly"

if [[ -d "$REPORTS_DIR" ]]; then
  latest_report=$(find "$REPORTS_DIR" -maxdepth 1 -name '*.md' -not -name '.gitkeep' | sort -r | head -1 || true)

  if [[ -n "$latest_report" ]]; then
    latest_report_date=$(basename "$latest_report" | sed -E 's/^([0-9]{4}-[0-9]{2}-[0-9]{2}).*/\1/')

    if date -d "$latest_report_date" +%s >/dev/null 2>&1; then
      report_epoch=$(date -d "$latest_report_date" +%s)
      days_since_report=$(( (today_epoch - report_epoch) / 86400 ))

      if [[ $days_since_report -lt 60 ]]; then
        review_type="monthly"
      fi
    fi
  fi
fi

latest_report_json="null"
if [[ -n "$latest_report_date" ]]; then
  latest_report_json="\"$latest_report_date\""
fi

if [[ "$summary_date" != "$today" ]]; then
  cat <<EOF
{"status":"STALE","summary_date":"$summary_date","days_old":$days_old,"review_type":"$review_type","latest_report":$latest_report_json}
EOF
  exit 0
fi

cat <<EOF
{"status":"OK","summary_date":"$summary_date","review_type":"$review_type","latest_report":$latest_report_json}
EOF
exit 0

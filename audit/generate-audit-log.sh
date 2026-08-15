#!/usr/bin/env bash
# Generates a real, git-derived audit log — run this from the repo root.
# Usage: bash audit/generate-audit-log.sh
# Output: audit/audit-log-export.csv (commit hash, author, date, task ID, message)

set -euo pipefail

OUT="audit/audit-log-export.csv"
echo "CommitHash,Author,Email,Date,TaskID,Type,Message" > "$OUT"

git log --pretty=format:'%h|%an|%ae|%ad|%s' --date=iso | while IFS='|' read -r hash author email date subject || [ -n "$hash" ]; do
  # Extract task ID if referenced anywhere in the subject, e.g. T06, T13
  task_id=$(echo "$subject" | grep -oE 'T[0-9]{2}' | head -1 || true)
  type=$(echo "$subject" | grep -oE '^[a-z]+' || true)
  # Escape double quotes in subject for CSV safety
  safe_subject=$(echo "$subject" | sed 's/"/""/g')
  echo "$hash,\"$author\",\"$email\",\"$date\",\"${task_id:-N/A}\",\"$type\",\"$safe_subject\"" >> "$OUT"
done

echo "Wrote $OUT"
echo ""
echo "--- Commit count per author (contribution balance check) ---"
git --no-pager log | git --no-pager shortlog -sne

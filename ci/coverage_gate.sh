#!/usr/bin/env bash
set -euo pipefail

threshold="${COVERAGE_THRESHOLD:-60}"
profile="${1:-coverage.out}"

if [[ ! -f "$profile" ]]; then
  echo "coverage profile not found: $profile" >&2
  exit 1
fi

total="$(go tool cover -func="$profile" | awk '/^total:/ {gsub(/%/, "", $3); print $3}')"
if [[ -z "$total" ]]; then
  echo "failed to read total coverage" >&2
  exit 1
fi

awk -v got="$total" -v min="$threshold" 'BEGIN {
  if (got + 0 < min + 0) {
    printf("coverage gate failed: got %.2f%%, need %.2f%%\n", got, min)
    exit 1
  }
  printf("coverage gate passed: got %.2f%% (min %.2f%%)\n", got, min)
}'

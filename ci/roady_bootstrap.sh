#!/usr/bin/env bash
set -euo pipefail

docs_dir="${1:-docs}"
analyze_flag="${2:-}"

echo "Running Roady preflight..."
bash ci/roady_preflight.sh

echo "Syncing docs into spec/plan..."
bash ci/roady_docs_sync.sh "${docs_dir}" "${analyze_flag}"

echo "Final Roady health check..."
roady status
roady drift detect

echo "Roady bootstrap completed."

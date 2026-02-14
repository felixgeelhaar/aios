#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  ci/roady_docs_sync.sh [docs-dir] [--analyze]

Default:
  docs-dir = docs

Examples:
  ci/roady_docs_sync.sh
  ci/roady_docs_sync.sh docs
  ci/roady_docs_sync.sh docs --analyze
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

docs_dir="${1:-docs}"
analyze="${2:-}"
if [[ ! -d "${docs_dir}" ]]; then
  echo "docs directory not found: ${docs_dir}" >&2
  exit 1
fi

if [[ "${analyze}" == "--analyze" ]]; then
  echo "Analyzing docs from ${docs_dir} with reconcile..."
  roady spec analyze "${docs_dir}" --reconcile
else
  echo "Skipping docs analyze. Pass --analyze to run spec inference."
fi

echo "Generating and approving plan..."
roady plan generate
roady plan approve

echo "Current project status..."
roady status
roady drift detect

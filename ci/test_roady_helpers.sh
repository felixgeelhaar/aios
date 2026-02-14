#!/usr/bin/env bash
set -euo pipefail

scripts=(
  "ci/roady_task.sh"
  "ci/roady_docs_sync.sh"
)

for script in "${scripts[@]}"; do
  if [[ ! -f "${script}" ]]; then
    echo "missing script: ${script}" >&2
    exit 1
  fi
  bash -n "${script}"
done

bash ci/roady_task.sh status >/dev/null
bash ci/roady_docs_sync.sh --help >/dev/null

echo "Roady helper smoke tests passed."

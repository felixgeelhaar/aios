#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  ci/roady_task.sh ready
  ci/roady_task.sh status
  ci/roady_task.sh start <task-id>
  ci/roady_task.sh complete <task-id> <evidence>
  ci/roady_task.sh verify <task-id> <evidence>
  ci/roady_task.sh cycle <task-id> <evidence>

Examples:
  ci/roady_task.sh start task-cli-export-status-report
  ci/roady_task.sh complete task-cli-export-status-report "implemented + tests pass"
  ci/roady_task.sh verify task-cli-export-status-report "go test ./... pass"
  ci/roady_task.sh cycle task-cli-export-status-report "implemented + go test ./... pass"
EOF
}

require_arg() {
  local value="${1:-}"
  local name="${2:-arg}"
  if [[ -z "${value}" ]]; then
    echo "missing required argument: ${name}" >&2
    usage
    exit 1
  fi
}

cmd="${1:-}"
case "${cmd}" in
  ready)
    roady task ready
    ;;
  status)
    roady status
    roady drift detect
    ;;
  start)
    task_id="${2:-}"
    require_arg "${task_id}" "task-id"
    roady task start "${task_id}"
    roady status
    ;;
  complete)
    task_id="${2:-}"
    evidence="${3:-}"
    require_arg "${task_id}" "task-id"
    require_arg "${evidence}" "evidence"
    roady task complete "${task_id}" -e "${evidence}"
    roady status
    ;;
  verify)
    task_id="${2:-}"
    evidence="${3:-}"
    require_arg "${task_id}" "task-id"
    require_arg "${evidence}" "evidence"
    roady task verify "${task_id}" -e "${evidence}"
    roady status
    roady drift detect
    ;;
  cycle)
    task_id="${2:-}"
    evidence="${3:-}"
    require_arg "${task_id}" "task-id"
    require_arg "${evidence}" "evidence"
    roady task complete "${task_id}" -e "${evidence}"
    roady task verify "${task_id}" -e "${evidence}"
    roady status
    roady drift detect
    ;;
  *)
    usage
    exit 1
    ;;
esac

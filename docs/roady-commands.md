# Roady Command Catalog

## Daily Flow
```bash
roady status
roady task ready
roady drift detect
```

## Spec and Plan
```bash
roady spec add "Feature Name" "Description"
roady spec analyze docs --reconcile
roady plan generate
roady plan approve
```

## Task Lifecycle
```bash
roady task start <task-id>
roady task complete <task-id> -e "implemented"
roady task verify <task-id> -e "tests passed"
```

## Helper Scripts
```bash
bash ci/roady_task.sh ready
bash ci/roady_task.sh cycle <task-id> "implemented + tests passed"
bash ci/roady_docs_sync.sh
bash ci/roady_docs_sync.sh docs --analyze
bash ci/test_roady_helpers.sh
bash ci/roady_preflight.sh
bash ci/roady_bootstrap.sh
bash ci/roady_bootstrap.sh docs --analyze
```

## Diagnostics
```bash
roady debt summary
roady debt report
roady drift detect
```

## Notes
- If `task start` fails with plan approval errors, run `roady plan approve` and retry.
- Keep evidence strings short and concrete (for example: `go test ./... pass`).
- CI runs Roady checks via `.github/workflows/ci.yml` (`test_roady_helpers.sh` and `roady_preflight.sh`).

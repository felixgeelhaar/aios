# Roady MCP and Task Workflow

See also: `docs/roady-commands.md` for a compact command reference.

## Start Roady MCP
Run Roady as an MCP server for local tools/agents:

```bash
roady mcp --transport stdio
```

Optional network transports:

```bash
roady mcp --transport http --addr :8080
roady mcp --transport ws --addr :8080
```

## Generate Tasks from Docs
Use Roady spec features as the source of truth, then generate tasks:

```bash
roady spec add "Feature Name" "Short feature description"
roady plan generate
roady status
```

If docs are highly structured with feature-style sections, try:

```bash
roady spec analyze docs --reconcile
```

If analysis cannot infer features, use explicit `roady spec add` entries.

Or run the helper script to perform analyze + plan generate + approve + status:

```bash
ci/roady_docs_sync.sh
ci/roady_docs_sync.sh docs --analyze
```

## Keep Todo List Accurate
During implementation:

```bash
roady task start <task-id>
roady task complete <task-id>
roady task verify <task-id> -e "evidence"
```

Or use the helper script to enforce a consistent lifecycle:

```bash
ci/roady_task.sh ready
ci/roady_task.sh start <task-id>
ci/roady_task.sh cycle <task-id> "implemented + go test ./... pass"
```

Sanity checks:

```bash
roady drift detect
roady status
```

Helper script smoke tests:

```bash
bash ci/test_roady_helpers.sh
```

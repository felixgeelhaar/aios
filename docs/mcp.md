# MCP Usage

## Run MCP Server
```bash
./aios --mode cli --command serve-mcp --mcp-transport stdio
./aios --mode cli --command serve-mcp --mcp-transport http --mcp-addr :8080
```

## Available Tools
- `evaluate_policy`
- `execute_skill`
- `sync_state`
- `validate_skill_dir`
- `run_fixture_suite`
- `doctor`
- `model_policy_packs`
- `analytics_summary`
- `marketplace_publish`
- `marketplace_list`
- `marketplace_install`
- `governance_audit_export`
- `governance_audit_verify`
- `runtime_execution_report_export`
- `package_skill`
- `uninstall_skill`
- `project_list`
- `project_track`
- `project_untrack`
- `project_inspect`
- `workspace_validate`
- `workspace_plan`
- `workspace_repair`

## Available Resources
- `aios://status/health`
- `aios://status/sync`
- `aios://status/build`
- `aios://projects/inventory`
- `aios://workspace/links`
- `aios://analytics/trend`
- `aios://marketplace/compatibility`
- `aios://help/commands`
- `docs://index`
- `docs://{name}` (e.g. `docs://prd`, `docs://tdd`, `docs://cli`)

## Example MCP Operations
- Validate and package a skill directory:
  1. Call `validate_skill_dir` with `{"skill_dir":"./my-skill"}`.
  2. Call `package_skill` with `{"skill_dir":"./my-skill"}`.
- Execute with policy hooks:
  - Call `execute_skill` with `{"id":"roadmap-reader","version":"0.1.0","input":{"query":"..."}}`.
  - Response includes `policy_telemetry` (`violations`, `redactions`, `blocked`).
- Run fixture tests via MCP:
  - Call `run_fixture_suite` with `{"skill_dir":"./my-skill"}`.
- Inspect model routing policy packs:
  - Call `model_policy_packs` with `{}`.
- Read analytics summary:
  - Call `analytics_summary` with `{}`.
- Publish and install marketplace skills:
  1. Call `marketplace_publish` with `{"skill_dir":"./my-skill"}`.
  2. Call `marketplace_list` with `{}`.
  3. Call `marketplace_install` with `{"skill_id":"roadmap-reader"}`.
- Export signed governance audit bundle:
  - Call `governance_audit_export` with `{}` (or `{"output":"./audit.json"}`).
- Verify signed governance audit bundle:
  - Call `governance_audit_verify` with `{}` (or `{"input":"./audit.json"}`).
- Export runtime execution report:
  - Call `runtime_execution_report_export` with `{}` (or `{"output":"./runtime-report.json"}`).
- Track and repair project workspace links:
  1. Call `project_track` with `{"path":"./repo-a"}`.
  2. Call `project_inspect` with `{"selector":"./repo-a"}`.
  3. Call `workspace_validate` with `{}`.
  4. Call `workspace_plan` with `{}`.
  5. Call `workspace_repair` with `{}`.

- `dependency-matrix`: dependency usage and package standards

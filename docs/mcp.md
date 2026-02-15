# MCP Usage

The Model Context Protocol (MCP) server enables AI agents to interact with AIOS capabilities.

## Run MCP Server

```bash
# Stdio transport (default, for direct agent integration)
aios mcp serve

# HTTP transport
aios mcp serve --transport http --addr :8080

# WebSocket transport
aios mcp serve --transport ws --addr :8081
```

## Available Tools

### Skills
- `skill_init` - Create skill scaffold
- `skill_sync` - Sync skill to agents
- `skill_sync_plan` - Dry-run sync plan
- `skill_test` - Run fixture tests
- `skill_lint` - Validate skill structure
- `skill_package` - Package skill for distribution
- `skill_uninstall` - Remove skill from agents
- `validate_skill_dir` - Validate skill directory

### Analytics & Projects
- `analytics_summary` - Get analytics overview
- `analytics_record` - Record analytics snapshot
- `analytics_trend` - Show trend data
- `project_list` - List tracked projects
- `project_track` - Add project
- `project_untrack` - Remove project
- `project_inspect` - Inspect project

### Workspace
- `workspace_validate` - Validate workspace links
- `workspace_plan` - Plan workspace repairs
- `workspace_repair` - Apply workspace repairs

### Marketplace
- `marketplace_publish` - Publish skill
- `marketplace_list` - List marketplace skills
- `marketplace_install` - Install skill
- `marketplace_matrix` - Show compatibility matrix

### Governance
- `governance_audit_export` - Export audit bundle
- `governance_audit_verify` - Verify audit bundle

### Runtime
- `runtime_execution_report_export` - Export execution report
- `doctor` - Run diagnostics

### Model Routing
- `model_policy_packs` - List policy packs

## Available Resources

- `aios://status/health` - Health status
- `aios://status/sync` - Sync state
- `aios://status/build` - Build info
- `aios://projects/inventory` - Project list
- `aios://workspace/links` - Workspace links
- `aios://analytics/trend` - Analytics trends
- `aios://marketplace/compatibility` - Marketplace matrix
- `aios://help/commands` - Command help
- `docs://index` - Documentation index
- `docs://{name}` - Specific doc (e.g., `docs://prd`, `docs://cli`)

## Example MCP Operations

### Validate and Package a Skill

```json
// Step 1: Validate
{
  "name": "validate_skill_dir",
  "arguments": {"skill_dir": "./my-skill"}
}

// Step 2: Package
{
  "name": "skill_package",
  "arguments": {"skill_dir": "./my-skill"}
}
```

### Run Skill Tests

```json
{
  "name": "skill_test",
  "arguments": {"skill_dir": "./my-skill"}
}
```

### Sync Skill to Agents

```json
{
  "name": "skill_sync",
  "arguments": {"skill_dir": "./my-skill"}
}
```

### List Projects

```json
{
  "name": "project_list",
  "arguments": {}
}
```

### Track New Project

```json
{
  "name": "project_track",
  "arguments": {"path": "./my-project"}
}
```

### Validate Workspace

```json
{
  "name": "workspace_validate",
  "arguments": {}
}
```

### Repair Workspace

```json
{
  "name": "workspace_plan",
  "arguments": {}
}
{
  "name": "workspace_repair",
  "arguments": {}
}
```

### Publish to Marketplace

```json
{
  "name": "marketplace_publish",
  "arguments": {"skill_dir": "./my-skill"}
}
{
  "name": "marketplace_list",
  "arguments": {}
}
{
  "name": "marketplace_install",
  "arguments": {"skill_id": "my-skill"}
}
```

### Export Audit Bundle

```json
{
  "name": "governance_audit_export",
  "arguments": {}
}
```

### Read Documentation

```json
{
  "name": "read_resource",
  "arguments": {"uri": "docs://cli"}
}
```

## Configuration

The MCP server can be configured via environment:

```bash
# Custom port
aios mcp serve --transport http --addr :9000

# With JSON output
aios mcp serve --output json
```

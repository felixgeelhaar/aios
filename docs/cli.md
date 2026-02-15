# CLI Usage

AIOS provides a modern CLI with subcommands. Use `aios <command>` or `aios <parent> <subcommand>`.

## Global Flags
- `--output text|json` - Output format (default: text)
- `--verbose` - Enable verbose output
- `-y, --yes` - Skip confirmation prompts for destructive operations

## Skills

Manage skill lifecycle: create, sync, test, lint, package, and uninstall.

```bash
# Create a new skill scaffold
aios skills init my-skill

# Sync skill to all installed agents
aios skills sync ./my-skill

# Dry-run: see what would be written
aios skills plan ./my-skill

# Run fixture tests
aios skills test ./my-skill

# Lint skill structure
aios skills lint ./my-skill

# Package for distribution
aios skills package ./my-skill

# Uninstall from all agents
aios skills uninstall ./my-skill
```

## Runtime & Status

```bash
# Show runtime health and status
aios status

# Show detailed version info
aios version

# Run diagnostic checks
aios doctor

# List detected agent clients
aios list-clients

# Show model policy packs
aios model-policy-packs
```

## Projects

Track and manage projects for skill routing.

```bash
# List tracked projects
aios project list

# Add a project
aios project add ./my-project

# Remove a project
aios project remove my-project-id

# Inspect project details
aios project inspect my-project-id
```

## Workspace

Validate and repair workspace symlinks.

```bash
# Check workspace health
aios workspace validate

# Plan repairs (dry-run)
aios workspace plan

# Apply repairs
aios workspace repair
```

## Analytics

Track usage trends and workspace metrics.

```bash
# Show analytics summary
aios analytics summary

# Record a snapshot
aios analytics record

# Show trends
aios analytics trend
```

## Marketplace

Publish and install skills from the marketplace.

```bash
# List available skills
aios marketplace list

# Install a skill
aios marketplace install ddd-expert

# Publish a skill
aios marketplace publish ./my-skill

# Show compatibility matrix
aios marketplace matrix
```

## Governance

Audit and compliance features.

```bash
# Export audit bundle
aios audit export

# Verify audit bundle
aios audit verify ./audit.json
```

## Runtime

Execution reporting.

```bash
# Export execution report
aios runtime execution-report
```

## MCP Server

Start the Model Context Protocol server.

```bash
# Stdio transport (default)
aios mcp serve

# HTTP transport
aios mcp serve --transport http --addr :8080

# WebSocket transport
aios mcp serve --transport ws --addr :8081
```

## TUI

Launch the interactive terminal UI.

```bash
aios tui
```

## Backup & Restore

```bash
# Backup client configs
aios backup-configs

# Restore from backup
aios restore-configs
aios restore-configs ./backup-2024-01-15
```

## OAuth

Connect external services.

```bash
# Connect Google Drive
aios connect-google-drive
```

Environment variables for OAuth:
- `AIOS_OAUTH_TOKEN` - Bypass callback, use token directly
- `AIOS_OAUTH_STATE` - Expected callback state (default: aios)
- `AIOS_OAUTH_TIMEOUT_SEC` - Callback timeout in seconds (default: 120)

## JSON Output

All commands support `--output json` for machine-readable output.

```bash
aios status --output json
aios list-clients --output json
aios skills sync ./my-skill --output json
```

## Examples

```bash
# Full skill workflow
aios skills init my-skill
aios skills lint ./my-skill
aios skills test ./my-skill
aios skills sync ./my-skill
aios skills package ./my-skill

# Project setup
aios project add ./my-project
aios workspace validate
aios workspace repair

# Check system health
aios doctor
aios status
```

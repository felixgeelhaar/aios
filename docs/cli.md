# CLI Usage

## Global Flags
- `--mode tray|cli` (default `tray`)
- `--command <name>` (CLI mode only)
- `--skill-dir <dir>` for skill commands, and as optional path for `restore-configs` / `export-status-report`
- `--mcp-transport stdio|http|ws` and `--mcp-addr :8080` for `serve-mcp`
- `--output text|json` for machine-readable responses
- OAuth env overrides for `connect-google-drive`:
  - `AIOS_OAUTH_TOKEN` bypasses callback and uses the provided token/code directly
  - `AIOS_OAUTH_STATE` sets expected callback state (default `aios`)
  - `AIOS_OAUTH_TIMEOUT_SEC` sets callback wait timeout (default `120`)

## Commands
- `status`: show runtime status and sync state
- `tray-status`: show tray-facing skills + connections state
- `version`: show version, commit, build date
- `doctor`: run readiness checks for workspace/client dirs
- `model-policy-packs`: inspect available model routing policy packs (`cost-first`, `quality-first`, `balanced`)
- `analytics-summary`: export tracked project + workspace link summary in text or JSON
- `analytics-record`: persist one analytics snapshot into `workspace/state/analytics-history.json`
- `analytics-trend`: show trend summary from persisted analytics snapshots
- `marketplace-publish`: publish a skill directory to local marketplace registry
- `marketplace-list`: list marketplace listings from local registry
- `marketplace-install`: install a marketplace skill by id (`--skill-dir <skill-id>`)
- `marketplace-matrix`: show compatibility matrix summary (skill/client support + verification)
- `audit-export`: export signed governance audit bundle (`--skill-dir <output-file>` optional)
- `audit-verify`: verify signature of governance audit bundle (`--skill-dir <input-file>` optional)
- `runtime-execution-report`: export structured runtime execution report (`--skill-dir <output-file>` optional)
- `project-list`: list tracked projects/folders inventory
- `project-add`: track a project folder (`--skill-dir <path>`)
- `project-remove`: untrack by path or id (`--skill-dir <path-or-id>`)
- `project-inspect`: inspect one tracked project by path or id (`--skill-dir <path-or-id>`)
- `workspace-validate`: validate managed workspace symlinks for tracked projects
- `workspace-plan`: show create/repair/skip actions for workspace symlinks
- `workspace-repair`: apply create/repair actions for workspace symlinks
- `tui`: open the terminal operations console (projects + workspace link operations)
- `sync`: validate and install skill into all clients
- `backup-configs`: snapshot Claude/Cursor/Windsurf config directories into `workspace/backups/<timestamp>`
- `restore-configs`: restore configs from latest backup (or `--skill-dir <backup-dir>`)
- `export-status-report`: write markdown readiness report (default `workspace/status-report.md` or `--skill-dir <output-file>`)
- `connect-google-drive`: connect Google Drive via local OAuth callback or token override
- `sync-plan`: dry-run listing of files that would be written
- `test-skill`: run fixture suite in `<skill-dir>/tests`
- `lint-skill`: run strict lint checks for skill structure and fixture pairing
- `init-skill`: scaffold a skill directory
- `serve-mcp`: run MCP server
- `help`: print command reference

## JSON Examples
```bash
./aios --mode cli --command status --output json
./aios --mode cli --command sync-plan --skill-dir ./my-skill --output json
./aios --mode cli --command version --output json
./aios --mode cli --command restore-configs --output json
./aios --mode cli --command export-status-report --skill-dir ./status.md --output json
AIOS_OAUTH_TOKEN=demo-token ./aios --mode cli --command connect-google-drive --output json
./aios --mode cli --command project-add --skill-dir ./repo-a --output json
./aios --mode cli --command workspace-validate --output json
./aios --mode cli --command workspace-repair --output json
./aios --mode cli --command analytics-summary --output json
./aios --mode cli --command analytics-record --output json
./aios --mode cli --command analytics-trend --output json
./aios --mode cli --command marketplace-publish --skill-dir ./my-skill --output json
./aios --mode cli --command marketplace-list --output json
./aios --mode cli --command marketplace-install --skill-dir roadmap-reader --output json
./aios --mode cli --command marketplace-matrix --output json
./aios --mode cli --command audit-export --output json
./aios --mode cli --command audit-verify --output json
./aios --mode cli --command runtime-execution-report --output json
```

# Troubleshooting

## `sync` fails with `skill-dir is required`
Cause: `--skill-dir` was not provided.
Fix:
```bash
./aios --mode cli --command sync --skill-dir ./my-skill
```

## `lint-skill` reports missing fixture pair
Cause: a `fixture_*.json` file does not have matching `expected_*.json` (or vice versa).
Fix: ensure matching suffixes in `tests/`.

## `serve-mcp` fails on HTTP/WS
Cause: address is in use or invalid transport value.
Fix:
```bash
./aios --mode cli --command serve-mcp --mcp-transport http --mcp-addr :8081
```

## `doctor` reports failed checks
Cause: CLI cannot create or access workspace/client paths.
Fix: set environment overrides and re-run:
```bash
export AIOS_WORKSPACE_DIR="$PWD/.aios"
./aios --mode cli --command doctor
```

## `package-skill` fails
Cause: invalid `skill.yaml` or JSON schema files.
Fix:
1. Run `lint-skill`.
2. Run `test-skill`.
3. Re-run `package-skill`.

## `connect-google-drive` times out
Cause: no OAuth callback was received before timeout.
Fix:
1. Increase timeout: `AIOS_OAUTH_TIMEOUT_SEC=300`.
2. Confirm callback state matches `AIOS_OAUTH_STATE`.
3. For local testing, bypass callback with `AIOS_OAUTH_TOKEN=<token>`.

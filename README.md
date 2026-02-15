# AIOS

AIOS is a Go monorepo that provides a CLI and supporting services for managing AI skill workflows, project inventory, and operational governance.

License: Apache-2.0

## Requirements

- Go 1.25+

## Quickstart

Build the CLI:

```bash
go build ./cmd/aios
```

Run locally:

```bash
go run ./cmd/aios
```

## Testing

Run the full test suite:

```bash
go test ./...
```

Run focused tests:

```bash
go test ./internal/runtime -run TestHealthCheck
```

Coverage gate (per-domain):

```bash
coverctl check --config .coverctl.yaml
```

Security scan:

```bash
go run github.com/nox-hq/nox/cli@latest scan .
```

## Repository Layout

- `cmd/aios`: CLI entrypoint and bootstrapping
- `internal/domain/*`: Domain models and value objects
- `internal/application/*`: Use cases and orchestration
- `internal/runtime`: Runtime adapters
- `docs/`: PRD, roadmap, and CLI documentation
- `ci/`: CI helper scripts

## Contributing

- Use conventional commits
- Keep changes scoped and add tests for behavior changes
- Run `go test ./...` and `coverctl check --config .coverctl.yaml` before submitting

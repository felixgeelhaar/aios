# Dependency Integration Matrix

This project standardizes on Felix Geelhaar ecosystem packages where applicable.

| Package | Purpose in AIOS | Where used |
|---|---|---|
| `github.com/felixgeelhaar/bolt` | Structured logging | `internal/core/logger.go` |
| `github.com/felixgeelhaar/fortify` | Retry + circuit breaker resilience | `internal/runtime/runtime.go` |
| `github.com/felixgeelhaar/statekit` | Sync/drift state machine | `internal/sync/engine.go` |
| `github.com/felixgeelhaar/mcp-go` | MCP server, tools, and resources | `internal/mcp/server.go`, `internal/core/cli.go` |

## Notes
- MCP integration is implemented with `mcp-go` in this repository.
- `github.com/felixgeelhaar/mcp` is not currently available as a Go module; use `mcp-go` for all MCP work unless a maintained replacement is published.

## Update Rule
When adding infrastructure-level functionality, check this matrix first and prefer the listed package before introducing alternatives.

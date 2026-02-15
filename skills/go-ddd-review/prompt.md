# Go DDD Code Review

Review Go code changes for Domain-Driven Design compliance. Evaluate the code against these rules and report violations.

## Layer Boundaries

- **Domain layer** (`internal/domain/`): Must NEVER import `os`, `io`, `net`, `database/sql`, `log`, or any infrastructure package. Must not import application or adapter packages. Each bounded context must be isolated — no cross-BC imports.
- **Application layer** (`internal/application/`): May import its own BC's domain package only. Must not import adapters, core, runtime, or other BCs' domain packages.
- **Adapter/Infrastructure layer** (`internal/core/`, `internal/adapters/`): May import domain and application. Implements port interfaces defined in domain.

## Domain Model Purity

- Domain types must be pure value objects or aggregates with no side effects.
- Business logic belongs in domain methods (`Validate()`, constructors, query/mutation methods on aggregates), not in application services.
- Application services should be thin orchestrators: load from repo, call domain methods, save.
- Port interfaces (for I/O) must be defined in the domain package, implemented in adapters.

## Aggregate Design

- Aggregates should expose mutation methods that enforce invariants (e.g., `Track()` checks for duplicates before appending).
- Query methods on aggregates should return copies, not references to internal state.
- Constructors should compute derived fields (e.g., `NewTestSkillResult` computes `Failed` count).

## Port/Adapter Separation

- Functions that perform I/O (file read/write, network, database) must NOT live in domain or governance/observability packages.
- I/O must go through port interfaces with adapter implementations.
- Each adapter should have a compile-time interface check: `var _ SomePort = someAdapter{}`.

## Review Output

For each violation found, report:
1. **File and line** where the violation occurs
2. **Rule violated** (which of the above categories)
3. **Severity** (error for boundary violations, warning for design improvements)
4. **Suggestion** for how to fix it

If no violations are found, confirm the code is DDD-compliant.

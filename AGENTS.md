# AGENTS Guide

This file is written for autonomous coding agents operating in the `aios` repository. Follow it strictly when making changes.

---

## 1. Architecture & Project Structure

`aios` is a Go monorepo following DDD + Clean Architecture principles.

High-level layout:

- `cmd/aios`: CLI entrypoint and bootstrapping.
- `internal/domain/*`: Pure domain models, aggregates, value objects, domain services.
- `internal/application/*`: Use cases and orchestration logic.
- `internal/runtime`: Runtime adapters (health, oauth, tokens, factory, etc.).
- `internal/sync`, `internal/rollout`, `internal/governance`, etc.: Bounded contexts.
- `docs/`: PRD, roadmap, TDD, MCP, CLI documentation.
- `ci/`: CI helper scripts and coverage gates.
- `.roady/`: Roady-managed spec, plan, state, and policy artifacts.

Dependency direction must always be:

`domain -> application -> adapters/runtime -> cmd`

Domain must never depend on application or runtime packages.

---

## 2. Build, Lint, and Test Commands

### Core Build

- Build CLI:
  ```bash
  go build ./cmd/aios
  ```

- Run CLI locally:
  ```bash
  go run ./cmd/aios
  ```

---

### Testing

- Run full test suite:
  ```bash
  go test ./...
  ```

- Run tests with coverage:
  ```bash
  go test ./... -coverprofile=coverage.out -covermode=atomic
  ```

- Run a single package:
  ```bash
  go test ./internal/runtime
  ```

- Run a single test by name:
  ```bash
  go test ./internal/runtime -run TestHealthCheck
  ```

- Run a single test with verbose output:
  ```bash
  go test ./internal/runtime -run TestHealthCheck -v
  ```

Prefer running focused tests during iteration.

---

### Coverage Policy (Mandatory)

Coverage is enforced per domain using `coverctl`.

- Enforce thresholds:
  ```bash
  coverctl check --config .coverctl.yaml
  ```

- Show highest-value gaps:
  ```bash
  coverctl debt --config .coverctl.yaml
  ```

Do not consider work complete if `coverctl check` fails.

---

### Static Analysis & Security

- Vet:
  ```bash
  go vet ./...
  ```

- Security scan (SAST/IaC/dependencies):
  ```bash
  go run github.com/nox-hq/nox/cli@latest scan .
  ```

CI scripts under `ci/` may enforce additional checks (coverage gates, Roady sync, etc.).

---

## 3. Coding Standards

### Formatting

- Always format with `gofmt`.
- Use tabs (Go default), not spaces.
- Imports must be grouped:
  1. Standard library
  2. Third-party
  3. Internal (`aios/...`)
- No unused imports.

---

### Naming Conventions

- Package names: short, lowercase, no underscores.
- Exported identifiers: `PascalCase`.
- Unexported identifiers: `camelCase`.
- Test functions: `TestXxx_BehaviorDescription`.
- Interfaces: usually noun-based (`Store`, `Engine`, `Registry`).

Avoid stutter (e.g., `sync.SyncEngine` is acceptable only if aligned with bounded context).

---

### Types & Design

- Prefer explicit types over `interface{}`.
- Use value objects for immutable domain concepts.
- Keep domain models free of infrastructure concerns.
- Inject dependencies via constructors.
- Define interfaces at the consumer side, not the provider side.
- Keep functions small and single-purpose.

---

### Error Handling

- Never ignore returned errors.
- Wrap errors with context using `fmt.Errorf("...: %w", err)`.
- Do not panic in domain or application layers.
- Panics are only acceptable at process boundaries (CLI boot failures).
- Return typed errors where behavior depends on classification.

---

### Logging & Side Effects

- Domain layer must not log.
- Application layer may log via injected abstractions.
- Avoid global state.
- Avoid hidden side effects in constructors.

---

## 4. Testing Guidelines

- Tests must live next to implementation as `*_test.go`.
- Prefer table-driven tests.
- Test behavior, not implementation details.
- Avoid mocking domain logic.
- Use real structs and in-memory fakes where possible.
- Add regression tests for every bug fix.

Critical areas:

- `internal/runtime`
- `internal/governance`
- `internal/sync`
- `internal/rollout`

Coverage must meet thresholds defined in `.coverctl.yaml`.

---

## 5. Roady Integration Rules

This repository is Roady-managed.

- `.roady/spec.yaml` defines product specification.
- `.roady/plan.json` defines execution plan.
- `.roady/state.json` tracks task state.
- `.roady/policy.yaml` defines WIP and governance constraints.

Agents must:

- Align changes with existing features/tasks.
- Avoid modifying `.roady/` artifacts unless explicitly instructed.
- Keep implementation consistent with `docs/prd.md` and `docs/roadmap.md`.

---

## 6. Commit & PR Standards

Follow Conventional Commits:

- `feat(scope): add rollout evaluation engine`
- `fix(runtime): prevent nil pointer in token store`
- `test(sync): add deterministic ordering regression`
- `chore(ci): tighten coverage gate`

PR requirements:

- Link Roady feature/task ID.
- Include behavioral summary.
- Include risk notes.
- Include verification output:
  - `go test ./...`
  - `coverctl check`
  - `nox scan`

No unrelated refactors in feature PRs.

---

## 7. Cursor / Copilot Rules

There are currently:

- No `.cursor/rules/`
- No `.cursorrules`
- No `.github/copilot-instructions.md`

If such files are added in the future, they override conflicting guidance here.

---

## 8. Agent Behavior Expectations

When operating autonomously:

- Run focused tests before full suite.
- Maintain architectural boundaries.
- Do not introduce circular dependencies.
- Do not weaken coverage thresholds.
- Prefer proper solutions over workarounds.
- Add tests with every behavioral change.
- Keep diffs minimal and scoped.

This is production-grade infrastructure. Treat every change as long-lived.

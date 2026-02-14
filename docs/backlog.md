
## Local Runtime Foundation

Build local-first tray/runtime in Go with local registry, OAuth connector setup, secure token storage, Claude Desktop integration, and baseline logging.

---

## Skill Artifact & Execution Model

Define installable skill artifacts with metadata, input/output schemas, guardrails, fixtures, and execution validation with suggest-mode default.

---

## Quick Skill Builder

Ship wizard-based builder to generate production-ready skills without JSON editing, including schema generation and fixture validation.

---

## Sync Engine & Drift Protection

Implement file watchers, schema normalization, auto-repair, and parity enforcement across client config directories.

---

## Cursor Adapter

Add Cursor MCP adapter and schema translation so the same skill works consistently in Claude Desktop and Cursor.

---

## Cloud Skill Registry

Provide hosted registry with publish/install/versioning plus signing and verification workflows for team distribution.

---

## Bundles & Rollout Controls

Support department/team bundles, staged rollout, rollback, and version enforcement controls.

---

## RBAC & Governance Controls

Introduce role-based publishing, connector/scope restrictions, approval workflows, and audit logging.

---

## Windsurf & CLI Mode

Add Windsurf client adapter and headless CLI mode for CI and developer automation.

---

## Model Routing Abstraction

Enable transparent multi-model execution with policy-based routing, fallback handling, and cost tracking.

---

## Policy Engine

Implement prompt-injection safeguards, redaction, context classification, and enforceable guardrail policies.

---

## Analytics & Observability Platform

Deliver usage analytics, execution success metrics, adoption tracking, and drift trend reporting dashboards.

---

## Marketplace Ecosystem

Provide private/public skill registries, verification badges, and compatibility metadata for external distribution.

---

## OS Keychain Token Store

Replace in-memory token storage with OS keychain/credential manager-backed secure storage adapter.

---

## Real Client Sync Installers

Implement real install/sync behavior for Claude, Cursor, and Windsurf adapters and wire them through the sync engine.

---

## Skill Schema Validation Pipeline

Parse and validate skill.yaml and JSON schemas with strict checks before install or execution.

---

## MCP Runtime Execution Integration

Expose runtime and sync operations through MCP tools and add integration tests for end-to-end task execution paths.

---

## Persistent Registry and Governance Enforcement

Persist registry, rollout, and governance state and enforce RBAC at publish/install time.

---

## CI Quality and Security Gates

Add CI workflows for go test, coverage threshold, and security scanning before merge.

---

## Production Connector Security

Implement OS keychain-backed token storage and remove in-memory token storage from runtime path.

---

## Operational CLI and MCP Serve

Implement practical CLI commands for status/sync and an MCP serve command that starts the local MCP server over stdio.

---

## Persistent Rollout Plans

Persist rollout plans to disk and support loading/listing for repeatable team deployments.

---

## End-to-End Integration Tests

Add integration tests covering skill build->install->runtime->MCP tool registration and sync state transitions.

---

## Configurable MCP Transport

Support configurable MCP transport selection (stdio/http/ws) and bind address through CLI flags.

---

## Actionable CLI Sync Command

Implement sync command to install a skill across Claude/Cursor/Windsurf adapters with explicit target and skill id inputs.

---

## Runtime Health Reporting

Add runtime health/readiness report APIs and expose them via CLI status output.

---

## Config-Driven CLI Runtime

Load runtime and client paths from environment variables to support local/dev/prod execution without code edits.

---

## Validated Skill Sync Flow

Require sync to read a skill artifact directory, validate skill.yaml and schemas, then install to all clients.

---

## MCP Status Resources

Expose runtime health and sync status as MCP resources for agents to inspect state without tool invocation.

---

## Skill Fixture Test Runner

Add a runner that validates skill fixtures/expected outputs and expose it via CLI command for local quality checks.

---

## Sync Dry-Run Planning

Add a dry-run mode that shows what files/configs would be written for sync without mutating client configs.

---

## MCP Skill Validation Tool

Expose skill directory validation as an MCP tool so agents can preflight skills before sync.

---

## CLI Skill Scaffold Command

Add init-skill CLI command to scaffold a new skill directory using the quick builder.

---

## CLI Help Surface

Add explicit help command that prints available CLI commands and required flags.

---

## MCP Fixture Runner Tool

Expose running skill fixture suites via MCP tool for agent-driven quality checks.

---

## CLI Version and Build Metadata

Add version command and injectable build metadata surfaced via CLI and MCP resources.

---

## Skill Lint Command

Add lint-skill command with strict structural checks beyond schema validation (required tests and prompt file).

---

## MCP Command Reference Resource

Expose CLI command reference as MCP resource for discoverability by agents.

---

## CLI JSON Output Mode

Support --output json for status, sync-plan, and version commands for automation-friendly parsing.

---

## MCP Build Info Resource

Expose build/version metadata as an MCP resource for remote inspection.

---

## Lint Fixture Pair Consistency

Extend lint-skill to validate fixture/expected file pairing and fail on missing counterparts.

---

## CLI Doctor Command

Add doctor command to check workspace and client config directory readiness and report issues.

---

## MCP Doctor Tool

Expose readiness diagnostics via MCP tool for agent-accessible health triage.

---

## CLI Usage Documentation

---

## Tracked Projects and Folders Inventory

Provide a persistent inventory of tracked projects/folders across local workspaces so teams can see where skills are installed and detect unmanaged repos.

---

## Workspace and Symlink Orchestration

Implement workspace directory and symlink orchestration for multi-agent setups, including validation and repair flows for broken links.

---

## TUI Operations Console

Add a terminal UI for team/project-level skill operations: browse tracked projects, inspect status, run sync/uninstall, and triage drift.

Add docs/cli.md documenting commands, flags, and JSON output examples.

---

## MCP Docs Resources

Expose project docs via MCP resources so agents can read core repository documentation through the MCP server.

---

## CLI Client Inventory Command

Add list-clients command to report configured client directories and installed skill artifacts per client.

---

## Skill Package Command

Add package-skill command to create a distributable zip artifact from a validated skill directory.

---

## CLI Uninstall Skill Command

Add uninstall-skill command to remove installed skill artifacts from Claude/Cursor/Windsurf client configs.

---

## MCP Package Skill Tool

Expose skill packaging via MCP tool so agents can create distributable artifacts.

---

## MCP Usage Documentation

Add docs/mcp.md covering tools/resources and examples for local usage.

---

## CLI Backup Client Configs

Add backup-configs command to snapshot Claude/Cursor/Windsurf config directories into a timestamped backup folder.

---

## MCP Uninstall Skill Tool

Expose uninstalling a skill from all clients via MCP tool using the same validated skill directory input.

---

## Troubleshooting Documentation

Add docs/troubleshooting.md with common command failures and recovery steps.

---

## CLI Restore Client Configs

Add restore-configs command to restore client directories from a backup snapshot path.

---

## MCP Docs Index Resource

Expose docs index resource listing available docs names and URIs.

---

## CLI Export Status Report

Add export-report command to write current status/doctor/build info to a markdown report file.

---

## Dependency Integration Matrix

Document where bolt, fortify, statekit, and mcp-go are used, and clarify that mcp-go is the maintained MCP module.

---

## Roady Task Lifecycle Script

Add a helper script that wraps start/complete/verify/status and drift checks for Roady tasks.

---

## Roady Docs Sync Script

Add a helper script to analyze docs, reconcile spec, regenerate plan, and show status.

---

## Roady Helper Script Smoke Tests

Add a CI-friendly smoke test script that validates roady helper scripts syntax and help output.

---

## Roady Command Catalog

Add a concise command catalog for roady workflows and local helper scripts.

---

## Roady Preflight Script

Add a preflight script that runs helper smoke tests plus roady status/drift/debt checks before work.

---

## Roady Bootstrap Script

Add a bootstrap helper that runs preflight, docs sync, and final status checks in sequence.

---

## Roady CI Workflow

Add a CI workflow that runs Roady helper smoke tests and preflight checks on push and pull requests.

---

## H1 Local OAuth Callback Server

Implement a local OAuth callback HTTP server for connector auth with state validation and callback result handling.

---

## H1 File Watch Drift Monitor

Implement a polling file watch monitor that detects local config drift and invokes auto-repair hooks.

---

## H1 OAuth Callback Runtime

Implement local OAuth callback server utilities with state validation and code capture for connector auth.

---

## H1 Google Drive Connector CLI Flow

Add connect-google-drive CLI command using OAuth callback runtime or token override and persist token via runtime connector path.

---

## H1 Tray Skills and Connections Surface

Implement tray state persistence and CLI accessors for skills and connector status to provide a basic tray-facing UI surface.

---

## H1 Onboarding Path Integration Test

Add an integration test that simulates core H1 onboarding: skill creation, google drive connect, sync install, and tray status verification.

---

## H3 Strict DDD Boundary Enforcement

Refactor remaining non-DDD modules so domain, application, and adapter layers are explicit and independently testable.

---

## H3 Model Router Policy Packs

Add policy-pack driven model routing profiles (cost-first, quality-first, balanced) with CLI and MCP inspection outputs.

---

## H3 Policy Redaction Runtime Hooks

Integrate runtime redaction and prompt-injection checks as mandatory pre-execution hooks with structured violation telemetry.

---

## H3 Analytics Export Surfaces

Add CLI and MCP analytics summaries for skill usage, success rates, and drift trends with machine-readable JSON output.

---

## H3 Marketplace Compatibility Validation

Implement compatibility contract checks and verification badge criteria validation before publish/install flows.

---

## H3 Policy-Aware Runtime Execution

Apply model policy-pack routing and policy hooks directly in runtime execution (not only MCP entrypoints), with deterministic behavior and tests.

---

## H3 Analytics Persistence and Trend Report

Persist analytics snapshots locally and add trend reports for success/drift over time via CLI and MCP resources.

---

## H3 Marketplace Publish and Install CLI/MCP

Add first-class publish/install/list marketplace operations over CLI and MCP, enforcing compatibility and verification contracts.

---

## H3 Governance Audit Export

Export signed audit bundles for policy decisions, rollout decisions, and marketplace verification outcomes.

---

## H3 Governance Audit Verify

Add signature verification for governance audit bundles and expose verify operations in CLI and MCP.

---

## H3 Marketplace Compatibility Matrix Resource

Expose compatibility matrix summaries (skill-by-client support and verification state) as MCP resources and CLI JSON output.

---

## H3 Runtime Execution Report Export

Export structured runtime execution reports (model, policy telemetry, outcomes) for post-run review and automation.

---

## H1 Local Kernel Scope

Horizon 1 (0-4 months): Local-first runtime with skill artifacts, OAuth connectors, sync engine, drift repair, CLI + MCP surface. Includes: tray runtime, skill model, quick builder, sync engine, drift protection, Claude/Cursor adapters, fixture runner, linting, doctor, health reporting, backup/restore, JSON output mode, MCP serve. Exit criteria: onboarding < 15 min, zero JSON editing, drift auto-resolve >= 90%, skill success rate >= 98%, connector bind > 95%.

---

## H2 Organizational Control Plane Scope

Horizon 2 (4-12 months): Team-wide skill distribution and version control. Includes: cloud skill registry, publish/install workflows, signing and verification, bundles and rollout controls, RBAC and governance, audit logging, Windsurf adapter, CLI headless mode, tracked project inventory, workspace orchestration, TUI operations console. Exit criteria: >= 3 teams using shared skills, org-wide rollout achieved, admin approval workflow functioning, version enforcement stable.

---

## H3 AI OS Platform Scope

Horizon 3 (12-24 months): Full AI control layer for organizations. Includes: model routing abstraction with policy packs, policy engine (prompt injection, redaction, context classification), analytics and observability platform, marketplace ecosystem (private/public registries, verification badges), governance audit export and verification, compatibility matrix, runtime execution report export. Exit criteria: policy enforcement coverage, audit completeness, enterprise adoption metrics, external skill developer publishes verified skill.

---

## H1 Non-Goals

Explicit v1 scope exclusions per PRD and TDD. NOT in scope for H1: no cloud registry, no org sharing, no RBAC, no marketplace, no policy engine, no hosted runtime, no advanced workflow graph builder, no centralized multi-tenant control plane, no model routing. These are deferred to H2/H3. Existing experimental implementations of these capabilities are not production-committed for H1.

---

## Security Invariants

Cross-cutting security requirements enforced across all horizons. No credentials embedded in skill artifacts - runtime rejects skills containing raw tokens, OAuth client secrets, or private keys. All connector tokens must use OS keychain adapter (macOS Keychain, Windows Credential Manager). Least-privilege connector scopes enforced at binding time. Policy hooks must run before model execution when enabled. No plaintext tokens written to disk. Skills must declare required scopes explicitly. Token expiration must be monitored and surfaced.

---

## CLI Contract Guarantees

All CLI commands must: support --output json with deterministic JSON schema, exit code 0 on success, exit non-zero on validation failure or error, produce structured error messages in JSON mode, support --mode cli for headless operation. Commands must not require interactive input when run with --mode cli. Flags must be validated before execution begins. Unknown flags must produce clear error messages.

---

## MCP Contract Guarantees

All MCP tools must: return structured results with explicit error fields, never panic during execution, validate inputs before processing, include execution metadata in responses. MCP resources must return consistent schema. Transport selection (stdio/http/ws) must not affect tool behavior. Tools must be discoverable via the MCP protocol. Resources must support URI-based access pattern.

---

## DDD Architecture Invariants

Strict dependency direction: domain -> application -> adapters/runtime -> cmd. Domain packages must never import application or runtime packages. Domain layer must not log, perform I/O, or depend on infrastructure. Application layer orchestrates domain logic via injected interfaces. Interfaces defined at consumer side, not provider side. No circular dependencies between packages. Each bounded context (sync, rollout, governance, etc.) owns its domain model.

---

## AC: Sync Engine

Acceptance criteria for Validated Skill Sync Flow and Sync Engine. (1) Must validate skill.yaml before writing to any client directory. (2) Must validate input/output JSON schemas before install. (3) Must fail fast on schema errors with structured error output. (4) Must support dry-run mode via sync-plan command. (5) Must not mutate client directories in dry-run mode. (6) Must normalize schema differences across clients (dash vs underscore, tool naming). (7) Must emit JSON output when --output json is set. (8) Must install to all configured client adapters in a single sync invocation.

---

## AC: OAuth and Connector Runtime

Acceptance criteria for OAuth Callback Runtime and Connector System. (1) Must validate OAuth state parameter on callback. (2) Must enforce configurable timeout (AIOS_OAUTH_TIMEOUT_SEC, default 120s). (3) Must support token override via AIOS_OAUTH_TOKEN env var for testing. (4) Must never log tokens or secrets to any output. (5) Must persist tokens exclusively to OS keychain adapter. (6) Must surface token expiration status in tray-status and health check. (7) Must bind connectors at execution time, not at install time. (8) Must reject skills that embed raw credentials.

---

## AC: Skill Artifact Model

Acceptance criteria for Skill Artifact and Execution Model. (1) Skill must include: metadata, version, required connectors, input/output schemas, guardrails, prompt template, test fixtures. (2) skill.yaml must conform to documented spec (id, name, version, type, clients, requires, inputs, outputs, guardrails, execution, tests). (3) Execution defaults to suggest mode. (4) Skill must pass lint and fixture validation before sync install. (5) Skill must be portable across supported clients without modification. (6) Skill version must follow semver. (7) No embedded credentials in any skill file.

---

## AC: Drift Detection and Repair

Acceptance criteria for Drift Protection and File Watch Monitor. (1) File watchers must observe all configured client config directories. (2) Must detect manual edits to client configs within polling interval. (3) Must validate detected differences before repair. (4) Must auto-repair safe config mismatches without user intervention. (5) Must surface non-repairable drift as alerts in status/health output. (6) Must maintain parity across all synced clients after repair. (7) Auto-resolve rate target: >= 90% of drift incidents.

---

## AC: OS Keychain Token Store

Acceptance criteria for OS Keychain Token Store and Production Connector Security. (1) Must use macOS Keychain on darwin, Windows Credential Manager on windows. (2) Must not fall back to in-memory storage in production runtime path. (3) Must fail securely with clear error if keychain is unavailable or locked. (4) Must never write plaintext tokens to disk or logs. (5) In-memory token store acceptable only for test fixtures. (6) Must support token refresh flow when connector supports it.

---

## AC: Runtime Health and Diagnostics

Acceptance criteria for Runtime Health Reporting and Doctor Command. (1) Health check must report: runtime status, sync state, connector status, token expiration. (2) Doctor must validate workspace directory existence and permissions. (3) Doctor must validate all configured client directories are accessible. (4) Must support environment variable overrides for all paths (AIOS_WORKSPACE_DIR). (5) Must output structured JSON when --output json is set. (6) Health endpoints must be exposed via MCP resources (aios://status/health, aios://status/sync). (7) Exit code must reflect health status (0 = healthy, non-zero = degraded).

---

## AC: Skill Linting and Testing

Acceptance criteria for Skill Lint, Fixture Test Runner, and Validation Pipeline. (1) lint-skill must check: skill.yaml presence and validity, prompt.md presence, schema files present, tests directory with fixtures. (2) Must validate fixture/expected file pairing - fail on missing counterparts. (3) test-skill must execute all fixtures in tests/ directory and compare against expected outputs. (4) Must report pass/fail per fixture with structured output. (5) Schema validation must reject malformed JSON schemas before install. (6) Both lint and test must be exposed as MCP tools. (7) Must support --output json for automation.

---

## AC: Quick Skill Builder

Acceptance criteria for Quick Skill Builder and init-skill. (1) Must scaffold complete skill directory structure (skill.yaml, prompt.md, schema.input.json, schema.output.json, tests/). (2) Generated skill must pass lint-skill without modification. (3) Must support skill type selection (reader, analyzer, screener, guardrail, action). (4) Generated schemas must be valid JSON Schema. (5) Must include at least one starter fixture/expected pair. (6) No JSON editing required by user to produce a working skill.

---

## AC: Cloud Registry and Distribution

Acceptance criteria for H2 Cloud Skill Registry and Distribution. (1) Must support skill publishing with version metadata. (2) Must enforce signing at publish time. (3) Must verify signatures before install. (4) Local tray syncs from org registry. (5) Version updates must propagate to all subscribed clients. (6) Must support rollback to previous version. (7) Must enforce version constraints per team/department. (8) Registry state must be persistent.

---

## AC: RBAC and Governance

Acceptance criteria for H2 RBAC and Governance Controls. (1) Must enforce role-based publishing permissions (admin, publisher, consumer). (2) Must restrict allowed connector scopes per role. (3) Must support skill approval workflows before distribution. (4) Must log all publish, install, and policy decisions to audit trail. (5) Audit bundles must be exportable and signed. (6) Audit signatures must be verifiable via CLI and MCP. (7) RBAC must be enforced at both publish and install time.

---

## AC: Rollout and Bundles

Acceptance criteria for H2 Bundles and Rollout Controls. (1) Must support department/team skill bundles. (2) Must support staged rollout with percentage-based targeting. (3) Must support rollback of individual skills and bundles. (4) Rollout plans must be persistable to disk. (5) Must support loading and listing persisted rollout plans. (6) Version enforcement must be visible in status output. (7) Must prevent downgrade unless explicitly forced.

---

## AC: Model Routing and Policy Engine

Acceptance criteria for H3 Model Routing and Policy Engine. (1) Model routing must support policy packs: cost-first, quality-first, balanced. (2) Routing must be transparent to skill execution. (3) Must support fallback to alternate model on failure. (4) Cost tracking per execution must be recorded. (5) Policy engine must detect prompt injection attempts. (6) Must redact sensitive data before model invocation. (7) Must classify context and enforce guardrail policies. (8) Policy hooks must run in runtime execution path, not only MCP entrypoints. (9) Violation telemetry must be structured and exportable.

---

## AC: Analytics and Observability

Acceptance criteria for H3 Analytics and Observability Platform. (1) Must track skill usage counts per skill and team. (2) Must track execution success/failure rates. (3) Must track adoption metrics across tracked projects. (4) Must persist analytics snapshots locally. (5) Must support trend reporting over time via CLI and MCP. (6) Must export machine-readable JSON. (7) Must track drift incidents and auto-resolve rates.

---

## AC: Marketplace Ecosystem

Acceptance criteria for H3 Marketplace. (1) Must support private org registries. (2) Must support public skill marketplace. (3) Must enforce compatibility contract checks before publish. (4) Must assign verification badges to skills meeting criteria. (5) Must include compatibility metadata (supported clients, versions). (6) Must expose compatibility matrix as MCP resource and CLI JSON output. (7) External developers must be able to publish verified skills. (8) Org must be able to install marketplace skills safely with contract validation.

---

## H1 Success Metrics

Measurable success criteria for Horizon 1 Local Kernel per PRD and Roadmap. (1) AI onboarding time < 15 minutes from install to working skill. (2) Zero manual JSON editing required for any user operation. (3) Successful connector binding rate > 95%. (4) Skill execution success rate > 98%. (5) Drift detection auto-resolve rate >= 90%. (6) At least 5 distinct skills built via Quick Builder. (7) At least 2 non-engineering users successfully onboarded. (8) 10-20 power users using weekly before H2 begins.

---

## H2 Success Metrics

Measurable success criteria for Horizon 2 Org Control Plane per PRD and Roadmap. (1) >= 3 teams using shared skills. (2) Org-wide skill rollout achieved at least once. (3) Admin approval workflow functioning end-to-end. (4) Version enforcement stable across all clients. (5) Org-wide skill rollout time < 1 hour. (6) Skill version adoption rate > 80%. (7) Measurable reduction in config support tickets. (8) First paid customers (Pro tier).

---

## H3 Success Metrics

Measurable success criteria for Horizon 3 AI OS Platform per PRD and Roadmap. (1) Policy enforcement coverage across all skill executions. (2) Audit trail completeness for all governance decisions. (3) Enterprise adoption metrics tracked and reported. (4) External skill developer publishes verified skill to marketplace. (5) Org installs marketplace skill safely with contract validation. (6) Model routing cost tracking operational. (7) Analytics dashboard used by platform teams for decisions.

---

## Data-Driven Agent Registry

Replace hardcoded 3-client list (Claude, Cursor, Windsurf) with a data-driven agent registry supporting 9 agents: OpenCode, Claude Code, Cursor, Codex, Gemini CLI, GitHub Copilot, Goose, Windsurf, Cline. Agent definitions are loaded from an embedded agents.json file with name, displayName, skillsDir, altSkillsDirs, globalSkillsDir, detectPaths, and universal flag. Domain layer defines AgentDefinition value object. No horizon naming in code.

---

## Universal Skill Routing with Symlinks

Implement duckrow-style universal skill routing. Skills are installed to a canonical .agents/skills/ directory. Universal agents (OpenCode, Codex, Gemini CLI, GitHub Copilot) read .agents/skills/ directly. Non-universal agents (Claude Code, Cursor, Goose, Windsurf, Cline) get symlinks from their agent-specific skills directory to the canonical location. Remove old per-client JSON/YAML config writers.

---

## Agent Auto-Detection and Targeting

Detect which AI coding agents are installed on the system by checking agent-specific config directories (e.g. ~/.claude, ~/.cursor, /opencode). Support --agents flag for selective targeting during sync/install/uninstall. Scan project folders to detect which agents have skills installed.

---

## Refactor Client References to Agent Registry

Replace all 16 hardcoded 3-client references across core/config, core/cli, core/tray_state, core/backup, core/restore, core/doctor, skillsync_adapters, skilluninstall_adapters, syncplan_adapters, runtime, registry/cloud, and marketplace/catalog to use the data-driven agent registry. Config struct uses dynamic agent dirs instead of ClaudeConfigDir/CursorConfigDir/WindsurfConfigDir.

---

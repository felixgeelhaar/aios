Product: AI OS (Working name: AI Skill Sync)
Doc Type: Technical Design Document (TDD)
Version: v0.1
Status: Draft
Primary Goal: Define a local-first AI skill runtime + sync system that evolves into an org control plane and ultimately an AI OS platform.
⸻ 0. Non-Negotiables

1. Local-first v1: no cloud dependency required to install/build/run skills.
2. No credentials inside skills: tokens live only in OS secure storage.
3. Zero JSON editing UX: users must never hand-edit hidden config files.
4. Cross-client parity: same skill behavior across supported clients.
5. Production-ready skills: typed, versioned, testable, observable.
6. Horizon discipline: v1 solves parity + builder + local connectors; Pro layers distribution + policy.
   ⸻
7. System Overview
   1.1 What We’re Building
   A tray application (menu-bar / system-tray) that provides:
   • A local runtime for skills (execution + connector binding)
   • A skill builder that generates production-grade skill artifacts
   • A sync engine that installs/updates skills into AI clients (Claude Desktop, then Cursor/Windsurf)
   • A connector layer (OAuth + API access) with secure credential storage
   • Basic observability and diagnostics
   1.2 What We’re Not Building (v1)
   • A hosted LLM runtime
   • A cloud marketplace
   • Org admin plane / RBAC
   • Full policy engine (OPA-like)
   • Arbitrary workflow graphs
   ⸻
8. Architecture
   2.1 Component Diagram (H1 Local Kernel)
   ┌───────────────────────────────┐
   │ Tray UI (WebView) │
   │ - Skills list │
   │ - Connections │
   │ - Quick Builder │
   │ - Logs │
   └───────────────┬───────────────┘
   │ local IPC/HTTP
   ┌───────────────▼───────────────┐
   │ Local Runtime (Go) │
   │ - Skill Registry (local) │
   │ - Skill Execution Engine │
   │ - Connector Manager │
   │ - Token Vault (OS secure) │
   │ - Sync Engine │
   │ - File Watchers │
   │ - Local API (MCP bridge opt.) │
   └───────┬─────────┬─────────────┘
   │ │
   │ ├───────────────┐
   │ │ OS Secure Store│
   │ │ Keychain/CM │
   │ └───────────────┘
   │
   ├───────────────┐
   │ AI Clients │
   │ - Claude Desk │
   │ - Cursor (H1.1)│
   │ - Windsurf (H2)│
   └───────────────┘
   2.2 Evolution (H2/H3)
   • H2 adds a cloud registry + team distribution
   • H3 adds policy engine, analytics, model routing abstraction, marketplace
   ⸻
9. Data Model
   3.1 Local Data Directory Layout
   ~/.aios/
   registry/
   skills/
   /
   skill.yaml
   prompt.md
   schema.input.json
   schema.output.json
   tests/
   fixture_01.json
   expected_01.json
   changelog.md
   bundles/
   /
   bundle.yaml
   connections/
   connections.yaml (metadata only; no secrets)
   state/
   installed_clients.json
   sync_state.json
   logs/
   runtime.log
   skills/
   .log
   No tokens stored in plaintext here.
   ⸻
10. Skill Artifact Spec (v0.1)
    Goal: an installable, portable, versioned unit.
    4.1 skill.yaml (example)
    id: roadmap-reader
    name: Roadmap Reader
    version: 0.1.0
    description: Summarizes roadmap docs and extracts risks, timeline, and dependencies.
    author: internal
    license: proprietary
    type: reader # reader | analyzer | screener | guardrail | action
    clients:

- claude_desktop
- cursor
  requires:
  connectors:
- id: gdrive
  scopes:
- files.readonly
  optional: false
  inputs:
  schema: schema.input.json
  outputs:
  schema: schema.output.json
  guardrails:
- id: no_sensitive_export
- id: cite_sources_if_available
- id: no_external_sharing
  execution:
  mode: suggest # suggest | auto
  prompt: prompt.md
  timeout_ms: 15000
  tests:
- name: basic_summary
  fixture: tests/fixture_01.json
  expected: tests/expected_01.json
  4.2 Prompt Template (prompt.md)
  • Structured prompt, including:
  • instructions
  • formatting requirements
  • safety constraints
  • tool usage guidance
  4.3 Typed Schemas
  • schema.input.json and schema.output.json allow:
  • consistent behavior across clients
  • validation
  • test harness reliability
  ⸻

5. Connector System
   5.1 Connector Definition
   Connectors are centrally implemented by the runtime.
   Skill declares dependency; runtime binds at execution.
   Key outcomes:
   • Skills remain safe and portable
   • Credentials never shipped inside skills
   • Admin/policy becomes feasible later
   5.2 OAuth Flow (Local)
   • Tray app exposes a local callback server: http://127.0.0.1:/oauth/callback
   • User clicks “Connect Google Drive”
   • Browser consent → callback → token exchange
   • Tokens saved to OS secure storage
   5.3 Token Storage
   • macOS: Keychain
   • Windows: Credential Manager
   • Linux (later): Secret Service / keyring
   5.4 Connector Capabilities (H1)
   • gdrive.read_folder(folder_id)
   • gdrive.read_file(file_id)
   • gdrive.search(query)
   In v1, keep minimal.
   ⸻
6. Skill Execution Engine
   6.1 Execution Modes
   • Suggest mode (default): returns proposed actions / answers without writing to external systems.
   • Auto mode (optional later): allowed only for specific action skills with explicit user approval.
   6.2 Execution Steps
7. Validate skill version and schemas
8. Validate connector bindings available
9. Fetch required context via connectors (if needed)
10. Construct structured prompt
11. Call model via client (Claude Desktop) or via optional local model adapter (H3)
12. Validate output against schema
13. Emit logs + execution trace
    6.3 Progressive Disclosure
    To reduce context load:
    • skill prompt has outline + triggers
    • full instructions loaded only when invoked
    (Implementation: store prompt sections; runtime composes minimal prompt by default.)
    ⸻
14. Sync Engine (Parity Layer)
    7.1 Supported Clients (Phased)
    • H1: Claude Desktop
    • H1.1: Cursor
    • H2: Windsurf
    7.2 Claude Desktop Integration
    The sync engine:
    • Detects Claude config directories
    • Installs skills in expected format
    • Normalizes tool schema names
    • Ensures required MCP/connector endpoints are reachable (if used)
    7.3 Drift Detection
    • File watchers observe:
    • Claude skill directory
    • Cursor MCP config
    • If drift detected:
    • validate differences
    • auto-repair if safe
    • show alert in tray UI
    7.4 Schema Normalization
    Known issues:
    • Dash vs underscore differences
    • Tool naming constraints
    • JSON schema quirks
    We include a normalization layer:
    • canonical tool names stored in registry
    • per-client adapter transforms accordingly
    ⸻
15. Tray UI
    8.1 UX Goals
    • “Install → Connect → Works”
    • No technical jargon
    • Minimal configuration
    • Clear error states
    8.2 UI Tabs
    • Skills: installed, enable/disable, update
    • Connections: connect/disconnect, status, scopes
    • Build: quick builder wizard
    • Logs: recent runs + errors
    ⸻
16. Security Model
    9.1 Threat Model (H1)
    Threats:
    • Malicious skill artifact
    • Prompt injection via docs
    • Token leakage
    • Local process compromise
    • Overbroad connector scopes
    Mitigations:
    • Skills are local by default; user-installed
    • Signed skill artifacts (H2+)
    • Least privilege scopes
    • Token vault via OS secure storage
    • Guardrail enforcement and redaction (H2/H3)
    9.2 “No Embedded Secrets” Enforcement
    Runtime rejects any skill artifact containing:
    • raw tokens
    • OAuth client secrets
    • private keys
    ⸻
17. Pro / Cloud Extension Architecture (H2)
    10.1 Cloud Registry
    • skill publishing
    • version enforcement
    • rollback
    • approvals
    Local tray app becomes:
    • client that syncs from org registry
    10.2 Admin Controls
    • allowed connectors
    • allowed scopes
    • approved skill sources
    • enforced versions
    10.3 Signing
    • skills signed at publish time
    • tray verifies signatures before install
    ⸻
18. AI OS Platform (H3)
    11.1 Model Abstraction
    • route requests to OpenAI/Claude/local models
    • cost tracking
    • policy-driven model selection
    11.2 Policy Engine
    • prompt injection filters
    • PII redaction
    • audit trails
    • RBAC
    11.3 Analytics
    • usage by skill / team
    • success/failure
    • time saved proxies
    • drift incidents
    ⸻
19. Operational Considerations
    12.1 Packaging & Updates
    • auto-updater
    • code signing required (macOS, Windows)
    • safe rollback for tray binary
    12.2 Telemetry (v1)
    Local-only by default.
    Opt-in for diagnostics.
    ⸻
20. Implementation Plan (Engineering Milestones)
    H1:
    • Go runtime skeleton
    • local registry + file layout
    • connector manager + OAuth
    • skill schema + validator
    • Claude Desktop adapter
    • quick builder (wizard)
    • logs UI
    H1.1:
    • Cursor adapter
    • improved drift normalization
    H2:
    • cloud registry + signing
    • org distribution + RBAC
    H3:
    • policy engine + analytics + model routing
    ⸻
21. Appendix: Key Interfaces (Draft)
    • SkillRegistry:
    • ListSkills()
    • InstallSkill(skillArtifact)
    • EnableSkill(skillID, client)
    • RunTest(skillID, fixture)
    • ConnectorManager:
    • Connect(connectorID)
    • GetStatus(connectorID)
    • Execute(connectorID, action, params)
    • SyncEngine:
    • DetectClients()
    • SyncSkill(skillID, client)
    • DriftScan()
    • RepairDrift()

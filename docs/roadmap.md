Product: AI OS (Working name: AI Skill Sync)
Horizon: 0–24 Months
Strategy: Local-first kernel → Org control plane → AI OS platform
⸻ 0. Roadmap Philosophy
We are not building everything at once.
Each phase must:
• Deliver independent value
• Validate a critical assumption
• Unlock the next layer
• Maintain architectural integrity
We move from:
Local parity → Organizational distribution → Policy & control → Platform
⸻
Horizon 1 — Local AI Kernel (0–4 Months)
Objective
Eliminate AI onboarding purgatory and configuration drift for power users.
Success Definition
• New user installs tray → working skill in < 15 minutes
• Zero JSON editing required
• Claude Desktop fully supported
• Google Drive connector stable
• Drift detection functioning
⸻
Phase 1.1 — Runtime Foundation (Weeks 1–4)
Deliver:
• Go runtime skeleton
• Local skill registry
• File structure
• Secure token storage integration
• Local OAuth server
• Basic tray UI (Skills + Connections)
• Claude Desktop adapter (manual install first)
Success Gate:
• User can install skill artifact locally
• User can connect Google Drive
• User can run Roadmap Reader
⸻
Phase 1.2 — Quick Skill Builder (Weeks 5–8)
Deliver:
• Wizard-based builder
• Skill type taxonomy
• Schema generation
• Prompt template compiler
• Test fixture runner
• Skill validation engine
Success Gate:
• PM can build Roadmap Reader skill without coding
• Skill passes test fixture
• Skill installs automatically into Claude
⸻
Phase 1.3 — Drift Protection & Stability (Weeks 9–12)
Deliver:
• File watchers
• Schema normalization layer
• Auto-repair of client config
• Logging UI
• Error handling improvements
Success Gate:
• Manual client config edits are auto-detected
• Parity maintained after Claude update
⸻
Phase 1.4 — Cursor Adapter (Weeks 13–16)
Deliver:
• Cursor MCP config adapter
• Schema translation
• Multi-client sync validation
Success Gate:
• Same skill works in Claude + Cursor
• No schema naming mismatch issues
⸻
Horizon 1 Exit Criteria
Before moving to H2:
• 10–20 power users using weekly
• ≥80% skill success rate
• Drift incidents auto-resolved ≥90%
• At least 5 distinct skills built via Quick Builder
• At least 2 non-engineering users successfully onboarded
If these fail → iterate before scaling.
⸻
Horizon 2 — Organizational Control Plane (4–12 Months)
Objective
Enable team-wide skill distribution and version control.
Move from:
Local parity
to
Org-level capability governance.
⸻
Phase 2.1 — Cloud Skill Registry (Months 4–6)
Deliver:
• Hosted registry
• Skill publishing flow
• Version management
• Signing & verification
• Local client sync from cloud
• Tracked projects/folders inventory across local workspaces and repos
Success Gate:
• User builds skill locally
• Publishes to org registry
• Teammate installs in one click
• Version update propagates
⸻
Phase 2.2 — Bundles & Rollout (Months 6–8)
Deliver:
• Skill bundles
• Department-level rollout
• Default skill sets per team
• Update enforcement
• Rollback controls
Success Gate:
• Org admin rolls out skill pack to 20+ users
• Version enforcement visible in tray
⸻
Phase 2.3 — RBAC & Policy Controls (Months 8–10)
Deliver:
• Role-based skill publishing
• Allowed connector scopes
• Skill approval workflow
• Audit logs
Success Gate:
• Security lead can restrict Google Drive scope
• Admin sees execution logs
⸻
Phase 2.4 — Windsurf & CLI Support (Months 10–12)
Deliver:
• Windsurf adapter
• Optional CLI mode
• Headless mode for CI/dev environments
• Workspace folder + symlink orchestration for multi-agent setups
• TUI operations surface for team/project skill management
Success Gate:
• Cross-client orchestration across 3 major clients
• Admin can inspect tracked projects and fix broken links/symlinks from one console
⸻
Horizon 2 Exit Criteria
• ≥3 teams using shared skills
• Org-wide rollout achieved at least once
• Admin approval workflow functioning
• Version enforcement stable
• First paid customers (Pro)
⸻
Horizon 3 — AI OS Platform (12–24 Months)
Objective
Become the control layer for AI capability inside organizations.
⸻
Phase 3.1 — Model Abstraction (Months 12–15)
Deliver:
• Model routing layer
• Policy-based model selection
• Cost tracking
• Fallback handling
Success Gate:
• Skill can run on multiple models transparently
⸻
Phase 3.2 — Policy Engine (Months 15–18)
Deliver:
• Prompt injection detection
• Data redaction
• Context classification
• Policy enforcement rules
Success Gate:
• Sensitive data automatically redacted
• Guardrail violation logs visible
⸻
Phase 3.3 — Analytics & Observability (Months 18–21)
Deliver:
• Skill usage analytics
• Execution success metrics
• Adoption tracking
• Drift trend reports
Success Gate:
• Platform teams use dashboard for decisions
⸻
Phase 3.4 — Marketplace (Months 21–24)
Deliver:
• Private registry
• Public skill marketplace
• Skill verification badges
• Compatibility metadata
Success Gate:
• External skill developer publishes verified skill
• Org installs safely
⸻
Strategic Guardrails
We do NOT:
• Replace LLM providers
• Build a general workflow automation tool
• Compete with Zapier or LangGraph
• Host all AI execution (unless strategically necessary later)
We remain:
Capability + control layer.
⸻
Monetization Evolution
H1:
• Free local version
H2:
• Pro plan (org registry + rollout)
H3:
• Enterprise plan (policy engine + analytics + compliance)
⸻
Decision Points
At each horizon boundary:
• Is usage organic?
• Are users building skills?
• Is parity pain validated?
• Are teams requesting sharing?
• Are security teams asking for visibility?
If yes → scale.
If not → refine.
⸻
Final Strategic Summary
Horizon 1:
Solve real friction.
Horizon 2:
Enable teams.
Horizon 3:
Own AI control layer.
This is the path from:
Tray app
to
AI Operating System.

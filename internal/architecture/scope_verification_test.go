package architecture_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Local Kernel Scope Verification ---
// Verifies local-first runtime capabilities exist: tray runtime, skill model,
// quick builder, sync engine, drift protection, agent registry + skill installer,
// fixture runner, linting, doctor, health reporting, backup/restore,
// JSON output mode, MCP serve.

func TestLocalKernelScope_TrayRuntimeExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "runtime", "runtime.go"))
	requireFileExists(t, filepath.Join(root, "internal", "runtime", "factory.go"))
	requireFileExists(t, filepath.Join(root, "internal", "core", "tray_state.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "runtime", "runtime.go"), "Runtime")
	requireExportedIdent(t, filepath.Join(root, "internal", "runtime", "factory.go"), "NewProductionRuntime")
}

func TestLocalKernelScope_SkillModelExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "skill", "artifact.go"))
	requireFileExists(t, filepath.Join(root, "internal", "skill", "schema_validation.go"))
	requireFileExists(t, filepath.Join(root, "internal", "skill", "executor.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "skill", "artifact.go"), "Artifact")
	requireExportedIdent(t, filepath.Join(root, "internal", "skill", "schema_validation.go"), "SkillSpec")
	requireExportedIdent(t, filepath.Join(root, "internal", "skill", "schema_validation.go"), "ValidateSkillSpec")
}

func TestLocalKernelScope_QuickBuilderExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "builder", "builder.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "builder", "builder.go"), "BuildSkill")
}

func TestLocalKernelScope_SyncEngineExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "sync", "engine.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "sync", "engine.go"), "Engine")
	requireExportedIdent(t, filepath.Join(root, "internal", "sync", "engine.go"), "NewEngine")
}

func TestLocalKernelScope_DriftProtectionExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "sync", "watcher.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "sync", "watcher.go"), "Watcher")
	requireExportedIdent(t, filepath.Join(root, "internal", "sync", "engine.go"), "DetectDrift")
}

func TestLocalKernelScope_ClientAdaptersExist(t *testing.T) {
	root := findRepoRoot(t)
	// Agent registry and skill installer replaced per-client adapters.
	requireFileExists(t, filepath.Join(root, "internal", "agents", "registry.go"))
	requireFileExists(t, filepath.Join(root, "internal", "agents", "installer.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "agents", "installer.go"), "SkillInstaller")
	requireExportedIdent(t, filepath.Join(root, "internal", "agents", "registry.go"), "LoadAll")
}

func TestLocalKernelScope_FixtureRunnerExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "skill", "fixture_runner.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "skill", "fixture_runner.go"), "RunFixtureSuite")
}

func TestLocalKernelScope_LintingExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "skill", "lint.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "skill", "lint.go"), "LintSkillDir")
}

func TestLocalKernelScope_DoctorExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "core", "doctor.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "core", "doctor.go"), "RunDoctor")
}

func TestLocalKernelScope_HealthReportingExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "runtime", "health.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "runtime", "health.go"), "HealthReport")
	requireExportedIdent(t, filepath.Join(root, "internal", "runtime", "health.go"), "Health")
}

func TestLocalKernelScope_BackupRestoreExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "core", "backup.go"))
	requireFileExists(t, filepath.Join(root, "internal", "core", "restore.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "core", "backup.go"), "BackupClientConfigs")
	requireExportedIdent(t, filepath.Join(root, "internal", "core", "restore.go"), "RestoreClientConfigs")
}

func TestLocalKernelScope_JSONOutputModeExists(t *testing.T) {
	root := findRepoRoot(t)
	cliFile := filepath.Join(root, "internal", "core", "cli.go")
	requireFileExists(t, cliFile)
	// Verify the CLI processes JSON output mode by looking for writeJSON usage.
	content, err := os.ReadFile(cliFile)
	if err != nil {
		t.Fatalf("read cli.go: %v", err)
	}
	src := string(content)
	if !strings.Contains(src, "writeJSON") {
		t.Fatal("cli.go does not contain writeJSON — JSON output mode missing")
	}
	if !strings.Contains(src, `"json"`) {
		t.Fatal("cli.go does not reference json output mode")
	}
}

func TestLocalKernelScope_MCPServeExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "mcp", "server.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "mcp", "server.go"), "NewServer")
}

func TestLocalKernelScope_DomainBoundedContextsExist(t *testing.T) {
	root := findRepoRoot(t)
	domainDir := filepath.Join(root, "internal", "domain")
	expectedBCs := []string{
		"onboarding",
		"skillpackage",
		"skilllint",
		"skilltest",
		"skillsync",
		"skilluninstall",
		"syncplan",
	}
	for _, bc := range expectedBCs {
		bcDir := filepath.Join(domainDir, bc)
		info, err := os.Stat(bcDir)
		if err != nil || !info.IsDir() {
			t.Errorf("missing domain bounded context: %s", bc)
		}
	}
}

func TestLocalKernelScope_ApplicationServicesExist(t *testing.T) {
	root := findRepoRoot(t)
	appDir := filepath.Join(root, "internal", "application")
	expectedServices := []string{
		"onboarding",
		"skillpackage",
		"skilllint",
		"skilltest",
		"skillsync",
		"skilluninstall",
		"syncplan",
	}
	for _, svc := range expectedServices {
		svcDir := filepath.Join(appDir, svc)
		info, err := os.Stat(svcDir)
		if err != nil || !info.IsDir() {
			t.Errorf("missing application service: %s", svc)
		}
	}
}

// --- Org Control Plane Scope Verification ---
// Verifies team-wide distribution capabilities: cloud skill registry,
// publish/install, signing, bundles/rollout, RBAC/governance, audit logging,
// agent uninstall capability, tracked project inventory, workspace orchestration.

func TestOrgControlPlaneScope_CloudRegistryExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "registry", "cloud.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "registry", "cloud.go"), "CloudRegistry")
	requireExportedIdent(t, filepath.Join(root, "internal", "registry", "cloud.go"), "Publish")
}

func TestOrgControlPlaneScope_PublishContractValidationExists(t *testing.T) {
	root := findRepoRoot(t)
	cloudFile := filepath.Join(root, "internal", "registry", "cloud.go")
	content, err := os.ReadFile(cloudFile)
	if err != nil {
		t.Fatalf("read cloud.go: %v", err)
	}
	src := string(content)
	// Registry must validate contracts at publish time.
	if !strings.Contains(src, "validatePublishContract") {
		t.Fatal("cloud registry missing publish contract validation")
	}
	// Governance layer must support verification of audit bundles.
	auditFile := filepath.Join(root, "internal", "governance", "audit.go")
	auditContent, err := os.ReadFile(auditFile)
	if err != nil {
		t.Fatalf("read audit.go: %v", err)
	}
	if !strings.Contains(string(auditContent), "Verify") {
		t.Fatal("governance audit missing verification capability")
	}
}

func TestOrgControlPlaneScope_RolloutControlsExist(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "rollout", "rollout.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "rollout", "rollout.go"), "Plan")
}

func TestOrgControlPlaneScope_RBACGovernanceExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "governance", "rbac.go"))
	requireFileExists(t, filepath.Join(root, "internal", "governance", "audit.go"))
	requireFileExists(t, filepath.Join(root, "internal", "governance", "enforcement.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "governance", "rbac.go"), "Role")
}

func TestOrgControlPlaneScope_AuditLoggingExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "governance", "audit.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "governance", "audit.go"), "AuditRecord")
}

func TestOrgControlPlaneScope_AgentUninstallCapabilityExists(t *testing.T) {
	root := findRepoRoot(t)
	// Agent registry replaced per-client adapters; verify uninstall capability.
	requireFileExists(t, filepath.Join(root, "internal", "agents", "installer.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "agents", "installer.go"), "UninstallSkill")
}

func TestOrgControlPlaneScope_ProjectInventoryExists(t *testing.T) {
	root := findRepoRoot(t)
	domainDir := filepath.Join(root, "internal", "domain", "projectinventory")
	info, err := os.Stat(domainDir)
	if err != nil || !info.IsDir() {
		t.Fatal("missing domain bounded context: projectinventory")
	}
	appDir := filepath.Join(root, "internal", "application", "projectinventory")
	info, err = os.Stat(appDir)
	if err != nil || !info.IsDir() {
		t.Fatal("missing application service: projectinventory")
	}
}

func TestOrgControlPlaneScope_WorkspaceOrchestrationExists(t *testing.T) {
	root := findRepoRoot(t)
	domainDir := filepath.Join(root, "internal", "domain", "workspaceorchestration")
	info, err := os.Stat(domainDir)
	if err != nil || !info.IsDir() {
		t.Fatal("missing domain bounded context: workspaceorchestration")
	}
	appDir := filepath.Join(root, "internal", "application", "workspaceorchestration")
	info, err = os.Stat(appDir)
	if err != nil || !info.IsDir() {
		t.Fatal("missing application service: workspaceorchestration")
	}
}

func TestOrgControlPlaneScope_RolloutStoreExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "rollout", "store.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "rollout", "store.go"), "Store")
}

// --- Platform Scope Verification ---
// Verifies full control layer capabilities: model routing with policy packs,
// policy engine, analytics/observability, marketplace, governance audit,
// compatibility matrix.

func TestPlatformScope_ModelRoutingExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "model", "router.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "model", "router.go"), "Router")
}

func TestPlatformScope_ModelRoutingPolicyPacks(t *testing.T) {
	root := findRepoRoot(t)
	routerFile := filepath.Join(root, "internal", "model", "router.go")
	content, err := os.ReadFile(routerFile)
	if err != nil {
		t.Fatalf("read router.go: %v", err)
	}
	src := string(content)
	// Must support cost-first, quality-first, balanced policy packs.
	policies := []string{"cost", "quality", "balanced"}
	for _, p := range policies {
		if !strings.Contains(strings.ToLower(src), p) {
			t.Errorf("model router missing policy pack: %s", p)
		}
	}
}

func TestPlatformScope_PolicyEngineExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "policy", "engine.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "policy", "engine.go"), "Engine")
}

func TestPlatformScope_PolicyEngineCapabilities(t *testing.T) {
	root := findRepoRoot(t)
	engineFile := filepath.Join(root, "internal", "policy", "engine.go")
	content, err := os.ReadFile(engineFile)
	if err != nil {
		t.Fatalf("read engine.go: %v", err)
	}
	src := string(content)
	// Must support prompt injection detection, redaction, and context evaluation.
	capabilities := []string{"injection", "redact", "sanitize"}
	for _, cap := range capabilities {
		if !strings.Contains(strings.ToLower(src), cap) {
			t.Errorf("policy engine missing capability: %s", cap)
		}
	}
}

func TestPlatformScope_AnalyticsObservabilityExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "observability", "metrics.go"))
	requireFileExists(t, filepath.Join(root, "internal", "observability", "history.go"))
	requireFileExists(t, filepath.Join(root, "internal", "observability", "gates.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "observability", "metrics.go"), "Metrics")
	requireExportedIdent(t, filepath.Join(root, "internal", "observability", "history.go"), "Snapshot")
	requireExportedIdent(t, filepath.Join(root, "internal", "observability", "history.go"), "BuildTrend")
}

func TestPlatformScope_MarketplaceExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "marketplace", "catalog.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "marketplace", "catalog.go"), "Catalog")
}

func TestPlatformScope_MarketplaceCapabilities(t *testing.T) {
	root := findRepoRoot(t)
	catalogFile := filepath.Join(root, "internal", "marketplace", "catalog.go")
	content, err := os.ReadFile(catalogFile)
	if err != nil {
		t.Fatalf("read catalog.go: %v", err)
	}
	src := string(content)
	// Must support compatibility checks, verification badges, private/public registries.
	capabilities := []string{"compat", "verif", "badge"}
	for _, cap := range capabilities {
		if !strings.Contains(strings.ToLower(src), cap) {
			t.Errorf("marketplace catalog missing capability: %s", cap)
		}
	}
}

func TestPlatformScope_GovernanceAuditExportExists(t *testing.T) {
	root := findRepoRoot(t)
	auditFile := filepath.Join(root, "internal", "governance", "audit.go")
	content, err := os.ReadFile(auditFile)
	if err != nil {
		t.Fatalf("read audit.go: %v", err)
	}
	src := string(content)
	// Must support export and signing of audit bundles.
	if !strings.Contains(src, "Export") && !strings.Contains(src, "export") {
		t.Fatal("governance audit missing export capability")
	}
}

func TestPlatformScope_GateEvaluationEngineExists(t *testing.T) {
	root := findRepoRoot(t)
	requireFileExists(t, filepath.Join(root, "internal", "observability", "gates.go"))
	requireExportedIdent(t, filepath.Join(root, "internal", "observability", "gates.go"), "Gate")
	requireExportedIdent(t, filepath.Join(root, "internal", "observability", "gates.go"), "EvaluateGates")
	requireExportedIdent(t, filepath.Join(root, "internal", "observability", "gates.go"), "AllPassed")
}

// --- Scope Exclusion Verification ---
// Verifies that local kernel code paths do not depend on org control plane
// or platform infrastructure. Domain and core application services must
// remain independent of distribution, governance, and platform packages.

func TestScopeExclusion_DomainDoesNotImportDistributionInfra(t *testing.T) {
	root := findRepoRoot(t)
	// Domain layer must not directly depend on distribution/platform infrastructure.
	infraPackages := []string{
		modulePrefix + "internal/registry",
		modulePrefix + "internal/rollout",
		modulePrefix + "internal/governance",
		modulePrefix + "internal/model",
		modulePrefix + "internal/policy",
		modulePrefix + "internal/marketplace",
	}
	checkNoForbiddenImports(t, filepath.Join(root, "internal", "domain"), infraPackages)
}

func TestScopeExclusion_CoreApplicationServicesDoNotImportDistributionInfra(t *testing.T) {
	root := findRepoRoot(t)
	// Core application services must not import distribution/platform infrastructure.
	coreServices := []string{
		"onboarding",
		"skillpackage",
		"skilllint",
		"skilltest",
		"skillsync",
		"skilluninstall",
		"syncplan",
	}
	infraPackages := []string{
		modulePrefix + "internal/registry",
		modulePrefix + "internal/rollout",
		modulePrefix + "internal/governance",
		modulePrefix + "internal/model",
		modulePrefix + "internal/policy",
		modulePrefix + "internal/marketplace",
	}
	for _, svc := range coreServices {
		svcDir := filepath.Join(root, "internal", "application", svc)
		if _, err := os.Stat(svcDir); err != nil {
			continue // Skip if service doesn't exist (already caught by scope tests).
		}
		checkNoForbiddenImports(t, svcDir, infraPackages)
	}
}

func TestScopeExclusion_DistributionAndPlatformPackagesAreIsolated(t *testing.T) {
	root := findRepoRoot(t)
	// Distribution and platform packages must exist as standalone modules.
	isolatedDirs := []struct {
		name string
		dir  string
	}{
		{"registry", filepath.Join(root, "internal", "registry")},
		{"rollout", filepath.Join(root, "internal", "rollout")},
		{"governance", filepath.Join(root, "internal", "governance")},
		{"model", filepath.Join(root, "internal", "model")},
		{"policy", filepath.Join(root, "internal", "policy")},
		{"marketplace", filepath.Join(root, "internal", "marketplace")},
	}
	for _, pkg := range isolatedDirs {
		info, err := os.Stat(pkg.dir)
		if err != nil || !info.IsDir() {
			t.Errorf("package %s should exist as isolated package", pkg.name)
		}
	}
	// Distribution/platform packages must not be imported by the sync engine.
	// Runtime is the integration layer and may wire these capabilities,
	// but the sync engine must remain independent.
	infraPackagePaths := []string{
		modulePrefix + "internal/registry",
		modulePrefix + "internal/rollout",
		modulePrefix + "internal/governance",
		modulePrefix + "internal/model",
		modulePrefix + "internal/policy",
		modulePrefix + "internal/marketplace",
	}
	checkNoForbiddenImports(t, filepath.Join(root, "internal", "sync"), infraPackagePaths)
}

func TestScopeExclusion_AgentsDoNotImportDistributionInfra(t *testing.T) {
	root := findRepoRoot(t)
	// Agent registry and skill installer must not import
	// distribution or platform packages.
	infraPackages := []string{
		modulePrefix + "internal/registry",
		modulePrefix + "internal/rollout",
		modulePrefix + "internal/governance",
		modulePrefix + "internal/model",
		modulePrefix + "internal/policy",
		modulePrefix + "internal/marketplace",
	}
	checkNoForbiddenImports(t, filepath.Join(root, "internal", "agents"), infraPackages)
}

// --- helpers ---

func requireFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("required file missing: %s", path)
	}
	if info.IsDir() {
		t.Fatalf("expected file but found directory: %s", path)
	}
}

// requireExportedIdent verifies that a Go source file declares an exported
// identifier (type, func, var, const) with the given name.
func requireExportedIdent(t *testing.T, filePath, name string) {
	t.Helper()
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse %s: %v", filePath, err)
	}
	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.Name == name {
						return
					}
				case *ast.ValueSpec:
					for _, n := range s.Names {
						if n.Name == name {
							return
						}
					}
				}
			}
		case *ast.FuncDecl:
			if d.Name.Name == name {
				return
			}
		}
	}
	t.Fatalf("exported identifier %q not found in %s", name, filePath)
}

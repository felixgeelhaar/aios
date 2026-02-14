package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/felixgeelhaar/aios/internal/builder"
	domainprojectinventory "github.com/felixgeelhaar/aios/internal/domain/projectinventory"
	domainskilllint "github.com/felixgeelhaar/aios/internal/domain/skilllint"
	domainskillpackage "github.com/felixgeelhaar/aios/internal/domain/skillpackage"
	domainskillsync "github.com/felixgeelhaar/aios/internal/domain/skillsync"
	domainskilltest "github.com/felixgeelhaar/aios/internal/domain/skilltest"
	domainskilluninstall "github.com/felixgeelhaar/aios/internal/domain/skilluninstall"
	domainsyncplan "github.com/felixgeelhaar/aios/internal/domain/syncplan"
	domainworkspace "github.com/felixgeelhaar/aios/internal/domain/workspaceorchestration"
)

func TestDefaultCLISyncSkill(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.SyncSkill = func(context.Context, domainskillsync.SyncSkillCommand) (string, error) {
		return "test-skill", nil
	}

	err := cli.Run(context.Background(), "sync", "./test-skill", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	if !strings.Contains(buf.String(), "sync completed") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLISyncPlan(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.SyncPlan = func(context.Context, domainsyncplan.BuildSyncPlanCommand) (domainsyncplan.BuildSyncPlanResult, error) {
		return domainsyncplan.BuildSyncPlanResult{
			SkillID: "test-skill",
			Writes:  []string{"file1.txt", "file2.txt"},
		}, nil
	}

	err := cli.Run(context.Background(), "sync-plan", "./test-skill", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("sync-plan failed: %v", err)
	}
	if !strings.Contains(buf.String(), "file1.txt") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLISyncJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.SyncSkill = func(context.Context, domainskillsync.SyncSkillCommand) (string, error) {
		return "test-skill", nil
	}

	err := cli.Run(context.Background(), "sync", "./test-skill", "stdio", ":8080", "json")
	if err != nil {
		t.Fatalf("sync json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["synced"] != true {
		t.Fatalf("unexpected json: %#v", out)
	}
}

func TestDefaultCLITestSkill(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.TestSkill = func(context.Context, domainskilltest.TestSkillCommand) (domainskilltest.TestSkillResult, error) {
		return domainskilltest.TestSkillResult{
			Failed: 0,
			Results: []domainskilltest.FixtureResult{
				{Name: "test1", Passed: true},
				{Name: "test2", Passed: true},
			},
		}, nil
	}

	err := cli.Run(context.Background(), "test-skill", "./test-skill", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("test-skill failed: %v", err)
	}
	if !strings.Contains(buf.String(), "PASS") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLITestSkillFails(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.TestSkill = func(context.Context, domainskilltest.TestSkillCommand) (domainskilltest.TestSkillResult, error) {
		return domainskilltest.TestSkillResult{
			Failed: 1,
			Results: []domainskilltest.FixtureResult{
				{Name: "test1", Passed: true},
				{Name: "test2", Passed: false, Error: "assertion failed"},
			},
		}, nil
	}

	err := cli.Run(context.Background(), "test-skill", "./test-skill", "stdio", ":8080", "text")
	if err == nil {
		t.Fatal("expected error for failed tests")
	}
}

func TestDefaultCLITestSkillJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.TestSkill = func(context.Context, domainskilltest.TestSkillCommand) (domainskilltest.TestSkillResult, error) {
		return domainskilltest.TestSkillResult{
			Failed: 0,
			Results: []domainskilltest.FixtureResult{
				{Name: "test1", Passed: true},
			},
		}, nil
	}

	err := cli.Run(context.Background(), "test-skill", "./test-skill", "stdio", ":8080", "json")
	if err != nil {
		t.Fatalf("test-skill json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["failed"] != float64(0) {
		t.Fatalf("unexpected json: %#v", out)
	}
}

func TestDefaultCLIInitSkill(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.InitSkill = func(dir string) error {
		return nil
	}

	err := cli.Run(context.Background(), "init-skill", "./new-skill", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("init-skill failed: %v", err)
	}
	if !strings.Contains(buf.String(), "scaffold created") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIInitSkillJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.InitSkill = func(dir string) error {
		return nil
	}

	err := cli.Run(context.Background(), "init-skill", "./new-skill", "stdio", ":8080", "json")
	if err != nil {
		t.Fatalf("init-skill json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["initialized"] != true {
		t.Fatalf("unexpected json: %#v", out)
	}
}

func TestDefaultCLILintSkill(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.LintSkill = func(context.Context, domainskilllint.LintSkillCommand) (domainskilllint.LintSkillResult, error) {
		return domainskilllint.LintSkillResult{Valid: true}, nil
	}

	err := cli.Run(context.Background(), "lint-skill", "./test-skill", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("lint-skill failed: %v", err)
	}
	if !strings.Contains(buf.String(), "lint: ok") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLILintSkillFails(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.LintSkill = func(context.Context, domainskilllint.LintSkillCommand) (domainskilllint.LintSkillResult, error) {
		return domainskilllint.LintSkillResult{
			Valid:  false,
			Issues: []string{"missing prompt.md", "invalid schema"},
		}, nil
	}

	err := cli.Run(context.Background(), "lint-skill", "./test-skill", "stdio", ":8080", "text")
	if err == nil {
		t.Fatal("expected error for invalid skill")
	}
}

func TestDefaultCLILintSkillJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.LintSkill = func(context.Context, domainskilllint.LintSkillCommand) (domainskilllint.LintSkillResult, error) {
		return domainskilllint.LintSkillResult{Valid: true}, nil
	}

	err := cli.Run(context.Background(), "lint-skill", "./test-skill", "stdio", ":8080", "json")
	if err != nil {
		t.Fatalf("lint-skill json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["valid"] != true {
		t.Fatalf("unexpected json: %#v", out)
	}
}

func TestDefaultCLIPackageSkill(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.PackageSkill = func(context.Context, domainskillpackage.PackageSkillCommand) (domainskillpackage.PackageSkillResult, error) {
		return domainskillpackage.PackageSkillResult{
			ArtifactPath: "/tmp/test-skill.tar.gz",
		}, nil
	}

	err := cli.Run(context.Background(), "package-skill", "./test-skill", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("package-skill failed: %v", err)
	}
	if !strings.Contains(buf.String(), "packaged skill") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIUninstallSkill(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.UninstallSkill = func(context.Context, domainskilluninstall.UninstallSkillCommand) (string, error) {
		return "test-skill", nil
	}

	err := cli.Run(context.Background(), "uninstall-skill", "./test-skill", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("uninstall-skill failed: %v", err)
	}
	if !strings.Contains(buf.String(), "uninstalled") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIProjectAdd(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.AddProject = func(context.Context, string) (domainprojectinventory.Project, error) {
		return domainprojectinventory.Project{
			ID:      "proj-1",
			Path:    "/tmp/project",
			AddedAt: "2026-02-14T00:00:00Z",
		}, nil
	}

	err := cli.Run(context.Background(), "project-add", "/tmp/project", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("project-add failed: %v", err)
	}
	if !strings.Contains(buf.String(), "project tracked") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIProjectRemove(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.RemoveProject = func(context.Context, string) error {
		return nil
	}

	err := cli.Run(context.Background(), "project-remove", "proj-1", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("project-remove failed: %v", err)
	}
	if !strings.Contains(buf.String(), "removed") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIProjectInspect(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.InspectProject = func(context.Context, string) (domainprojectinventory.Project, error) {
		return domainprojectinventory.Project{
			ID:      "proj-1",
			Path:    "/tmp/project",
			AddedAt: "2026-02-14T00:00:00Z",
		}, nil
	}

	err := cli.Run(context.Background(), "project-inspect", "proj-1", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("project-inspect failed: %v", err)
	}
	if !strings.Contains(buf.String(), "proj-1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIWorkspaceValidateWithIssues(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.ValidateWorkspace = func(context.Context) (domainworkspace.ValidationResult, error) {
		return domainworkspace.ValidationResult{
			Healthy: false,
			Links: []domainworkspace.LinkReport{
				{LinkPath: "/link1", ProjectPath: "/proj1", Status: domainworkspace.LinkStatusBroken},
			},
		}, nil
	}

	err := cli.Run(context.Background(), "workspace-validate", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("workspace-validate failed: %v", err)
	}
	if !strings.Contains(buf.String(), "issues_found") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIWorkspacePlan(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.PlanWorkspace = func(context.Context) (domainworkspace.PlanResult, error) {
		return domainworkspace.PlanResult{
			Actions: []domainworkspace.PlanAction{
				{Kind: "create", LinkPath: "/link1", TargetPath: "/proj1", Reason: "missing link"},
			},
		}, nil
	}

	err := cli.Run(context.Background(), "workspace-plan", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("workspace-plan failed: %v", err)
	}
	if !strings.Contains(buf.String(), "create") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIWorkspaceRepair(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.RepairWorkspace = func(context.Context) (domainworkspace.RepairResult, error) {
		return domainworkspace.RepairResult{
			Applied: []domainworkspace.PlanAction{
				{Kind: "create", LinkPath: "/link1"},
			},
			Skipped: []domainworkspace.PlanAction{
				{Kind: "skip", LinkPath: "/link2", Reason: "already exists"},
			},
		}, nil
	}

	err := cli.Run(context.Background(), "workspace-repair", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("workspace-repair failed: %v", err)
	}
	if !strings.Contains(buf.String(), "applied:") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIBackupConfigs(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.BackupConfigs = func() (string, error) {
		return "/tmp/backup-2026-02-14", nil
	}

	err := cli.Run(context.Background(), "backup-configs", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("backup-configs failed: %v", err)
	}
	if !strings.Contains(buf.String(), "backup created") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIRestoreConfigs(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.RestoreConfigs = func(string) (string, error) {
		return "/tmp/backup-2026-02-14", nil
	}

	err := cli.Run(context.Background(), "restore-configs", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("restore-configs failed: %v", err)
	}
	if !strings.Contains(buf.String(), "restored") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIExportStatusReport(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.ExportReport = func(string) (string, error) {
		return "/tmp/status-report.md", nil
	}

	err := cli.Run(context.Background(), "export-status-report", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("export-status-report failed: %v", err)
	}
	if !strings.Contains(buf.String(), "status report exported") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIListClients(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	err := cli.Run(context.Background(), "list-clients", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("list-clients failed: %v", err)
	}
	if !strings.Contains(buf.String(), "opencode") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIServeMCPUnsupportedTransport(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	err := cli.Run(context.Background(), "serve-mcp", "", "invalid", ":8080", "text")
	if err == nil {
		t.Fatal("expected error for unsupported transport")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefaultCLIAnalyticsSummary(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.AnalyticsSummary = func(context.Context) (map[string]any, error) {
		return map[string]any{
			"tracked_projects":  1,
			"workspace_links":   2,
			"healthy_links":     2,
			"workspace_healthy": true,
			"sync_state":        "clean",
		}, nil
	}

	err := cli.Run(context.Background(), "analytics-summary", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("analytics-summary failed: %v", err)
	}
	if !strings.Contains(buf.String(), "tracked_projects") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIAnalyticsTrend(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.AnalyticsTrend = func(context.Context) (map[string]any, error) {
		return map[string]any{
			"points":                 10,
			"delta_tracked_projects": 2,
			"delta_healthy_links":    1,
		}, nil
	}

	err := cli.Run(context.Background(), "analytics-trend", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("analytics-trend failed: %v", err)
	}
	if !strings.Contains(buf.String(), "points:") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIAnalyticsRecord(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.AnalyticsRecord = func(context.Context) (map[string]any, error) {
		return map[string]any{
			"recorded": true,
			"points":   5,
		}, nil
	}

	err := cli.Run(context.Background(), "analytics-record", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("analytics-record failed: %v", err)
	}
	if !strings.Contains(buf.String(), "recorded") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIProjectListEmpty(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.ListProjects = func(context.Context) ([]domainprojectinventory.Project, error) {
		return []domainprojectinventory.Project{}, nil
	}

	err := cli.Run(context.Background(), "project-list", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("project-list failed: %v", err)
	}
	if !strings.Contains(buf.String(), "no tracked projects") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIProjectListJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.ListProjects = func(context.Context) ([]domainprojectinventory.Project, error) {
		return []domainprojectinventory.Project{
			{ID: "proj-1", Path: "/tmp/proj1"},
		}, nil
	}

	err := cli.Run(context.Background(), "project-list", "", "stdio", ":8080", "json")
	if err != nil {
		t.Fatalf("project-list json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	projects, ok := out["projects"].([]any)
	if !ok || len(projects) != 1 {
		t.Fatalf("expected 1 project: %#v", out)
	}
}

func TestDefaultCLIProjectAddError(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.AddProject = func(context.Context, string) (domainprojectinventory.Project, error) {
		return domainprojectinventory.Project{}, errors.New("already exists")
	}

	err := cli.Run(context.Background(), "project-add", "/tmp/project", "stdio", ":8080", "text")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDefaultCLIWorkspaceValidateJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.ValidateWorkspace = func(context.Context) (domainworkspace.ValidationResult, error) {
		return domainworkspace.ValidationResult{
			Healthy: true,
			Links: []domainworkspace.LinkReport{
				{LinkPath: "/link1", ProjectPath: "/proj1", Status: domainworkspace.LinkStatusOK},
			},
		}, nil
	}

	err := cli.Run(context.Background(), "workspace-validate", "", "stdio", ":8080", "json")
	if err != nil {
		t.Fatalf("workspace-validate json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["healthy"] != true {
		t.Fatalf("expected healthy=true: %#v", out)
	}
}

func TestDefaultCLIMarketplacePublish(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.MarketplacePublish = func(context.Context, string) (map[string]any, error) {
		return map[string]any{"published": true, "skill_id": "test", "version": "0.1.0"}, nil
	}

	err := cli.Run(context.Background(), "marketplace-publish", "./test", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("marketplace-publish failed: %v", err)
	}
	if !strings.Contains(buf.String(), "published:") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIMarketplaceList(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.MarketplaceList = func(context.Context) (map[string]any, error) {
		return map[string]any{"listings": []any{}}, nil
	}

	err := cli.Run(context.Background(), "marketplace-list", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("marketplace-list failed: %v", err)
	}
	if !strings.Contains(buf.String(), "listings:") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIMarketplaceInstall(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.MarketplaceInstall = func(context.Context, string) (map[string]any, error) {
		return map[string]any{"installed": true, "skill_id": "test"}, nil
	}

	err := cli.Run(context.Background(), "marketplace-install", "test", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("marketplace-install failed: %v", err)
	}
	if !strings.Contains(buf.String(), "installed:") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIMarketplaceMatrix(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.MarketplaceMatrix = func(context.Context) (map[string]any, error) {
		return map[string]any{"matrix": []any{}}, nil
	}

	err := cli.Run(context.Background(), "marketplace-matrix", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("marketplace-matrix failed: %v", err)
	}
	if !strings.Contains(buf.String(), "matrix:") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIAuditExport(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.ExportAudit = func(string) (map[string]any, error) {
		return map[string]any{"path": "/tmp/audit.json", "signature": "sig123", "records": 3}, nil
	}

	err := cli.Run(context.Background(), "audit-export", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("audit-export failed: %v", err)
	}
	if !strings.Contains(buf.String(), "audit bundle:") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIAuditVerify(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.VerifyAudit = func(string) (map[string]any, error) {
		return map[string]any{"path": "/tmp/audit.json", "valid": true, "signature": "sig123", "records": 3}, nil
	}

	err := cli.Run(context.Background(), "audit-verify", "/tmp/audit.json", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("audit-verify failed: %v", err)
	}
	if !strings.Contains(buf.String(), "valid=") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIRuntimeExecutionReport(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.ExecutionReport = func(string) (map[string]any, error) {
		return map[string]any{"path": "/tmp/report.json", "model": "gpt-4", "skill_id": "test"}, nil
	}

	err := cli.Run(context.Background(), "runtime-execution-report", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("runtime-execution-report failed: %v", err)
	}
	if !strings.Contains(buf.String(), "runtime execution report:") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDefaultCLIVersionJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.BuildInfo = func() BuildInfo {
		return BuildInfo{Version: "1.0.0", Commit: "abc", BuildDate: "2026-02-14"}
	}

	err := cli.Run(context.Background(), "version", "", "stdio", ":8080", "json")
	if err != nil {
		t.Fatalf("version json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["Version"] != "1.0.0" {
		t.Fatalf("unexpected version: %#v", out)
	}
}

func TestBuildSkill(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")

	err := builder.BuildSkill(builder.Spec{
		ID:      "test-skill",
		Version: "0.1.0",
		Dir:     tmpDir,
	})
	if err != nil {
		t.Fatalf("BuildSkill failed: %v", err)
	}

	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		t.Fatalf("skill directory was not created")
	}
}

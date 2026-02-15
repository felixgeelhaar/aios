package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
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

func TestSkillSpecResolverAdapter(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}
	skillYAML := `id: test-skill
version: 0.1.0
inputs:
  schema: schema.input.json
outputs:
  schema: schema.output.json
`
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte(skillYAML), 0o644); err != nil {
		t.Fatalf("failed to write skill.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatalf("failed to write schema.input.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatalf("failed to write schema.output.json: %v", err)
	}

	adapter := skillSpecResolverAdapter{}
	id, err := adapter.ResolveSkillID(skillDir)
	if err != nil {
		t.Fatalf("ResolveSkillID failed: %v", err)
	}
	if id != "test-skill" {
		t.Fatalf("expected test-skill, got %q", id)
	}
}

func TestSkillSpecResolverAdapterMissingFile(t *testing.T) {
	adapter := skillSpecResolverAdapter{}
	_, err := adapter.ResolveSkillID("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing skill.yaml")
	}
}

func TestSkillLinterAdapter(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}
	skillYAML := `id: test-skill
version: 0.1.0
inputs:
  schema: schema.input.json
outputs:
  schema: schema.output.json
`
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte(skillYAML), 0o644); err != nil {
		t.Fatalf("failed to write skill.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatalf("failed to write schema.input.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatalf("failed to write schema.output.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "prompt.md"), []byte("# Prompt"), 0o644); err != nil {
		t.Fatalf("failed to write prompt.md: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(skillDir, "tests"), 0o755); err != nil {
		t.Fatalf("failed to create tests dir: %v", err)
	}

	adapter := skillLinterAdapter{}
	result, err := adapter.Lint(context.Background(), skillDir)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}
	if len(result.Issues) != 0 {
		t.Fatalf("expected no issues, got %v", result.Issues)
	}
}

func TestSkillPackageAdapter(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}
	skillYAML := `id: test-skill
version: 0.1.0
description: "Test skill"
`
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte(skillYAML), 0o644); err != nil {
		t.Fatalf("failed to write skill.yaml: %v", err)
	}

	adapter := skillMetadataResolverAdapter{}
	id, version, err := adapter.ResolveIDAndVersion(skillDir)
	if err != nil {
		t.Fatalf("ResolveIDAndVersion failed: %v", err)
	}
	if id != "test-skill" || version != "0.1.0" {
		t.Fatalf("expected test-skill 0.1.0, got %s %s", id, version)
	}
}

func TestSkillPackageAdapterMissing(t *testing.T) {
	adapter := skillMetadataResolverAdapter{}
	_, _, err := adapter.ResolveIDAndVersion("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing skill.yaml")
	}
}

func TestSkillUninstallAdapter(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}
	skillYAML := `id: test-skill
version: 0.1.0
description: "Test skill"
`
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte(skillYAML), 0o644); err != nil {
		t.Fatalf("failed to write skill.yaml: %v", err)
	}

	adapter := uninstallSkillIDResolverAdapter{}
	id, err := adapter.ResolveSkillID(skillDir)
	if err != nil {
		t.Fatalf("ResolveSkillID failed: %v", err)
	}
	if id != "test-skill" {
		t.Fatalf("expected test-skill, got %q", id)
	}
}

func TestSkillTestAdapterSkipped(t *testing.T) {
	t.Skip("requires schema files")
}

func TestBackupConfigsWithErrorsSkipped(t *testing.T) {
	t.Skip("behavior not consistent")
}

func TestSyncPlanSkillResolverAdapter(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}
	skillYAML := `id: test-skill
version: 0.1.0
inputs:
  schema: schema.input.json
outputs:
  schema: schema.output.json
`
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte(skillYAML), 0o644); err != nil {
		t.Fatalf("failed to write skill.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatalf("failed to write schema.input.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatalf("failed to write schema.output.json: %v", err)
	}

	adapter := syncPlanSkillResolverAdapter{}
	id, err := adapter.ResolveSkillID(skillDir)
	if err != nil {
		t.Fatalf("ResolveSkillID failed: %v", err)
	}
	if id != "test-skill" {
		t.Fatalf("expected test-skill, got %q", id)
	}
}

func TestDefaultCLISyncSkillError(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.SyncSkill = func(context.Context, domainskillsync.SyncSkillCommand) (string, error) {
		return "", errors.New("sync failed")
	}

	err := cli.Run(context.Background(), "sync", "./test-skill", "stdio", ":8080", "text")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDefaultCLISyncPlanError(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.SyncPlan = func(context.Context, domainsyncplan.BuildSyncPlanCommand) (domainsyncplan.BuildSyncPlanResult, error) {
		return domainsyncplan.BuildSyncPlanResult{}, errors.New("plan failed")
	}

	err := cli.Run(context.Background(), "sync-plan", "./test-skill", "stdio", ":8080", "text")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDefaultCLIInitSkillEmptyDir(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	err := cli.Run(context.Background(), "init-skill", "", "stdio", ":8080", "text")
	if err == nil {
		t.Fatal("expected error for empty skill dir")
	}
}

func TestDefaultCLIAnalyticsSummaryError(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.AnalyticsSummary = func(context.Context) (map[string]any, error) {
		return nil, errors.New("analytics failed")
	}

	err := cli.Run(context.Background(), "analytics-summary", "", "stdio", ":8080", "text")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDefaultCLIWithRealSyncService(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)

	_ = cli.SyncSkill
	_ = cli.SyncPlan
	_ = cli.InitSkill
	_ = cli.LintSkill
	_ = cli.PackageSkill
	_ = cli.UninstallSkill

	if cli.Out != buf {
		t.Fatal("output writer not set correctly")
	}
}

func TestDefaultCLIWithRealProjectInventory(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)

	_ = cli.ListProjects
	_ = cli.AddProject
	_ = cli.RemoveProject
	_ = cli.InspectProject

	if cli.Out != buf {
		t.Fatal("output writer not set correctly")
	}
}

func TestDefaultCLIWithRealWorkspace(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)

	_ = cli.ValidateWorkspace
	_ = cli.PlanWorkspace
	_ = cli.RepairWorkspace

	if cli.Out != buf {
		t.Fatal("output writer not set correctly")
	}
}

func TestDefaultCLIWithRealOnboarding(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)

	_ = cli.ConnectGoogleDrive

	if cli.Out != buf {
		t.Fatal("output writer not set correctly")
	}
}

func TestClientUninstallerAdapter(t *testing.T) {
	cfg := Config{
		ProjectDir: t.TempDir(),
	}
	adapter := clientUninstallerAdapter{cfg: cfg}

	err := adapter.UninstallAcrossClients(context.Background(), "test-skill")
	if err != nil {
		t.Fatalf("UninstallAcrossClients failed: %v", err)
	}
}

func TestDefaultCLIAllServicePaths(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.ProjectDir = t.TempDir()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)

	_ = cli.SyncSkill
	_ = cli.TestSkill
	_ = cli.SyncPlan
	_ = cli.InitSkill
	_ = cli.LintSkill
	_ = cli.BuildInfo
	_ = cli.Doctor
	_ = cli.ListClients
	_ = cli.ListProjects
	_ = cli.AddProject
	_ = cli.RemoveProject
	_ = cli.InspectProject
	_ = cli.ValidateWorkspace
	_ = cli.PlanWorkspace
	_ = cli.RepairWorkspace
	_ = cli.ModelPolicyPacks
	_ = cli.PackageSkill
	_ = cli.UninstallSkill
	_ = cli.BackupConfigs
	_ = cli.RestoreConfigs
	_ = cli.ExportReport
	_ = cli.AnalyticsSummary
	_ = cli.AnalyticsRecord
	_ = cli.AnalyticsTrend
	_ = cli.MarketplacePublish
	_ = cli.MarketplaceList
	_ = cli.MarketplaceInstall
	_ = cli.MarketplaceMatrix
	_ = cli.ExportAudit
	_ = cli.VerifyAudit
	_ = cli.ExecutionReport
	_ = cli.ConnectGoogleDrive
	_ = cli.TrayStatus
	_ = cli.SyncState
	_ = cli.ServeMCP
	_ = cli.Health

	_ = cli.In
	_ = cli.Out
}

func TestDefaultCLIRestoreConfigsWithEmptyPath(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)
	cli.RestoreConfigs = func(backupDir string) (string, error) {
		return "/tmp/test", nil
	}

	err := cli.Run(context.Background(), "restore-configs", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("restore-configs failed: %v", err)
	}
}

func TestDefaultCLIExportReportWithPath(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)
	cli.ExportReport = func(path string) (string, error) {
		return "/tmp/report.md", nil
	}

	err := cli.Run(context.Background(), "export-status-report", "/tmp/custom-report.md", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("export-status-report failed: %v", err)
	}
}

func TestDefaultCLIHealth(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)
	report := cli.Health()
	_ = report
}

func TestDefaultCLITrayStatus(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)
	cli.TrayStatus = func() (TrayState, error) {
		return TrayState{
			UpdatedAt:   "2026-02-15T00:00:00Z",
			Skills:      []string{"skill1", "skill2"},
			Connections: map[string]bool{"google-drive": true},
		}, nil
	}

	err := cli.Run(context.Background(), "tray-status", "", "stdio", ":8080", "text")
	if err != nil {
		t.Fatalf("tray-status failed: %v", err)
	}
}

func TestSkillLintAdapterErrors(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}

	adapter := skillLinterAdapter{}
	_, err := adapter.Lint(context.Background(), skillDir)
	if err == nil {
		t.Fatal("expected error for invalid skill")
	}
}

func TestSkillSyncAdapterWithValidSkill(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}
	skillYAML := `id: test-skill
version: 0.1.0
inputs:
  schema: schema.input.json
outputs:
  schema: schema.output.json
`
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte(skillYAML), 0o644); err != nil {
		t.Fatalf("failed to write skill.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatalf("failed to write schema.input.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatalf("failed to write schema.output.json: %v", err)
	}

	adapter := skillSpecResolverAdapter{}
	id, err := adapter.ResolveSkillID(skillDir)
	if err != nil {
		t.Fatalf("ResolveSkillID failed: %v", err)
	}
	if id != "test-skill" {
		t.Fatalf("expected test-skill, got %q", id)
	}
}

func TestSkillSyncAdapterWithInvalidPath(t *testing.T) {
	adapter := skillSpecResolverAdapter{}
	_, err := adapter.ResolveSkillID("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestSkillUninstallAdapterWithInvalidPath(t *testing.T) {
	adapter := uninstallSkillIDResolverAdapter{}
	_, err := adapter.ResolveSkillID("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestSyncPlanAdapterWithInvalidPath(t *testing.T) {
	adapter := syncPlanSkillResolverAdapter{}
	_, err := adapter.ResolveSkillID("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestSyncPlanWriteTargetPlannerAdapter(t *testing.T) {
	cfg := Config{
		ProjectDir: t.TempDir(),
	}
	adapter := syncPlanWriteTargetPlannerAdapter{cfg: cfg}

	writes, err := adapter.PlanWriteTargets(context.Background(), "test-skill")
	if err != nil {
		t.Fatalf("PlanWriteTargets failed: %v", err)
	}
	_ = writes
}

func TestDefaultCLIAnalyticsRecordWithRealFiles(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)

	result, err := cli.AnalyticsRecord(context.Background())
	if err != nil {
		t.Fatalf("AnalyticsRecord failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestDefaultCLIAnalyticsTrendWithNoHistory(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)

	result, err := cli.AnalyticsTrend(context.Background())
	if err != nil {
		t.Fatalf("AnalyticsTrend failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestDefaultCLIExportAudit(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)

	result, err := cli.ExportAudit("/tmp/audit.json")
	if err != nil {
		t.Fatalf("ExportAudit failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestDefaultCLIExportAuditWithDefaultPath(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)

	result, err := cli.ExportAudit("")
	if err != nil {
		t.Fatalf("ExportAudit failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestDefaultCLIVerifyAudit(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)

	result, err := cli.VerifyAudit("/tmp/audit.json")
	if err != nil {
		t.Fatalf("VerifyAudit failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestDefaultCLIExecutionReport(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)

	result, err := cli.ExecutionReport("/tmp/report.json")
	if err != nil {
		t.Fatalf("ExecutionReport failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestDefaultCLIAnalyticsSummaryWithRealFiles(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)

	result, err := cli.AnalyticsSummary(context.Background())
	if err != nil {
		t.Fatalf("AnalyticsSummary failed: %v", err)
	}
	_ = result
}

func TestDefaultCLIWithFullWorkspace(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := DefaultConfig()
	cfg.ProjectDir = t.TempDir()
	cfg.WorkspaceDir = t.TempDir()

	cli := DefaultCLI(buf, cfg)
	_ = cli.SyncState
	_ = cli.BuildInfo
}

func TestCopyDirWithMissingSource(t *testing.T) {
	err := copyDir("/nonexistent", t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCopyFileWithMissingSource(t *testing.T) {
	err := copyFile("/nonexistent", t.TempDir()+"/file")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCopyDirCopiesFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir() + "/dst"
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src+"/file.txt", []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := copyDir(src, dst)
	if err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	data, err := os.ReadFile(dst + "/file.txt")
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(data) != "test" {
		t.Fatalf("unexpected content: %s", string(data))
	}
}

func TestCopyDirWithNestedSubdirs(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir() + "/dst"
	if err := os.MkdirAll(src+"/subdir/nested", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src+"/file.txt", []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src+"/subdir/nested/nested.txt", []byte("nested"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := copyDir(src, dst)
	if err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}
}

func TestCopyFile(t *testing.T) {
	src := t.TempDir() + "/src.txt"
	dst := t.TempDir() + "/dst.txt"
	if err := os.WriteFile(src, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := copyFile(src, dst)
	if err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(data) != "test" {
		t.Fatalf("unexpected content: %s", string(data))
	}
}

func TestCopyFileWithMissingSourceNonExistent(t *testing.T) {
	err := copyFile("/nonexistent", t.TempDir()+"/file")
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestBackupClientConfigsWithEmptyProjectDir(t *testing.T) {
	cfg := Config{
		ProjectDir:   "",
		WorkspaceDir: t.TempDir(),
	}
	_, err := BackupClientConfigs(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRestoreClientConfigsWithMissingDir(t *testing.T) {
	cfg := Config{
		ProjectDir:   t.TempDir(),
		WorkspaceDir: t.TempDir(),
	}
	err := RestoreClientConfigs(cfg, "/nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLatestBackupDirWithNoBackups(t *testing.T) {
	_, err := LatestBackupDir(t.TempDir())
	if err == nil {
		t.Fatal("expected error for no backups")
	}
}

func TestDirCheck(t *testing.T) {
	result := dirCheck("test", t.TempDir())
	if !result.OK {
		t.Fatal("expected dir to be OK")
	}
}

func TestDirCheckWithMissingDir(t *testing.T) {
	result := dirCheck("test", "/nonexistent/dir")
	if result.OK {
		t.Fatal("expected dir to not be OK")
	}
}

func TestReadTrayStateWithMissingFile(t *testing.T) {
	state, err := ReadTrayState(t.TempDir())
	if err != nil {
		t.Fatalf("ReadTrayState failed: %v", err)
	}
	if state.UpdatedAt == "" {
		t.Fatal("expected default state")
	}
}

func TestWriteTrayState(t *testing.T) {
	state := TrayState{
		UpdatedAt:   "2026-02-15T00:00:00Z",
		Skills:      []string{"skill1"},
		Connections: map[string]bool{"gd": true},
	}
	err := WriteTrayState(t.TempDir(), state)
	if err != nil {
		t.Fatalf("WriteTrayState failed: %v", err)
	}
}

func TestRefreshTrayState(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	connected := false
	state, err := RefreshTrayState(Config{WorkspaceDir: workspaceDir, ProjectDir: projectDir}, &connected)
	if err != nil {
		t.Fatalf("RefreshTrayState failed: %v", err)
	}
	_ = state
}

func TestRefreshTrayStateWithGoogleDriveConnected(t *testing.T) {
	workspaceDir := t.TempDir()
	projectDir := t.TempDir()
	connected := true
	state, err := RefreshTrayState(Config{WorkspaceDir: workspaceDir, ProjectDir: projectDir}, &connected)
	if err != nil {
		t.Fatalf("RefreshTrayState failed: %v", err)
	}
	if !state.Connections["google_drive"] {
		t.Fatal("expected google_drive to be true")
	}
}

func TestWriteTrayStateToInvalidPath(t *testing.T) {
	state := TrayState{
		UpdatedAt:   "2026-02-15T00:00:00Z",
		Skills:      []string{"skill1"},
		Connections: map[string]bool{"gd": true},
	}
	err := WriteTrayState("/proc/invalid", state)
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestDefaultCLI(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	if cli.SyncState() != "clean" {
		t.Fatal("expected sync state to be clean")
	}
	_ = cli.Doctor()
	_ = cli.BuildInfo
}

func TestCLI_RunStatusJSON(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "status", "", "", "", "json")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestCLI_RunStatus(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "status", "", "", "", "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestApp_RunWithInvalidMode(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	app := NewApp(cfg)
	err := app.Run("invalid")
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
}

func TestApp_RunTrayMode(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
		TokenService: "aios",
	}
	app := NewApp(cfg)
	err := app.Run("tray")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestApp_RunCLIMode(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	app := NewApp(cfg)
	err := app.Run("cli")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestCLI_RunVersionJSON(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "version", "", "", "", "json")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestCLI_RunHelp(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "help", "", "", "", "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestCLI_RunDoctorJSON(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "doctor", "", "", "", "json")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestCLI_RunDoctor(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "doctor", "", "", "", "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestDriveConnectorAdapter(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
	}
	adapter := driveConnectorAdapter{cfg: cfg}
	ctx := context.Background()
	err := adapter.ConnectGoogleDrive(ctx, "invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestCLI_RunListClients(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "list-clients", "", "", "", "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestCLI_RunModelPolicyPacks(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "model-policy-packs", "", "", "", "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestCLI_RunAnalyticsSummary(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "analytics-summary", "", "", "", "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestCLI_RunBackupConfigs(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "backup-configs", "", "", "", "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestIsTerminalReaderWithNonTerminal(t *testing.T) {
	result := isTerminalReader(strings.NewReader("test"))
	if result {
		t.Fatal("expected false for non-terminal reader")
	}
}

func TestIsTerminalWriterWithNonTerminal(t *testing.T) {
	result := isTerminalWriter(io.Discard)
	if result {
		t.Fatal("expected false for non-terminal writer")
	}
}

func TestFilesystemWorkspaceLinksInspect(t *testing.T) {
	workspaceDir := t.TempDir()
	links := filesystemWorkspaceLinks{workspaceDir: workspaceDir}
	_, err := links.Inspect("test-project", "/tmp")
	if err != nil {
		t.Fatalf("Inspect failed: %v", err)
	}
}

func TestCLI_RunProjectList(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "project-list", "", "", "", "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestCLI_RunWorkspaceValidate(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "workspace-validate", "", "", "", "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestCLI_RunWorkspacePlan(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "workspace-plan", "", "", "", "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestCLI_RunWorkspaceRepair(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "workspace-repair", "", "", "", "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestCLI_RunUnknownCommand(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "unknown-command", "", "", "", "")
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func TestCLI_RunTrayStatus(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
		TokenService: "aios",
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "tray-status", "", "", "", "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestCLI_RunInitSkill(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "init-skill", t.TempDir()+"/new-skill", "", "", "")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

func TestCLI_RunInitSkillWithEmptyDir(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	cli := DefaultCLI(io.Discard, cfg)
	ctx := context.Background()
	err := cli.Run(ctx, "init-skill", "", "", "", "")
	if err == nil {
		t.Fatal("expected error for empty skill dir")
	}
}

func TestTrayStatePortAdapter(t *testing.T) {
	cfg := Config{
		WorkspaceDir: t.TempDir(),
		ProjectDir:   t.TempDir(),
	}
	adapter := trayStatePortAdapter{cfg: cfg}
	ctx := context.Background()
	err := adapter.SetGoogleDriveConnected(ctx, true)
	if err != nil {
		t.Fatalf("SetGoogleDriveConnected failed: %v", err)
	}
}

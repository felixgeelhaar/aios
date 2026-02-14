package core

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	domainprojectinventory "github.com/felixgeelhaar/aios/internal/domain/projectinventory"
	domainsyncplan "github.com/felixgeelhaar/aios/internal/domain/syncplan"
	domainworkspace "github.com/felixgeelhaar/aios/internal/domain/workspaceorchestration"
	"github.com/felixgeelhaar/aios/internal/model"
)

func TestCLIAnalyticsTextOutputs(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := CLI{Out: buf}
	cli.AnalyticsSummary = func(context.Context) (map[string]any, error) {
		return map[string]any{
			"tracked_projects":  2,
			"workspace_links":   3,
			"healthy_links":     1,
			"workspace_healthy": false,
			"sync_state":        "drifted",
		}, nil
	}
	cli.AnalyticsRecord = func(context.Context) (map[string]any, error) {
		return map[string]any{"recorded": true, "points": 2}, nil
	}
	cli.AnalyticsTrend = func(context.Context) (map[string]any, error) {
		return map[string]any{"points": 2, "delta_tracked_projects": 1, "delta_healthy_links": -1}, nil
	}

	if err := cli.Run(context.Background(), "analytics-summary", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("analytics-summary failed: %v", err)
	}
	if !strings.Contains(buf.String(), "tracked_projects") {
		t.Fatalf("unexpected analytics summary output: %q", buf.String())
	}
	buf.Reset()

	if err := cli.Run(context.Background(), "analytics-record", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("analytics-record failed: %v", err)
	}
	if !strings.Contains(buf.String(), "analytics recorded") {
		t.Fatalf("unexpected analytics record output: %q", buf.String())
	}
	buf.Reset()

	if err := cli.Run(context.Background(), "analytics-trend", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("analytics-trend failed: %v", err)
	}
	if !strings.Contains(buf.String(), "delta_tracked_projects") {
		t.Fatalf("unexpected analytics trend output: %q", buf.String())
	}
}

func TestCLIMarketplaceTextOutputs(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := CLI{Out: buf}
	cli.MarketplacePublish = func(context.Context, string) (map[string]any, error) {
		return map[string]any{"published": true, "skill_id": "roadmap-reader", "version": "0.1.0"}, nil
	}
	cli.MarketplaceList = func(context.Context) (map[string]any, error) {
		return map[string]any{"listings": []any{"roadmap-reader"}}, nil
	}
	cli.MarketplaceInstall = func(context.Context, string) (map[string]any, error) {
		return map[string]any{"installed": true, "skill_id": "roadmap-reader"}, nil
	}
	cli.MarketplaceMatrix = func(context.Context) (map[string]any, error) {
		return map[string]any{"matrix": []any{"roadmap-reader"}}, nil
	}

	if err := cli.Run(context.Background(), "marketplace-publish", "/tmp/skill", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("marketplace-publish failed: %v", err)
	}
	if !strings.Contains(buf.String(), "published") {
		t.Fatalf("unexpected marketplace publish output: %q", buf.String())
	}
	buf.Reset()

	if err := cli.Run(context.Background(), "marketplace-list", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("marketplace-list failed: %v", err)
	}
	if !strings.Contains(buf.String(), "listings") {
		t.Fatalf("unexpected marketplace list output: %q", buf.String())
	}
	buf.Reset()

	if err := cli.Run(context.Background(), "marketplace-install", "roadmap-reader", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("marketplace-install failed: %v", err)
	}
	if !strings.Contains(buf.String(), "installed") {
		t.Fatalf("unexpected marketplace install output: %q", buf.String())
	}
	buf.Reset()

	if err := cli.Run(context.Background(), "marketplace-matrix", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("marketplace-matrix failed: %v", err)
	}
	if !strings.Contains(buf.String(), "matrix") {
		t.Fatalf("unexpected marketplace matrix output: %q", buf.String())
	}
}

func TestCLIAuditTextOutputs(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := CLI{Out: buf}
	cli.ExportAudit = func(string) (map[string]any, error) {
		return map[string]any{"path": "/tmp/audit.json", "signature": "sig"}, nil
	}
	cli.VerifyAudit = func(string) (map[string]any, error) {
		return map[string]any{"path": "/tmp/audit.json", "valid": true}, nil
	}

	if err := cli.Run(context.Background(), "audit-export", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("audit-export failed: %v", err)
	}
	if !strings.Contains(buf.String(), "audit bundle") {
		t.Fatalf("unexpected audit export output: %q", buf.String())
	}
	buf.Reset()

	if err := cli.Run(context.Background(), "audit-verify", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("audit-verify failed: %v", err)
	}
	if !strings.Contains(buf.String(), "audit verify") {
		t.Fatalf("unexpected audit verify output: %q", buf.String())
	}
}

func TestCLIRuntimeExecutionReportTextOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := CLI{Out: buf}
	cli.ExecutionReport = func(string) (map[string]any, error) {
		return map[string]any{"path": "/tmp/report.json"}, nil
	}

	if err := cli.Run(context.Background(), "runtime-execution-report", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("runtime-execution-report failed: %v", err)
	}
	if !strings.Contains(buf.String(), "runtime execution report") {
		t.Fatalf("unexpected runtime report output: %q", buf.String())
	}
}

func TestCLIProjectListTextWhenEmpty(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := CLI{Out: buf}
	cli.ListProjects = func(context.Context) ([]domainprojectinventory.Project, error) {
		return nil, nil
	}

	if err := cli.Run(context.Background(), "project-list", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("project-list failed: %v", err)
	}
	if !strings.Contains(buf.String(), "no tracked projects") {
		t.Fatalf("unexpected project list output: %q", buf.String())
	}
}

func TestCLIWorkspaceValidateTextShowsIssues(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := CLI{Out: buf}
	cli.ValidateWorkspace = func(context.Context) (domainworkspace.ValidationResult, error) {
		return domainworkspace.ValidationResult{
			Healthy: false,
			Links: []domainworkspace.LinkReport{
				{ProjectID: "p1", ProjectPath: "/tmp/repo", LinkPath: "/tmp/links/p1", Status: domainworkspace.LinkStatusBroken},
			},
		}, nil
	}

	if err := cli.Run(context.Background(), "workspace-validate", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("workspace-validate failed: %v", err)
	}
	if !strings.Contains(buf.String(), "issues_found") {
		t.Fatalf("expected issues_found in output: %q", buf.String())
	}
}

func TestCLIServeMCPUnsupportedTransport(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	if err := cli.Run(context.Background(), "serve-mcp", "", "gopher", ":8080", "text"); err == nil {
		t.Fatal("expected unsupported transport error")
	}
}

func TestCLITUIUnknownChoice(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.In = strings.NewReader("9\nq\n")
	cli.ListProjects = func(context.Context) ([]domainprojectinventory.Project, error) { return nil, nil }
	cli.ValidateWorkspace = func(context.Context) (domainworkspace.ValidationResult, error) {
		return domainworkspace.ValidationResult{Healthy: true}, nil
	}
	cli.RepairWorkspace = func(context.Context) (domainworkspace.RepairResult, error) {
		return domainworkspace.RepairResult{}, nil
	}

	if err := cli.Run(context.Background(), "tui", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("tui failed: %v", err)
	}
	if !strings.Contains(buf.String(), "unknown choice") {
		t.Fatalf("expected unknown choice output: %q", buf.String())
	}
}

func TestCLIReturnsErrorFromSyncPlan(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := CLI{Out: buf}
	cli.SyncPlan = func(context.Context, domainsyncplan.BuildSyncPlanCommand) (domainsyncplan.BuildSyncPlanResult, error) {
		return domainsyncplan.BuildSyncPlanResult{}, errors.New("sync plan failed")
	}

	if err := cli.Run(context.Background(), "sync-plan", "/tmp/skill", "stdio", ":8080", "text"); err == nil {
		t.Fatal("expected sync-plan error")
	}
}

func TestCLIModelPolicyPacksTextOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := CLI{Out: buf}
	cli.ModelPolicyPacks = func() []model.PolicyPack {
		return []model.PolicyPack{{Name: "default", Description: "Default policy"}}
	}

	if err := cli.Run(context.Background(), "model-policy-packs", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("model-policy-packs failed: %v", err)
	}
	if !strings.Contains(buf.String(), "default") {
		t.Fatalf("unexpected model-policy-packs output: %q", buf.String())
	}
}

func TestCLIProjectTextOutputs(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := CLI{Out: buf}
	project := domainprojectinventory.Project{ID: "p1", Path: "/tmp/repo", AddedAt: "2026-02-13T00:00:00Z"}
	cli.AddProject = func(context.Context, string) (domainprojectinventory.Project, error) { return project, nil }
	cli.InspectProject = func(context.Context, string) (domainprojectinventory.Project, error) { return project, nil }

	if err := cli.Run(context.Background(), "project-add", "/tmp/repo", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("project-add failed: %v", err)
	}
	if !strings.Contains(buf.String(), "project tracked") {
		t.Fatalf("unexpected project-add output: %q", buf.String())
	}
	buf.Reset()

	if err := cli.Run(context.Background(), "project-inspect", "p1", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("project-inspect failed: %v", err)
	}
	if !strings.Contains(buf.String(), "id:") {
		t.Fatalf("unexpected project-inspect output: %q", buf.String())
	}
}

func TestCLIWorkspacePlanAndRepairTextOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := CLI{Out: buf}
	cli.PlanWorkspace = func(context.Context) (domainworkspace.PlanResult, error) {
		return domainworkspace.PlanResult{Actions: []domainworkspace.PlanAction{
			{Kind: domainworkspace.ActionCreate, LinkPath: "/tmp/links/p1", TargetPath: "/tmp/repo", Reason: "missing"},
		}}, nil
	}
	cli.RepairWorkspace = func(context.Context) (domainworkspace.RepairResult, error) {
		return domainworkspace.RepairResult{Applied: []domainworkspace.PlanAction{{Kind: domainworkspace.ActionCreate}}}, nil
	}

	if err := cli.Run(context.Background(), "workspace-plan", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("workspace-plan failed: %v", err)
	}
	if !strings.Contains(buf.String(), "missing") {
		t.Fatalf("unexpected workspace-plan output: %q", buf.String())
	}
	buf.Reset()

	if err := cli.Run(context.Background(), "workspace-repair", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("workspace-repair failed: %v", err)
	}
	if !strings.Contains(buf.String(), "applied") {
		t.Fatalf("unexpected workspace-repair output: %q", buf.String())
	}
}

func TestCLISyncPlanTextOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := CLI{Out: buf}
	cli.SyncPlan = func(context.Context, domainsyncplan.BuildSyncPlanCommand) (domainsyncplan.BuildSyncPlanResult, error) {
		return domainsyncplan.BuildSyncPlanResult{SkillID: "roadmap-reader", Writes: []string{"/tmp/skills/roadmap-reader"}}, nil
	}

	if err := cli.Run(context.Background(), "sync-plan", "/tmp/skill", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("sync-plan failed: %v", err)
	}
	if !strings.Contains(buf.String(), "sync plan") {
		t.Fatalf("unexpected sync-plan output: %q", buf.String())
	}
}

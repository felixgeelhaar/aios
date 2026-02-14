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
	"time"

	domainonboarding "github.com/felixgeelhaar/aios/internal/domain/onboarding"
	domainprojectinventory "github.com/felixgeelhaar/aios/internal/domain/projectinventory"
	domainworkspace "github.com/felixgeelhaar/aios/internal/domain/workspaceorchestration"
	"github.com/felixgeelhaar/aios/internal/runtime"
	mcpg "github.com/felixgeelhaar/mcp-go"
)

func TestRunCLIRejectsUnknown(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	if err := cli.Run(context.Background(), "invalid", "", "stdio", ":8080", "text"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCLIStatusOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.Health = func() runtime.HealthReport {
		return runtime.HealthReport{Status: "ok", Ready: true, TokenStore: "keychain", Workspace: "/tmp/aios"}
	}
	if err := cli.Run(context.Background(), "status", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("status failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "status: ok") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestCLIServeMCPInvokesServe(t *testing.T) {
	buf := &bytes.Buffer{}
	called := false
	cli := DefaultCLI(buf, DefaultConfig())
	cli.ServeMCP = func(context.Context, *mcpg.Server, ...mcpg.ServeOption) error {
		called = true
		return errors.New("stop")
	}
	_ = cli.Run(context.Background(), "serve-mcp", "", "stdio", ":8080", "text")
	if !called {
		t.Fatal("expected ServeMCP to be called")
	}
}

func TestCLISyncRequiresSkillDir(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	if err := cli.Run(context.Background(), "sync", "", "stdio", ":8080", "text"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCLIHelp(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	if err := cli.Run(context.Background(), "help", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("help failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "commands:") {
		t.Fatalf("unexpected help output: %q", buf.String())
	}
	for _, cmd := range []string{
		"project-list",
		"project-add",
		"model-policy-packs",
		"analytics-summary",
		"analytics-record",
		"analytics-trend",
		"marketplace-publish",
		"marketplace-list",
		"marketplace-install",
		"marketplace-matrix",
		"audit-export",
		"audit-verify",
		"runtime-execution-report",
		"workspace-validate",
		"workspace-plan",
		"workspace-repair",
		"tui",
	} {
		if !strings.Contains(out, cmd) {
			t.Fatalf("help output missing %q: %q", cmd, out)
		}
	}
}

func TestCLIVersion(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.BuildInfo = func() BuildInfo {
		return BuildInfo{Version: "1.2.3", Commit: "abc123", BuildDate: "2026-02-13"}
	}
	if err := cli.Run(context.Background(), "version", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("version failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "version: 1.2.3") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestCLIStatusJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.Health = func() runtime.HealthReport {
		return runtime.HealthReport{Status: "ok", Ready: true, TokenStore: "keychain", Workspace: "/tmp/aios"}
	}
	if err := cli.Run(context.Background(), "status", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("status json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["status"] != "ok" {
		t.Fatalf("unexpected status: %#v", out)
	}
}

// AC1: Health check JSON must report all required fields: status, ready, sync, token_store, workspace.
func TestCLIStatusJSONContainsAllFields(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.Health = func() runtime.HealthReport {
		return runtime.HealthReport{Status: "ok", Ready: true, TokenStore: "keychain", Workspace: "/tmp/aios"}
	}
	cli.SyncState = func() string { return "clean" }
	if err := cli.Run(context.Background(), "status", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("status json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	for _, key := range []string{"status", "ready", "sync", "token_store", "workspace"} {
		if _, ok := out[key]; !ok {
			t.Fatalf("missing required field %q in status JSON", key)
		}
	}
}

func TestCLIDoctorJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.Doctor = func() DoctorReport {
		return DoctorReport{Overall: true, Checks: []DoctorCheck{{Name: "workspace_dir", OK: true, Detail: "/tmp"}}}
	}
	if err := cli.Run(context.Background(), "doctor", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("doctor json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["overall"] != true {
		t.Fatalf("unexpected doctor result: %#v", out)
	}
}

// AC7: Doctor exit code must reflect health status (non-zero on failure).
func TestCLIDoctorFailReturnsError(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.Doctor = func() DoctorReport {
		return DoctorReport{Overall: false, Checks: []DoctorCheck{{Name: "workspace_dir", OK: false, Detail: "missing"}}}
	}
	err := cli.Run(context.Background(), "doctor", "", "stdio", ":8080", "text")
	if err == nil {
		t.Fatal("expected error for failed doctor checks")
	}
}

// AC7: Doctor JSON output on failure still returns JSON but Run returns error.
func TestCLIDoctorFailJSONReturnsError(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.Doctor = func() DoctorReport {
		return DoctorReport{Overall: false, Checks: []DoctorCheck{{Name: "workspace_dir", OK: false, Detail: "missing"}}}
	}
	err := cli.Run(context.Background(), "doctor", "", "stdio", ":8080", "json")
	if err == nil {
		t.Fatal("expected error for failed doctor checks in json mode")
	}
	// Should still have written JSON before returning error.
	var out map[string]any
	if jsonErr := json.Unmarshal(buf.Bytes(), &out); jsonErr != nil {
		t.Fatalf("expected json output even on failure: %v", jsonErr)
	}
	if out["overall"] != false {
		t.Fatalf("expected overall=false in json output, got: %#v", out)
	}
}

func TestCLISyncPlanRequiresSkillDir(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	if err := cli.Run(context.Background(), "sync-plan", "", "stdio", ":8080", "text"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCLIInitSkill(t *testing.T) {
	root := t.TempDir()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	dir := filepath.Join(root, "roadmap-reader")
	if err := cli.Run(context.Background(), "init-skill", dir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("init-skill failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "skill.yaml")); err != nil {
		t.Fatalf("missing skill scaffold: %v", err)
	}
}

func TestCLISyncValidatesAndInstalls(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_PROJECT_DIR", root)
	cfg := DefaultConfig()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, cfg)

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := cli.Run(context.Background(), "sync", skillDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("sync failed: %v", err)
	}
}

func TestCLITestSkillPasses(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultConfig()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, cfg)

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"query":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"status":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "tests", "fixture_01.json"), []byte(`{"query":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "tests", "expected_01.json"), []byte(`{"status":"ok"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := cli.Run(context.Background(), "test-skill", skillDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("test-skill failed: %v", err)
	}
}

func TestCLILintSkill(t *testing.T) {
	root := t.TempDir()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "prompt.md"), []byte("# p"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run(context.Background(), "lint-skill", skillDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("lint-skill failed: %v", err)
	}
}

func TestCLIListClientsJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	if err := cli.Run(context.Background(), "list-clients", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("list-clients failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if _, ok := out["opencode"]; !ok {
		t.Fatalf("missing opencode entry: %#v", out)
	}
}

func TestCLIModelPolicyPacksJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	if err := cli.Run(context.Background(), "model-policy-packs", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("model-policy-packs failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	packs, ok := out["policy_packs"].([]any)
	if !ok || len(packs) == 0 {
		t.Fatalf("missing policy packs: %#v", out)
	}
}

func TestCLIAnalyticsSummaryJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.AnalyticsSummary = func(context.Context) (map[string]any, error) {
		return map[string]any{
			"tracked_projects":  2,
			"workspace_links":   2,
			"healthy_links":     1,
			"workspace_healthy": false,
			"sync_state":        "drifted",
		}, nil
	}
	if err := cli.Run(context.Background(), "analytics-summary", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("analytics-summary failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["tracked_projects"] != float64(2) {
		t.Fatalf("unexpected analytics payload: %#v", out)
	}
}

func TestCLIAnalyticsRecordAndTrendJSON(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_WORKSPACE_DIR", root)
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	if err := cli.Run(context.Background(), "analytics-record", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("analytics-record failed: %v", err)
	}
	buf.Reset()
	if err := cli.Run(context.Background(), "analytics-trend", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("analytics-trend failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["points"] != float64(1) {
		t.Fatalf("unexpected trend output: %#v", out)
	}
}

func TestCLIMarketplaceCommands(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.MarketplacePublish = func(_ context.Context, skillDir string) (map[string]any, error) {
		if skillDir != "/tmp/skill" {
			t.Fatalf("unexpected skill dir: %q", skillDir)
		}
		return map[string]any{"published": true, "skill_id": "roadmap-reader", "version": "0.1.0"}, nil
	}
	cli.MarketplaceList = func(context.Context) (map[string]any, error) {
		return map[string]any{"listings": []any{map[string]any{"skill_id": "roadmap-reader"}}}, nil
	}
	cli.MarketplaceInstall = func(_ context.Context, skillID string) (map[string]any, error) {
		if skillID != "roadmap-reader" {
			t.Fatalf("unexpected skill id: %q", skillID)
		}
		return map[string]any{"installed": true, "skill_id": skillID}, nil
	}
	cli.MarketplaceMatrix = func(context.Context) (map[string]any, error) {
		return map[string]any{"matrix": []any{map[string]any{"skill_id": "roadmap-reader"}}}, nil
	}

	if err := cli.Run(context.Background(), "marketplace-publish", "/tmp/skill", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("marketplace-publish failed: %v", err)
	}
	buf.Reset()
	if err := cli.Run(context.Background(), "marketplace-list", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("marketplace-list failed: %v", err)
	}
	buf.Reset()
	if err := cli.Run(context.Background(), "marketplace-install", "roadmap-reader", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("marketplace-install failed: %v", err)
	}
	buf.Reset()
	if err := cli.Run(context.Background(), "marketplace-matrix", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("marketplace-matrix failed: %v", err)
	}
}

func TestCLIAuditExport(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.ExportAudit = func(path string) (map[string]any, error) {
		return map[string]any{"path": "/tmp/audit.json", "signature": "abc", "records": 3}, nil
	}
	if err := cli.Run(context.Background(), "audit-export", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("audit-export failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["signature"] != "abc" {
		t.Fatalf("unexpected audit output: %#v", out)
	}
}

func TestCLIAuditVerify(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.VerifyAudit = func(path string) (map[string]any, error) {
		return map[string]any{"path": "/tmp/audit.json", "valid": true, "signature": "abc"}, nil
	}
	if err := cli.Run(context.Background(), "audit-verify", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("audit-verify failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["valid"] != true {
		t.Fatalf("unexpected audit verify output: %#v", out)
	}
}

func TestCLIRuntimeExecutionReport(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.ExecutionReport = func(path string) (map[string]any, error) {
		return map[string]any{"path": "/tmp/runtime-report.json", "model": "gpt-4.1"}, nil
	}
	if err := cli.Run(context.Background(), "runtime-execution-report", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("runtime-execution-report failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["path"] != "/tmp/runtime-report.json" {
		t.Fatalf("unexpected runtime report output: %#v", out)
	}
}

func TestCLIProjectInventoryCommands(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	added := domainprojectinventory.Project{
		ID:      "p1",
		Path:    "/tmp/repo",
		AddedAt: "2026-02-13T00:00:00Z",
	}
	cli.ListProjects = func(context.Context) ([]domainprojectinventory.Project, error) {
		return []domainprojectinventory.Project{added}, nil
	}
	cli.AddProject = func(_ context.Context, path string) (domainprojectinventory.Project, error) {
		if path != "/tmp/repo" {
			t.Fatalf("unexpected project path: %q", path)
		}
		return added, nil
	}
	cli.RemoveProject = func(_ context.Context, selector string) error {
		if selector != "p1" {
			t.Fatalf("unexpected selector: %q", selector)
		}
		return nil
	}
	cli.InspectProject = func(_ context.Context, selector string) (domainprojectinventory.Project, error) {
		if selector != "p1" {
			t.Fatalf("unexpected selector: %q", selector)
		}
		return added, nil
	}

	if err := cli.Run(context.Background(), "project-add", "/tmp/repo", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("project-add failed: %v", err)
	}
	buf.Reset()
	if err := cli.Run(context.Background(), "project-list", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("project-list failed: %v", err)
	}
	buf.Reset()
	if err := cli.Run(context.Background(), "project-inspect", "p1", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("project-inspect failed: %v", err)
	}
	buf.Reset()
	if err := cli.Run(context.Background(), "project-remove", "p1", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("project-remove failed: %v", err)
	}
}

func TestCLIWorkspaceCommands(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	cli.ValidateWorkspace = func(context.Context) (domainworkspace.ValidationResult, error) {
		return domainworkspace.ValidationResult{
			Healthy: true,
			Links: []domainworkspace.LinkReport{
				{ProjectID: "p1", ProjectPath: "/tmp/repo", LinkPath: "/tmp/links/p1", Status: domainworkspace.LinkStatusOK},
			},
		}, nil
	}
	cli.PlanWorkspace = func(context.Context) (domainworkspace.PlanResult, error) {
		return domainworkspace.PlanResult{
			Actions: []domainworkspace.PlanAction{
				{Kind: domainworkspace.ActionSkip, LinkPath: "/tmp/links/p1", TargetPath: "/tmp/repo", Reason: "already healthy"},
			},
		}, nil
	}
	cli.RepairWorkspace = func(context.Context) (domainworkspace.RepairResult, error) {
		return domainworkspace.RepairResult{}, nil
	}

	if err := cli.Run(context.Background(), "workspace-validate", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("workspace-validate failed: %v", err)
	}
	buf.Reset()
	if err := cli.Run(context.Background(), "workspace-plan", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("workspace-plan failed: %v", err)
	}
	buf.Reset()
	if err := cli.Run(context.Background(), "workspace-repair", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("workspace-repair failed: %v", err)
	}
}

func TestCLITUIQuit(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.In = strings.NewReader("q\n")
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
	out := buf.String()
	if !strings.Contains(out, "AIOS Operations Console") || !strings.Contains(out, "bye") {
		t.Fatalf("unexpected tui output: %q", out)
	}
}

func TestCLITUIProjectsAndValidate(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.In = strings.NewReader("1\n2\nq\n")
	cli.ListProjects = func(context.Context) ([]domainprojectinventory.Project, error) {
		return []domainprojectinventory.Project{
			{ID: "p1", Path: "/tmp/repo", AddedAt: "2026-02-13T00:00:00Z"},
		}, nil
	}
	cli.ValidateWorkspace = func(context.Context) (domainworkspace.ValidationResult, error) {
		return domainworkspace.ValidationResult{
			Healthy: true,
		}, nil
	}
	cli.RepairWorkspace = func(context.Context) (domainworkspace.RepairResult, error) {
		return domainworkspace.RepairResult{}, nil
	}

	if err := cli.Run(context.Background(), "tui", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("tui failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/tmp/repo") {
		t.Fatalf("expected project listing in tui output: %q", out)
	}
	if !strings.Contains(out, "workspace links: healthy") {
		t.Fatalf("expected validation status in tui output: %q", out)
	}
}

func TestCLIProjectRemoveNotFound(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.RemoveProject = func(context.Context, string) error {
		return errors.New("project not found")
	}
	if err := cli.Run(context.Background(), "project-remove", "missing", "stdio", ":8080", "text"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCLIProjectInspectNotFound(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.InspectProject = func(context.Context, string) (domainprojectinventory.Project, error) {
		return domainprojectinventory.Project{}, errors.New("project not found")
	}
	if err := cli.Run(context.Background(), "project-inspect", "missing", "stdio", ":8080", "text"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCLIPackageSkill(t *testing.T) {
	root := t.TempDir()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run(context.Background(), "package-skill", skillDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("package-skill failed: %v", err)
	}
}

func TestCLIBackupConfigs(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_WORKSPACE_DIR", root)
	t.Setenv("AIOS_PROJECT_DIR", filepath.Join(root, "project"))
	cfg := DefaultConfig()
	skillsDir := filepath.Join(cfg.ProjectDir, ".agents", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, cfg)
	if err := cli.Run(context.Background(), "backup-configs", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("backup-configs failed: %v", err)
	}
	var out map[string]string
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["backup"] == "" {
		t.Fatalf("missing backup path: %#v", out)
	}
}

func TestCLIRestoreConfigs(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_WORKSPACE_DIR", root)
	t.Setenv("AIOS_PROJECT_DIR", filepath.Join(root, "project"))
	cfg := DefaultConfig()
	skillsDir := filepath.Join(cfg.ProjectDir, ".agents", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	backupRoot := filepath.Join(cfg.WorkspaceDir, "backups", "20260213-000000")
	if err := os.MkdirAll(filepath.Join(backupRoot, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(backupRoot, "skills", "restored.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, cfg)
	if err := cli.Run(context.Background(), "restore-configs", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("restore-configs failed: %v", err)
	}
	restoredFile := filepath.Join(cfg.ProjectDir, ".agents", "skills", "restored.txt")
	if _, err := os.Stat(restoredFile); err != nil {
		t.Fatalf("missing restored file: %v", err)
	}
}

func TestCLIExportStatusReport(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_WORKSPACE_DIR", root)
	t.Setenv("AIOS_PROJECT_DIR", filepath.Join(root, "project"))
	cfg := DefaultConfig()
	for _, p := range []string{cfg.WorkspaceDir, cfg.ProjectDir} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, cfg)
	if err := cli.Run(context.Background(), "export-status-report", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("export-status-report failed: %v", err)
	}

	var out map[string]string
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	report := out["report"]
	if report == "" {
		t.Fatalf("missing report path: %#v", out)
	}
	if _, err := os.Stat(report); err != nil {
		t.Fatalf("missing report file: %v", err)
	}
}

func TestCLIUninstallSkill(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_PROJECT_DIR", root)
	cfg := DefaultConfig()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, cfg)

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := cli.Run(context.Background(), "sync", skillDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	if err := cli.Run(context.Background(), "uninstall-skill", skillDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cfg.ProjectDir, ".agents", "skills", "roadmap-reader")); !os.IsNotExist(err) {
		t.Fatalf("expected canonical skill removed, stat err: %v", err)
	}
}

func TestCLIConnectGoogleDriveWithTokenOverride(t *testing.T) {
	t.Setenv("AIOS_OAUTH_TOKEN", "token-123")
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	gotToken := ""
	cli.ConnectGoogleDrive = func(_ context.Context, cmd domainonboarding.ConnectGoogleDriveCommand) (domainonboarding.ConnectGoogleDriveResult, error) {
		gotToken = cmd.TokenOverride
		return domainonboarding.ConnectGoogleDriveResult{}, nil
	}

	if err := cli.Run(context.Background(), "connect-google-drive", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("connect-google-drive failed: %v", err)
	}
	if gotToken != "token-123" {
		t.Fatalf("unexpected token: %q", gotToken)
	}
}

func TestCLIConnectGoogleDriveOAuthFlow(t *testing.T) {
	t.Setenv("AIOS_OAUTH_TOKEN", "")
	t.Setenv("AIOS_OAUTH_STATE", "s1")
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	gotToken := ""
	cli.ConnectGoogleDrive = func(_ context.Context, cmd domainonboarding.ConnectGoogleDriveCommand) (domainonboarding.ConnectGoogleDriveResult, error) {
		if cmd.State != "s1" {
			t.Fatalf("unexpected state: %q", cmd.State)
		}
		gotToken = cmd.TokenOverride
		return domainonboarding.ConnectGoogleDriveResult{CallbackURL: "http://127.0.0.1:9999/oauth/callback"}, nil
	}

	if err := cli.Run(context.Background(), "connect-google-drive", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("connect-google-drive failed: %v", err)
	}
	if gotToken != "" {
		t.Fatalf("unexpected token: %q", gotToken)
	}
	out := buf.String()
	if !strings.Contains(out, "oauth callback listening:") || !strings.Contains(out, "google drive connected") {
		t.Fatalf("unexpected output: %q", out)
	}
}

// AC2: CLI reads AIOS_OAUTH_TIMEOUT_SEC and passes it as command timeout.
func TestCLIConnectGoogleDriveRespectsTimeoutEnv(t *testing.T) {
	t.Setenv("AIOS_OAUTH_TOKEN", "tok-1")
	t.Setenv("AIOS_OAUTH_TIMEOUT_SEC", "45")
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	var gotTimeout time.Duration
	cli.ConnectGoogleDrive = func(_ context.Context, cmd domainonboarding.ConnectGoogleDriveCommand) (domainonboarding.ConnectGoogleDriveResult, error) {
		gotTimeout = cmd.Timeout
		return domainonboarding.ConnectGoogleDriveResult{}, nil
	}

	if err := cli.Run(context.Background(), "connect-google-drive", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("connect-google-drive failed: %v", err)
	}
	if gotTimeout != 45*time.Second {
		t.Fatalf("expected 45s timeout, got %v", gotTimeout)
	}
}

// AC2: Default timeout when env var is not set.
func TestCLIConnectGoogleDriveDefaultTimeout(t *testing.T) {
	t.Setenv("AIOS_OAUTH_TOKEN", "tok-1")
	t.Setenv("AIOS_OAUTH_TIMEOUT_SEC", "")
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	var gotTimeout time.Duration
	cli.ConnectGoogleDrive = func(_ context.Context, cmd domainonboarding.ConnectGoogleDriveCommand) (domainonboarding.ConnectGoogleDriveResult, error) {
		gotTimeout = cmd.Timeout
		return domainonboarding.ConnectGoogleDriveResult{}, nil
	}

	if err := cli.Run(context.Background(), "connect-google-drive", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("connect-google-drive failed: %v", err)
	}
	if gotTimeout != 120*time.Second {
		t.Fatalf("expected 120s default timeout, got %v", gotTimeout)
	}
}

// AC4: Token value must not appear in CLI text output.
func TestCLIConnectGoogleDriveOutputDoesNotLeakToken(t *testing.T) {
	t.Setenv("AIOS_OAUTH_TOKEN", "super-secret-token-xyz")
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	cli.ConnectGoogleDrive = func(_ context.Context, cmd domainonboarding.ConnectGoogleDriveCommand) (domainonboarding.ConnectGoogleDriveResult, error) {
		return domainonboarding.ConnectGoogleDriveResult{}, nil
	}

	if err := cli.Run(context.Background(), "connect-google-drive", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("connect-google-drive failed: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "super-secret-token-xyz") {
		t.Fatalf("CLI output must not contain the token value, got: %q", out)
	}
}

// AC4: Token value must not appear in CLI JSON output either.
func TestCLIConnectGoogleDriveJSONDoesNotLeakToken(t *testing.T) {
	t.Setenv("AIOS_OAUTH_TOKEN", "json-secret-tok")
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	cli.ConnectGoogleDrive = func(_ context.Context, cmd domainonboarding.ConnectGoogleDriveCommand) (domainonboarding.ConnectGoogleDriveResult, error) {
		return domainonboarding.ConnectGoogleDriveResult{}, nil
	}

	if err := cli.Run(context.Background(), "connect-google-drive", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("connect-google-drive failed: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "json-secret-tok") {
		t.Fatalf("JSON output must not contain the token value, got: %q", out)
	}
}

// AC6: tray-status surfaces connection state (google_drive: true/false).
func TestCLITrayStatusSurfacesConnectionState(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.TrayStatus = func() (TrayState, error) {
		return TrayState{
			UpdatedAt:   "2026-02-13T00:00:00Z",
			Skills:      []string{},
			Connections: map[string]bool{"google_drive": true},
		}, nil
	}

	if err := cli.Run(context.Background(), "tray-status", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("tray-status failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	conns, ok := out["connections"].(map[string]any)
	if !ok {
		t.Fatalf("expected connections map, got: %#v", out)
	}
	if conns["google_drive"] != true {
		t.Fatalf("expected google_drive=true in tray-status, got: %#v", conns)
	}
}

// AC6: tray-status text output shows connection status.
func TestCLITrayStatusTextShowsConnection(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.TrayStatus = func() (TrayState, error) {
		return TrayState{
			UpdatedAt:   "2026-02-13T00:00:00Z",
			Skills:      []string{},
			Connections: map[string]bool{"google_drive": false},
		}, nil
	}

	if err := cli.Run(context.Background(), "tray-status", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("tray-status failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "google_drive") {
		t.Fatalf("tray-status text must show google_drive connection, got: %q", out)
	}
}

// AC7: ConnectGoogleDrive function is resolved at call time (late binding),
// not at DefaultCLI() construction time. The function field can be replaced
// after construction, proving connectors are bound at execution time.
func TestCLIConnectGoogleDriveIsLateBound(t *testing.T) {
	t.Setenv("AIOS_OAUTH_TOKEN", "late-bound-tok")
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	callCount := 0
	cli.ConnectGoogleDrive = func(_ context.Context, cmd domainonboarding.ConnectGoogleDriveCommand) (domainonboarding.ConnectGoogleDriveResult, error) {
		callCount++
		return domainonboarding.ConnectGoogleDriveResult{}, nil
	}

	if err := cli.Run(context.Background(), "connect-google-drive", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("connect-google-drive failed: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected ConnectGoogleDrive called once, got %d", callCount)
	}
}

func TestCLITrayStatusJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.TrayStatus = func() (TrayState, error) {
		return TrayState{
			UpdatedAt: "2026-02-13T00:00:00Z",
			Skills:    []string{"roadmap-reader"},
			Connections: map[string]bool{
				"google_drive": true,
			},
		}, nil
	}

	if err := cli.Run(context.Background(), "tray-status", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("tray-status failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["updated_at"] == "" {
		t.Fatalf("missing updated_at: %#v", out)
	}
}

func TestCLISyncJSON(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_PROJECT_DIR", root)
	cfg := DefaultConfig()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, cfg)

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := cli.Run(context.Background(), "sync", skillDir, "stdio", ":8080", "json"); err != nil {
		t.Fatalf("sync json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["synced"] != true {
		t.Fatalf("expected synced=true, got: %#v", out)
	}
	if out["skill_id"] != "roadmap-reader" {
		t.Fatalf("unexpected skill_id: %#v", out)
	}
}

func TestCLITestSkillJSON(t *testing.T) {
	root := t.TempDir()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"query":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"status":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "tests", "fixture_01.json"), []byte(`{"query":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "tests", "expected_01.json"), []byte(`{"status":"ok"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := cli.Run(context.Background(), "test-skill", skillDir, "stdio", ":8080", "json"); err != nil {
		t.Fatalf("test-skill json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["results"] == nil {
		t.Fatalf("missing results in json output: %#v", out)
	}
}

func TestCLIInitSkillJSON(t *testing.T) {
	root := t.TempDir()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	dir := filepath.Join(root, "roadmap-reader")
	if err := cli.Run(context.Background(), "init-skill", dir, "stdio", ":8080", "json"); err != nil {
		t.Fatalf("init-skill json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["initialized"] != true {
		t.Fatalf("expected initialized=true, got: %#v", out)
	}
	if out["skill_dir"] != dir {
		t.Fatalf("unexpected skill_dir: %#v", out)
	}
}

func TestCLILintSkillJSON(t *testing.T) {
	root := t.TempDir()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "prompt.md"), []byte("# p"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run(context.Background(), "lint-skill", skillDir, "stdio", ":8080", "json"); err != nil {
		t.Fatalf("lint-skill json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["valid"] != true {
		t.Fatalf("expected valid=true, got: %#v", out)
	}
}

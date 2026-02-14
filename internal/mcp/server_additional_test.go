package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/felixgeelhaar/aios/internal/sync"
)

func TestNewServerConstructs(t *testing.T) {
	srv := NewServer("0.1.0")
	if srv == nil {
		t.Fatal("expected server instance")
	}
	if len(srv.Tools()) == 0 {
		t.Fatal("expected tools to be registered")
	}
}

func TestSyncStateToolHandlesNilEngine(t *testing.T) {
	srv := NewServerWithDeps("0.1.0", ServerDeps{})
	tool, ok := srv.GetTool("sync_state")
	if !ok {
		t.Fatal("missing sync_state tool")
	}
	_, err := tool.Execute(context.Background(), json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error when sync engine is missing")
	}
}

func TestSyncStateToolReturnsState(t *testing.T) {
	srv := NewServerWithDeps("0.1.0", ServerDeps{Sync: sync.NewEngine()})
	tool, ok := srv.GetTool("sync_state")
	if !ok {
		t.Fatal("missing sync_state tool")
	}
	out, err := tool.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("sync_state failed: %v", err)
	}
	if out.(string) == "" {
		t.Fatal("expected non-empty sync state")
	}
}

func TestModelPolicyPacksToolReturnsPacks(t *testing.T) {
	srv := NewServerWithDeps("0.1.0", ServerDeps{Sync: sync.NewEngine()})
	tool, ok := srv.GetTool("model_policy_packs")
	if !ok {
		t.Fatal("missing model_policy_packs tool")
	}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("model_policy_packs failed: %v", err)
	}
	body, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected output: %#v", result)
	}
	if _, ok := body["policy_packs"]; !ok {
		t.Fatalf("missing policy_packs: %#v", body)
	}
}

func TestDoctorToolUsesDefaultWhenNil(t *testing.T) {
	srv := NewServerWithDeps("0.1.0", ServerDeps{})
	tool, ok := srv.GetTool("doctor")
	if !ok {
		t.Fatal("missing doctor tool")
	}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("doctor failed: %v", err)
	}
	body, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected doctor output: %#v", result)
	}
	if body["overall"] != true {
		t.Fatalf("expected overall=true: %#v", body)
	}
}

func TestValidateSkillDirAndPackageSkillTools(t *testing.T) {
	root := t.TempDir()
	workDir := filepath.Join(root, "workspace")
	t.Setenv("AIOS_WORKSPACE_DIR", workDir)

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
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"status":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "tests", "fixture_01.json"), []byte(`{"q":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "tests", "expected_01.json"), []byte(`{"status":"ok"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := NewServerWithDeps("0.1.0", ServerDeps{Sync: sync.NewEngine()})

	validateTool, ok := srv.GetTool("validate_skill_dir")
	if !ok {
		t.Fatal("missing validate_skill_dir tool")
	}
	validateOut, err := validateTool.Execute(context.Background(), json.RawMessage(`{"skill_dir":"`+skillDir+`"}`))
	if err != nil {
		t.Fatalf("validate_skill_dir failed: %v", err)
	}
	validateMap, ok := validateOut.(map[string]any)
	if !ok {
		t.Fatalf("unexpected validate_skill_dir output: %#v", validateOut)
	}
	if validateMap["valid"] != true {
		t.Fatalf("expected valid=true: %#v", validateMap)
	}

	fixtureTool, ok := srv.GetTool("run_fixture_suite")
	if !ok {
		t.Fatal("missing run_fixture_suite tool")
	}
	fixtureOut, err := fixtureTool.Execute(context.Background(), json.RawMessage(`{"skill_dir":"`+skillDir+`"}`))
	if err != nil {
		t.Fatalf("run_fixture_suite failed: %v", err)
	}
	fixtureMap, ok := fixtureOut.(map[string]any)
	if !ok {
		t.Fatalf("unexpected run_fixture_suite output: %#v", fixtureOut)
	}
	if fixtureMap["passed"] == nil {
		t.Fatalf("missing passed count: %#v", fixtureMap)
	}

	packageTool, ok := srv.GetTool("package_skill")
	if !ok {
		t.Fatal("missing package_skill tool")
	}
	packageOut, err := packageTool.Execute(context.Background(), json.RawMessage(`{"skill_dir":"`+skillDir+`"}`))
	if err != nil {
		t.Fatalf("package_skill failed: %v", err)
	}
	packageMap, ok := packageOut.(map[string]any)
	if !ok {
		t.Fatalf("unexpected package_skill output: %#v", packageOut)
	}
	artifact, _ := packageMap["artifact"].(string)
	if artifact == "" {
		t.Fatalf("missing artifact path: %#v", packageMap)
	}
	if _, err := os.Stat(artifact); err != nil {
		t.Fatalf("missing artifact file: %v", err)
	}
}

func TestMarketplacePublishAndListTools(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_WORKSPACE_DIR", root)

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
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"status":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := NewServerWithDeps("0.1.0", ServerDeps{Sync: sync.NewEngine()})
	publishTool, ok := srv.GetTool("marketplace_publish")
	if !ok {
		t.Fatal("missing marketplace_publish tool")
	}
	if _, err := publishTool.Execute(context.Background(), json.RawMessage(`{"skill_dir":"`+skillDir+`"}`)); err != nil {
		t.Fatalf("marketplace_publish failed: %v", err)
	}

	listTool, ok := srv.GetTool("marketplace_list")
	if !ok {
		t.Fatal("missing marketplace_list tool")
	}
	listOut, err := listTool.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("marketplace_list failed: %v", err)
	}
	listMap, ok := listOut.(map[string]any)
	if !ok {
		t.Fatalf("unexpected marketplace_list output: %#v", listOut)
	}
	if listMap["listings"] == nil {
		t.Fatalf("missing listings: %#v", listMap)
	}
}

func TestUninstallSkillToolUsesDeps(t *testing.T) {
	srv := NewServerWithDeps("0.1.0", ServerDeps{
		Uninstall: func(string) (string, error) {
			return "roadmap-reader", nil
		},
	})
	tool, ok := srv.GetTool("uninstall_skill")
	if !ok {
		t.Fatal("missing uninstall_skill tool")
	}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{"skill_dir":"/tmp/skill"}`))
	if err != nil {
		t.Fatalf("uninstall_skill failed: %v", err)
	}
	body, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected uninstall output: %#v", result)
	}
	if body["uninstalled"] != "roadmap-reader" {
		t.Fatalf("unexpected uninstall result: %#v", body)
	}
}

func TestBuildInfoResourceReturnsJSON(t *testing.T) {
	srv := NewServerWithDeps("0.1.0", ServerDeps{Version: "1.2.3", Commit: "abc", BuildDate: "2026-02-13"})
	res, ok := srv.GetResource("aios://status/build")
	if !ok {
		t.Fatal("missing aios://status/build resource")
	}
	content, err := res.Read(context.Background(), "aios://status/build")
	if err != nil {
		t.Fatalf("read build resource failed: %v", err)
	}
	if !strings.Contains(content.Text, "1.2.3") {
		t.Fatalf("unexpected build resource content: %q", content.Text)
	}
}

func TestDocsResourceReadsMarkdown(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	root := cwd
	for i := 0; i < 6; i++ {
		if _, statErr := os.Stat(filepath.Join(root, "docs", "prd.md")); statErr == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			break
		}
		root = parent
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "prd.md")); err != nil {
		t.Fatalf("docs/prd.md not found from %s: %v", cwd, err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	srv := NewServerWithDeps("0.1.0", ServerDeps{})
	res, ok := srv.GetResource("docs://{name}")
	if !ok {
		t.Fatal("missing docs resource")
	}
	content, err := res.Read(context.Background(), "docs://prd")
	if err != nil {
		t.Fatalf("read docs resource failed: %v", err)
	}
	if content.MimeType != "text/markdown" {
		t.Fatalf("unexpected mime type: %q", content.MimeType)
	}
}

func TestDocsIndexResourceListsDocs(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	root := cwd
	for i := 0; i < 6; i++ {
		if _, statErr := os.Stat(filepath.Join(root, "docs", "prd.md")); statErr == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			break
		}
		root = parent
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "prd.md")); err != nil {
		t.Fatalf("docs/prd.md not found from %s: %v", cwd, err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	srv := NewServerWithDeps("0.1.0", ServerDeps{})
	res, ok := srv.GetResource("docs://index")
	if !ok {
		t.Fatal("missing docs index resource")
	}
	content, err := res.Read(context.Background(), "docs://index")
	if err != nil {
		t.Fatalf("read docs index resource failed: %v", err)
	}
	if !strings.Contains(content.Text, "docs") {
		t.Fatalf("unexpected docs index content: %q", content.Text)
	}
}

func TestProjectInspectToolReturnsProject(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_WORKSPACE_DIR", root)

	repoPath := filepath.Join(root, "repo")
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatal(err)
	}

	srv := NewServerWithDeps("0.1.0", ServerDeps{})
	trackTool, ok := srv.GetTool("project_track")
	if !ok {
		t.Fatal("missing project_track tool")
	}
	trackOut, err := trackTool.Execute(context.Background(), json.RawMessage(`{"path":"`+repoPath+`"}`))
	if err != nil {
		t.Fatalf("project_track failed: %v", err)
	}
	trackMap, ok := trackOut.(map[string]any)
	if !ok {
		t.Fatalf("unexpected project_track output: %#v", trackOut)
	}
	projectID, _ := trackMap["id"].(string)
	if projectID == "" {
		t.Fatalf("missing project id: %#v", trackMap)
	}

	inspectTool, ok := srv.GetTool("project_inspect")
	if !ok {
		t.Fatal("missing project_inspect tool")
	}
	inspectOut, err := inspectTool.Execute(context.Background(), json.RawMessage(`{"selector":"`+projectID+`"}`))
	if err != nil {
		t.Fatalf("project_inspect failed: %v", err)
	}
	inspectMap, ok := inspectOut.(map[string]any)
	if !ok {
		t.Fatalf("unexpected project_inspect output: %#v", inspectOut)
	}
	if inspectMap["id"] != projectID {
		t.Fatalf("unexpected project_inspect id: %#v", inspectMap)
	}
}

func TestAnalyticsTrendResourceReturnsJSON(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_WORKSPACE_DIR", root)

	srv := NewServerWithDeps("0.1.0", ServerDeps{})
	res, ok := srv.GetResource("aios://analytics/trend")
	if !ok {
		t.Fatal("missing aios://analytics/trend resource")
	}
	content, err := res.Read(context.Background(), "aios://analytics/trend")
	if err != nil {
		t.Fatalf("read analytics trend resource failed: %v", err)
	}
	if content.MimeType != "application/json" {
		t.Fatalf("unexpected mime type: %q", content.MimeType)
	}
}

func TestNewServerUninstallSkillUsesDefaultHandler(t *testing.T) {
	root := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	canonicalDir := filepath.Join(root, ".agents", "skills", "roadmap-reader")
	if err := os.MkdirAll(canonicalDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(canonicalDir, "SKILL.md"), []byte("---\nname: roadmap-reader\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := NewServer("0.1.0")
	tool, ok := srv.GetTool("uninstall_skill")
	if !ok {
		t.Fatal("missing uninstall_skill tool")
	}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{"skill_dir":"`+skillDir+`"}`))
	if err != nil {
		t.Fatalf("uninstall_skill failed: %v", err)
	}
	body, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected uninstall output: %#v", result)
	}
	if body["uninstalled"] != "roadmap-reader" {
		t.Fatalf("unexpected uninstall result: %#v", body)
	}
}

func TestUninstallSkillToolErrorsWithoutDeps(t *testing.T) {
	srv := NewServerWithDeps("0.1.0", ServerDeps{})
	tool, ok := srv.GetTool("uninstall_skill")
	if !ok {
		t.Fatal("missing uninstall_skill tool")
	}
	_, err := tool.Execute(context.Background(), json.RawMessage(`{"skill_dir":"/tmp/skill"}`))
	if err == nil {
		t.Fatal("expected uninstall error when deps.Uninstall is nil")
	}
}

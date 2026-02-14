package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	applicationproject "github.com/felixgeelhaar/aios/internal/application/projectinventory"
	applicationworkspace "github.com/felixgeelhaar/aios/internal/application/workspaceorchestration"
)

func TestProjectTrackAndWorkspaceRepairServices(t *testing.T) {
	root := t.TempDir()

	repoPath := filepath.Join(root, "repo-a")
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatal(err)
	}

	repo := mcpProjectInventoryRepository{workspaceDir: root}
	projectService := applicationproject.NewService(repo, mcpPathCanonicalizer{})
	workspaceService := applicationworkspace.NewService(
		mcpInventoryProjectSource{repo: repo},
		mcpFilesystemWorkspaceLinks{workspaceDir: root},
	)

	project, err := projectService.Track(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("track failed: %v", err)
	}
	inspected, err := projectService.Inspect(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("inspect failed: %v", err)
	}
	if inspected.Path != project.Path {
		t.Fatalf("unexpected inspected project: %#v", inspected)
	}

	projects, err := projectService.List(context.Background())
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected one project, got %d", len(projects))
	}

	validation, err := workspaceService.Validate(context.Background())
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if validation.Healthy {
		t.Fatalf("expected unhealthy workspace before repair: %#v", validation)
	}

	repair, err := workspaceService.Repair(context.Background())
	if err != nil {
		t.Fatalf("repair failed: %v", err)
	}
	if len(repair.Applied) != 1 {
		t.Fatalf("expected one applied repair, got %d (skipped=%d)", len(repair.Applied), len(repair.Skipped))
	}

	validation, err = workspaceService.Validate(context.Background())
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if !validation.Healthy {
		t.Fatal("expected healthy workspace after repair")
	}

	if err := projectService.Untrack(context.Background(), project.ID); err != nil {
		t.Fatalf("untrack failed: %v", err)
	}
	projects, err = projectService.List(context.Background())
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(projects) != 0 {
		t.Fatalf("expected empty inventory, got %d", len(projects))
	}
}

func TestH2ToolsExecuteEndToEnd(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_WORKSPACE_DIR", root)

	repoPath := filepath.Join(root, "repo-tool")
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatal(err)
	}

	srv := NewServerWithDeps("0.1.0", ServerDeps{})

	projectTrack, ok := srv.GetTool("project_track")
	if !ok {
		t.Fatal("missing project_track tool")
	}
	trackOut, err := projectTrack.Execute(context.Background(), json.RawMessage(`{"path":"`+repoPath+`"}`))
	if err != nil {
		t.Fatalf("project_track failed: %v", err)
	}
	trackMap, ok := trackOut.(map[string]any)
	if !ok {
		t.Fatalf("unexpected project_track output: %#v", trackOut)
	}
	projectID, _ := trackMap["id"].(string)
	if projectID == "" {
		t.Fatalf("missing project id in output: %#v", trackMap)
	}

	projectList, ok := srv.GetTool("project_list")
	if !ok {
		t.Fatal("missing project_list tool")
	}
	listOut, err := projectList.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("project_list failed: %v", err)
	}
	listMap, ok := listOut.(map[string]any)
	if !ok {
		t.Fatalf("unexpected project_list output: %#v", listOut)
	}
	if got := sliceLen(listMap["projects"]); got != 1 {
		t.Fatalf("expected one tracked project, got %d", got)
	}

	workspaceValidate, ok := srv.GetTool("workspace_validate")
	if !ok {
		t.Fatal("missing workspace_validate tool")
	}
	validateOut, err := workspaceValidate.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("workspace_validate failed: %v", err)
	}
	validateMap, ok := validateOut.(map[string]any)
	if !ok {
		t.Fatalf("unexpected workspace_validate output: %#v", validateOut)
	}
	if healthy, _ := validateMap["healthy"].(bool); healthy {
		t.Fatalf("expected unhealthy before repair: %#v", validateMap)
	}

	workspacePlan, ok := srv.GetTool("workspace_plan")
	if !ok {
		t.Fatal("missing workspace_plan tool")
	}
	planOut, err := workspacePlan.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("workspace_plan failed: %v", err)
	}
	planMap, ok := planOut.(map[string]any)
	if !ok {
		t.Fatalf("unexpected workspace_plan output: %#v", planOut)
	}
	if got := sliceLen(planMap["actions"]); got == 0 {
		t.Fatalf("expected non-empty workspace plan: %#v", planMap)
	}

	workspaceRepair, ok := srv.GetTool("workspace_repair")
	if !ok {
		t.Fatal("missing workspace_repair tool")
	}
	if _, err := workspaceRepair.Execute(context.Background(), json.RawMessage(`{}`)); err != nil {
		t.Fatalf("workspace_repair failed: %v", err)
	}

	validateOut, err = workspaceValidate.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("workspace_validate failed: %v", err)
	}
	validateMap, ok = validateOut.(map[string]any)
	if !ok {
		t.Fatalf("unexpected workspace_validate output: %#v", validateOut)
	}
	if healthy, _ := validateMap["healthy"].(bool); !healthy {
		t.Fatalf("expected healthy after repair: %#v", validateMap)
	}

	projectUntrack, ok := srv.GetTool("project_untrack")
	if !ok {
		t.Fatal("missing project_untrack tool")
	}
	if _, err := projectUntrack.Execute(context.Background(), json.RawMessage(`{"selector":"`+projectID+`"}`)); err != nil {
		t.Fatalf("project_untrack failed: %v", err)
	}
}

func sliceLen(v any) int {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return 0
	}
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		return rv.Len()
	default:
		return 0
	}
}

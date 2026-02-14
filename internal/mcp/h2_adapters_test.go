package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	domainproject "github.com/felixgeelhaar/aios/internal/domain/projectinventory"
	domainworkspace "github.com/felixgeelhaar/aios/internal/domain/workspaceorchestration"
)

func TestMcpProjectInventorySaveAndLoad(t *testing.T) {
	root := t.TempDir()
	repo := mcpProjectInventoryRepository{workspaceDir: root}
	project := domainproject.Project{ID: "p1", Path: "/tmp/repo", AddedAt: "2026-02-13T00:00:00Z"}

	if err := repo.Save(context.Background(), domainproject.Inventory{Projects: []domainproject.Project{project}}); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	inv, err := repo.Load(context.Background())
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(inv.Projects) != 1 {
		t.Fatalf("expected one project, got %d", len(inv.Projects))
	}
}

func TestMcpWorkspaceLinksInspectAndEnsure(t *testing.T) {
	root := t.TempDir()
	links := mcpFilesystemWorkspaceLinks{workspaceDir: root}
	projectID := "p1"
	projectPath := filepath.Join(root, "repo")
	linkPath := mcpWorkspaceLinkPath(root, projectID)

	report, err := links.Inspect(projectID, projectPath)
	if err != nil {
		t.Fatalf("inspect missing: %v", err)
	}
	if report.Status != domainworkspace.LinkStatusMissing {
		t.Fatalf("expected missing status, got %s", report.Status)
	}

	if err := os.MkdirAll(filepath.Dir(linkPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(linkPath, []byte("conflict"), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err = links.Inspect(projectID, projectPath)
	if err != nil {
		t.Fatalf("inspect conflict: %v", err)
	}
	if report.Status != domainworkspace.LinkStatusConflict {
		t.Fatalf("expected conflict status, got %s", report.Status)
	}

	if err := os.Remove(linkPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "other"), linkPath); err != nil {
		t.Fatal(err)
	}
	report, err = links.Inspect(projectID, projectPath)
	if err != nil {
		t.Fatalf("inspect broken: %v", err)
	}
	if report.Status != domainworkspace.LinkStatusBroken {
		t.Fatalf("expected broken status, got %s", report.Status)
	}

	if err := os.Remove(linkPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(projectPath, linkPath); err != nil {
		t.Fatal(err)
	}
	report, err = links.Inspect(projectID, projectPath)
	if err != nil {
		t.Fatalf("inspect ok: %v", err)
	}
	if report.Status != domainworkspace.LinkStatusOK {
		t.Fatalf("expected ok status, got %s", report.Status)
	}

	if err := os.Remove(linkPath); err != nil {
		t.Fatal(err)
	}
	if err := links.Ensure(projectID, projectPath); err != nil {
		t.Fatalf("ensure create failed: %v", err)
	}
	if _, err := os.Lstat(linkPath); err != nil {
		t.Fatalf("missing symlink: %v", err)
	}

	if err := os.Remove(linkPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "other"), linkPath); err != nil {
		t.Fatal(err)
	}
	if err := links.Ensure(projectID, projectPath); err != nil {
		t.Fatalf("ensure replace failed: %v", err)
	}

	if err := os.Remove(linkPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(linkPath, []byte("conflict"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := links.Ensure(projectID, projectPath); err == nil {
		t.Fatal("expected error when link path is not a symlink")
	}
}

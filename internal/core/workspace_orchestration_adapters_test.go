package core

import (
	"os"
	"path/filepath"
	"testing"

	domainworkspace "github.com/felixgeelhaar/aios/internal/domain/workspaceorchestration"
)

func TestFilesystemWorkspaceLinksInspectStatuses(t *testing.T) {
	root := t.TempDir()
	links := filesystemWorkspaceLinks{workspaceDir: root}
	projectID := "p1"
	projectPath := filepath.Join(root, "project")
	linkPath := workspaceLinkPath(root, projectID)

	report, err := links.Inspect(projectID, projectPath)
	if err != nil {
		t.Fatalf("inspect missing link: %v", err)
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
	wrongTarget := filepath.Join(root, "other")
	if err := os.Symlink(wrongTarget, linkPath); err != nil {
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
}

func TestFilesystemWorkspaceLinksEnsure(t *testing.T) {
	root := t.TempDir()
	links := filesystemWorkspaceLinks{workspaceDir: root}
	projectID := "p1"
	projectPath := filepath.Join(root, "project")
	linkPath := workspaceLinkPath(root, projectID)

	if err := links.Ensure(projectID, projectPath); err != nil {
		t.Fatalf("ensure create failed: %v", err)
	}
	if _, err := os.Lstat(linkPath); err != nil {
		t.Fatalf("expected symlink to exist: %v", err)
	}

	other := filepath.Join(root, "other")
	if err := os.Remove(linkPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, linkPath); err != nil {
		t.Fatal(err)
	}
	if err := links.Ensure(projectID, projectPath); err != nil {
		t.Fatalf("ensure replace failed: %v", err)
	}
	current, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("readlink failed: %v", err)
	}
	if filepath.Clean(current) != filepath.Clean(projectPath) {
		t.Fatalf("expected symlink to %s, got %s", projectPath, current)
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

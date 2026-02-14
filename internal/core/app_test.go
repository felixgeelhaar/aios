package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppRunRejectsUnknownMode(t *testing.T) {
	app := NewApp(DefaultConfig())
	if err := app.Run("bad"); err == nil {
		t.Fatal("expected error for unknown mode")
	}
}

func TestAppRunCreatesWorkspaceForCLIMode(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultConfig()
	cfg.WorkspaceDir = filepath.Join(root, "workspace")

	app := NewApp(cfg)
	if err := app.Run("cli"); err != nil {
		t.Fatalf("expected cli mode to start: %v", err)
	}
	if _, err := os.Stat(cfg.WorkspaceDir); err != nil {
		t.Fatalf("workspace directory was not created: %v", err)
	}
}

func TestAppRunTrayModeInitializesState(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultConfig()
	cfg.WorkspaceDir = filepath.Join(root, "workspace")
	cfg.ProjectDir = filepath.Join(root, "project")
	cfg.TokenService = "aios"

	app := NewApp(cfg)
	if err := app.Run("tray"); err != nil {
		t.Fatalf("tray mode should initialize: %v", err)
	}
	if _, err := os.Stat(trayStatePath(cfg.WorkspaceDir)); err != nil {
		t.Fatalf("expected tray state file to exist: %v", err)
	}
}

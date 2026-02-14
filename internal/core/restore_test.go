package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/felixgeelhaar/aios/internal/domain/agentregistry"
)

func TestRestoreClientConfigs(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultConfig()
	cfg.ProjectDir = filepath.Join(root, "project")

	backup := filepath.Join(root, "backup")
	// Create backup with canonical skills dir content.
	if err := os.MkdirAll(filepath.Join(backup, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(backup, "skills", "x.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RestoreClientConfigs(cfg, backup); err != nil {
		t.Fatalf("restore failed: %v", err)
	}
	restoredPath := filepath.Join(cfg.ProjectDir, agentregistry.CanonicalSkillsDir, "x.txt")
	if _, err := os.Stat(restoredPath); err != nil {
		t.Fatalf("missing restored file: %v", err)
	}
}

func TestRestoreClientConfigsRequiresBackupDir(t *testing.T) {
	cfg := DefaultConfig()
	if err := RestoreClientConfigs(cfg, ""); err == nil {
		t.Fatal("expected error for empty backup dir")
	}
}

func TestLatestBackupDirReturnsMostRecent(t *testing.T) {
	root := t.TempDir()
	backups := filepath.Join(root, "backups")
	if err := os.MkdirAll(filepath.Join(backups, "20260213T120000Z"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(backups, "20260213T120500Z"), 0o755); err != nil {
		t.Fatal(err)
	}
	latest, err := LatestBackupDir(root)
	if err != nil {
		t.Fatalf("latest backup failed: %v", err)
	}
	if !strings.HasSuffix(latest, "20260213T120500Z") {
		t.Fatalf("unexpected latest backup dir: %s", latest)
	}
}

func TestLatestBackupDirErrorsWhenEmpty(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "backups"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := LatestBackupDir(root); err == nil {
		t.Fatal("expected error when no backups found")
	}
}

package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/felixgeelhaar/aios/internal/domain/agentregistry"
)

func TestBackupClientConfigs(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultConfig()
	cfg.WorkspaceDir = root
	cfg.ProjectDir = filepath.Join(root, "project")

	// Create canonical skills directory with a file.
	skillsDir := filepath.Join(cfg.ProjectDir, agentregistry.CanonicalSkillsDir)
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "x.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	backup, err := BackupClientConfigs(cfg)
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(backup, "skills", "x.txt")); err != nil {
		t.Fatalf("missing backup file: %v", err)
	}
}

func TestCopyDirMissingSourceIsNoop(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "missing")
	dst := filepath.Join(root, "dest")

	if err := copyDir(src, dst); err != nil {
		t.Fatalf("copyDir should ignore missing source: %v", err)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("expected destination directory to exist: %v", err)
	}
}

func TestCopyDirCopiesNestedDirectories(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")

	if err := os.MkdirAll(filepath.Join(src, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "nested", "file.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := copyDir(src, dst); err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "nested", "file.txt")); err != nil {
		t.Fatalf("missing nested file: %v", err)
	}
}

func TestCopyFileMissingSourceReturnsError(t *testing.T) {
	root := t.TempDir()
	if err := copyFile(filepath.Join(root, "missing.txt"), filepath.Join(root, "out.txt")); err == nil {
		t.Fatal("expected error for missing source file")
	}
}

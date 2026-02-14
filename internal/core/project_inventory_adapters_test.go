package core

import (
	"context"
	"path/filepath"
	"testing"

	domain "github.com/felixgeelhaar/aios/internal/domain/projectinventory"
)

func TestFileProjectInventoryRepositoryLoadSave(t *testing.T) {
	root := t.TempDir()
	repo := fileProjectInventoryRepository{workspaceDir: root}
	inv := domain.Inventory{
		Projects: []domain.Project{
			{ID: "b", Path: "/z", AddedAt: "2026-02-13T00:00:00Z"},
			{ID: "a", Path: "/a", AddedAt: "2026-02-13T00:00:00Z"},
		},
	}
	if err := repo.Save(context.Background(), inv); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	loaded, err := repo.Load(context.Background())
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded.Projects) != 2 {
		t.Fatalf("unexpected projects length: %d", len(loaded.Projects))
	}
	if loaded.Projects[0].Path != "/a" {
		t.Fatalf("expected sorted projects, got %#v", loaded.Projects)
	}
	if _, err := (absPathCanonicalizer{}).Canonicalize(filepath.Join(".", "x")); err != nil {
		t.Fatalf("canonicalize failed: %v", err)
	}
}

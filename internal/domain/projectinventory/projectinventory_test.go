package projectinventory_test

import (
	"testing"

	"github.com/felixgeelhaar/aios/internal/domain/projectinventory"
)

func TestNormalizeSelector(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  /path/to/project  ", "/path/to/project"},
		{"abc123", "abc123"},
		{"   ", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := projectinventory.NormalizeSelector(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeSelector(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestProjectID(t *testing.T) {
	id := projectinventory.ProjectID("/home/user/project")
	if id == "" {
		t.Fatal("expected non-empty project ID")
	}
	// Same input produces same ID (deterministic).
	if id2 := projectinventory.ProjectID("/home/user/project"); id != id2 {
		t.Errorf("expected deterministic ID, got %q and %q", id, id2)
	}
	// Different input produces different ID.
	if id3 := projectinventory.ProjectID("/home/user/other"); id == id3 {
		t.Error("expected different IDs for different paths")
	}
}

func TestFindBySelector_ByID(t *testing.T) {
	inv := projectinventory.Inventory{
		Projects: []projectinventory.Project{
			{ID: "abc123", Path: "/home/user/project-a"},
		},
	}
	p, ok := inv.FindBySelector("abc123")
	if !ok {
		t.Fatal("expected to find project by ID")
	}
	if p.Path != "/home/user/project-a" {
		t.Errorf("unexpected path: %s", p.Path)
	}
}

func TestFindBySelector_ByPath(t *testing.T) {
	inv := projectinventory.Inventory{
		Projects: []projectinventory.Project{
			{ID: "abc123", Path: "/home/user/project-a"},
		},
	}
	p, ok := inv.FindBySelector("/home/user/project-a")
	if !ok {
		t.Fatal("expected to find project by path")
	}
	if p.ID != "abc123" {
		t.Errorf("unexpected ID: %s", p.ID)
	}
}

func TestFindBySelector_NotFound(t *testing.T) {
	inv := projectinventory.Inventory{
		Projects: []projectinventory.Project{
			{ID: "abc123", Path: "/home/user/project-a"},
		},
	}
	_, ok := inv.FindBySelector("nonexistent")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestFindBySelector_EmptyInventory(t *testing.T) {
	inv := projectinventory.Inventory{}
	_, ok := inv.FindBySelector("anything")
	if ok {
		t.Fatal("expected not found on empty inventory")
	}
}

func TestTrack_NewProject(t *testing.T) {
	inv := projectinventory.Inventory{}
	added := inv.Track(projectinventory.Project{ID: "abc", Path: "/a"})
	if !added {
		t.Fatal("expected project to be added")
	}
	if len(inv.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(inv.Projects))
	}
}

func TestTrack_DuplicateByID(t *testing.T) {
	inv := projectinventory.Inventory{
		Projects: []projectinventory.Project{{ID: "abc", Path: "/a"}},
	}
	added := inv.Track(projectinventory.Project{ID: "abc", Path: "/b"})
	if added {
		t.Fatal("expected duplicate to be rejected")
	}
	if len(inv.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(inv.Projects))
	}
}

func TestTrack_DuplicateByPath(t *testing.T) {
	inv := projectinventory.Inventory{
		Projects: []projectinventory.Project{{ID: "abc", Path: "/a"}},
	}
	added := inv.Track(projectinventory.Project{ID: "def", Path: "/a"})
	if added {
		t.Fatal("expected duplicate by path to be rejected")
	}
}

func TestUntrack_Found(t *testing.T) {
	inv := projectinventory.Inventory{
		Projects: []projectinventory.Project{
			{ID: "abc", Path: "/a"},
			{ID: "def", Path: "/b"},
		},
	}
	removed := inv.Untrack("abc")
	if !removed {
		t.Fatal("expected project to be removed")
	}
	if len(inv.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(inv.Projects))
	}
	if inv.Projects[0].ID != "def" {
		t.Errorf("wrong project remained: %s", inv.Projects[0].ID)
	}
}

func TestUntrack_ByPath(t *testing.T) {
	inv := projectinventory.Inventory{
		Projects: []projectinventory.Project{{ID: "abc", Path: "/a"}},
	}
	removed := inv.Untrack("/a")
	if !removed {
		t.Fatal("expected project to be removed by path")
	}
	if len(inv.Projects) != 0 {
		t.Fatalf("expected 0 projects, got %d", len(inv.Projects))
	}
}

func TestUntrack_NotFound(t *testing.T) {
	inv := projectinventory.Inventory{
		Projects: []projectinventory.Project{{ID: "abc", Path: "/a"}},
	}
	removed := inv.Untrack("nonexistent")
	if removed {
		t.Fatal("expected not found")
	}
	if len(inv.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(inv.Projects))
	}
}

func TestUntrack_EmptyInventory(t *testing.T) {
	inv := projectinventory.Inventory{}
	removed := inv.Untrack("anything")
	if removed {
		t.Fatal("expected not found on empty inventory")
	}
}

func TestSortedProjects(t *testing.T) {
	inv := projectinventory.Inventory{
		Projects: []projectinventory.Project{
			{ID: "c", Path: "/z/project"},
			{ID: "a", Path: "/a/project"},
			{ID: "b", Path: "/m/project"},
		},
	}
	sorted := inv.SortedProjects()
	if len(sorted) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(sorted))
	}
	if sorted[0].Path != "/a/project" {
		t.Errorf("expected /a/project first, got %s", sorted[0].Path)
	}
	if sorted[1].Path != "/m/project" {
		t.Errorf("expected /m/project second, got %s", sorted[1].Path)
	}
	if sorted[2].Path != "/z/project" {
		t.Errorf("expected /z/project third, got %s", sorted[2].Path)
	}
}

func TestSortedProjects_DoesNotMutateOriginal(t *testing.T) {
	inv := projectinventory.Inventory{
		Projects: []projectinventory.Project{
			{ID: "b", Path: "/b"},
			{ID: "a", Path: "/a"},
		},
	}
	_ = inv.SortedProjects()
	if inv.Projects[0].Path != "/b" {
		t.Error("original inventory was mutated")
	}
}

func TestSortedProjects_Empty(t *testing.T) {
	inv := projectinventory.Inventory{}
	sorted := inv.SortedProjects()
	if len(sorted) != 0 {
		t.Fatalf("expected 0 projects, got %d", len(sorted))
	}
}

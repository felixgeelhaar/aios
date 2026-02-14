package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/felixgeelhaar/aios/internal/domain/agentregistry"
)

func TestRefreshTrayStateCollectsSkillsAndConnection(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultConfig()
	cfg.WorkspaceDir = root
	cfg.ProjectDir = filepath.Join(root, "project")

	// Create canonical skills directory with a skill subdirectory.
	skillDir := filepath.Join(cfg.ProjectDir, agentregistry.CanonicalSkillsDir, "roadmap-reader")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: roadmap-reader\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	connected := true
	state, err := RefreshTrayState(cfg, &connected)
	if err != nil {
		t.Fatalf("refresh tray state: %v", err)
	}

	if !state.Connections["google_drive"] {
		t.Fatalf("expected google_drive connected: %#v", state.Connections)
	}
	if len(state.Skills) < 1 {
		t.Fatalf("expected at least 1 skill, got: %#v", state.Skills)
	}
}

func TestReadTrayStateDefaultsWhenMissing(t *testing.T) {
	root := t.TempDir()
	state, err := ReadTrayState(root)
	if err != nil {
		t.Fatalf("read tray state: %v", err)
	}
	if state.Connections["google_drive"] != false {
		t.Fatalf("expected google_drive=false, got: %#v", state.Connections)
	}
	if len(state.Skills) != 0 {
		t.Fatalf("expected no skills, got: %#v", state.Skills)
	}
}

func TestReadTrayStateAddsMissingConnectionKey(t *testing.T) {
	root := t.TempDir()
	path := trayStatePath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(TrayState{UpdatedAt: "2026-02-13T00:00:00Z", Skills: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	state, err := ReadTrayState(root)
	if err != nil {
		t.Fatalf("read tray state: %v", err)
	}
	if state.Connections["google_drive"] != false {
		t.Fatalf("expected google_drive=false, got: %#v", state.Connections)
	}
}

func TestWriteTrayStatePersistsState(t *testing.T) {
	root := t.TempDir()
	state := TrayState{
		UpdatedAt:   "2026-02-13T00:00:00Z",
		Skills:      []string{"roadmap-reader"},
		Connections: map[string]bool{"google_drive": true},
	}
	if err := WriteTrayState(root, state); err != nil {
		t.Fatalf("write tray state failed: %v", err)
	}
	loaded, err := ReadTrayState(root)
	if err != nil {
		t.Fatalf("read tray state failed: %v", err)
	}
	if len(loaded.Skills) != 1 || loaded.Skills[0] != "roadmap-reader" {
		t.Fatalf("unexpected skills: %#v", loaded.Skills)
	}
	if !loaded.Connections["google_drive"] {
		t.Fatalf("expected google_drive=true, got: %#v", loaded.Connections)
	}
}

func TestReadTrayStateRejectsInvalidJSON(t *testing.T) {
	root := t.TempDir()
	path := trayStatePath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadTrayState(root); err == nil {
		t.Fatal("expected error for invalid json")
	}
}

package rollout

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStoreSaveLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rollout", "plan.json")
	store := NewStore(path)

	want := Plan{BundleName: "support", Targets: []string{"team-a", "team-b"}}
	if err := store.Save(want); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.BundleName != want.BundleName || len(got.Targets) != len(want.Targets) {
		t.Fatalf("unexpected plan: %#v", got)
	}
}

// AC4: Rollout plan persists as valid JSON to disk.
func TestStoreSave_ProducesValidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rollout", "plan.json")
	store := NewStore(path)
	plan := Plan{BundleName: "test-bundle", Targets: []string{"team-x"}}
	if err := store.Save(plan); err != nil {
		t.Fatalf("save: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var parsed Plan
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("persisted plan is not valid JSON: %v", err)
	}
	if parsed.BundleName != "test-bundle" {
		t.Fatalf("expected 'test-bundle', got %q", parsed.BundleName)
	}
}

// AC4: Store rejects empty path.
func TestStore_RejectsEmptyPath(t *testing.T) {
	store := NewStore("")
	if err := store.Save(Plan{BundleName: "x"}); err == nil {
		t.Fatal("expected error for empty path on save")
	}
	if _, err := store.Load(); err == nil {
		t.Fatal("expected error for empty path on load")
	}
}

// AC5: Multiple plans can be persisted and loaded independently.
func TestStore_MultiplePlans(t *testing.T) {
	dir := t.TempDir()
	plans := []struct {
		file string
		plan Plan
	}{
		{"plan-a.json", Plan{BundleName: "alpha", Targets: []string{"team-a"}}},
		{"plan-b.json", Plan{BundleName: "beta", Targets: []string{"team-b", "team-c"}}},
	}
	for _, p := range plans {
		store := NewStore(filepath.Join(dir, p.file))
		if err := store.Save(p.plan); err != nil {
			t.Fatalf("save %s: %v", p.file, err)
		}
	}
	for _, p := range plans {
		store := NewStore(filepath.Join(dir, p.file))
		loaded, err := store.Load()
		if err != nil {
			t.Fatalf("load %s: %v", p.file, err)
		}
		if loaded.BundleName != p.plan.BundleName {
			t.Fatalf("expected %q, got %q", p.plan.BundleName, loaded.BundleName)
		}
	}
}

// AC5: Load returns error for non-existent file.
func TestStoreLoad_MissingFile(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "missing.json"))
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// AC4/AC5: Round-trip preserves all plan fields.
func TestStore_RoundTripPreservesFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "roundtrip.json")
	store := NewStore(path)
	original := Plan{
		BundleName: "full-rollout",
		Targets:    []string{"canary", "early-adopters", "general"},
	}
	if err := store.Save(original); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.BundleName != original.BundleName {
		t.Fatalf("bundle name mismatch: %q vs %q", loaded.BundleName, original.BundleName)
	}
	if len(loaded.Targets) != len(original.Targets) {
		t.Fatalf("targets length mismatch: %d vs %d", len(loaded.Targets), len(original.Targets))
	}
	for i, target := range loaded.Targets {
		if target != original.Targets[i] {
			t.Fatalf("target[%d] mismatch: %q vs %q", i, target, original.Targets[i])
		}
	}
}

// AC5: Load rejects invalid JSON.
func TestStoreLoad_RejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	store := NewStore(path)
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parse rollout plan") {
		t.Fatalf("unexpected error: %v", err)
	}
}

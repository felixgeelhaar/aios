package sync

import "testing"

func TestDetectDrift(t *testing.T) {
	e := NewEngine()
	d := e.DetectDrift(
		map[string]string{"a": "1", "b": "2"},
		map[string]string{"a": "1", "b": "3"},
	)
	if len(d) != 1 || d[0] != "b" {
		t.Fatalf("unexpected drift: %#v", d)
	}
}

// AC3: Must validate detected differences before repair — no false positives.
func TestDetectDriftNoDriftWhenEqual(t *testing.T) {
	e := NewEngine()
	d := e.DetectDrift(
		map[string]string{"a": "1", "b": "2"},
		map[string]string{"a": "1", "b": "2"},
	)
	if len(d) != 0 {
		t.Fatalf("expected no drift, got %v", d)
	}
	if e.CurrentState() != "clean" {
		t.Fatalf("expected clean when no drift, got %s", e.CurrentState())
	}
}

// AC3: DetectDrift identifies missing keys as drift.
func TestDetectDriftMissingKey(t *testing.T) {
	e := NewEngine()
	d := e.DetectDrift(
		map[string]string{"a": "1", "b": "2", "c": "3"},
		map[string]string{"a": "1"},
	)
	if len(d) != 2 {
		t.Fatalf("expected 2 drifted keys, got %d: %v", len(d), d)
	}
}

// AC3: DetectDrift transitions engine to drifted state.
func TestDetectDriftTransitionsState(t *testing.T) {
	e := NewEngine()
	_ = e.DetectDrift(
		map[string]string{"a": "1"},
		map[string]string{"a": "2"},
	)
	if e.CurrentState() != "drifted" {
		t.Fatalf("expected drifted, got %s", e.CurrentState())
	}
}

// AC: EnsurePath creates directory structure.
func TestEnsurePathCreatesDir(t *testing.T) {
	dir := t.TempDir()
	e := NewEngine()
	target := dir + "/sub/dir"
	if err := e.EnsurePath(target); err != nil {
		t.Fatalf("ensure path: %v", err)
	}
}

// AC: EnsurePath rejects empty path.
func TestEnsurePathRejectsEmpty(t *testing.T) {
	e := NewEngine()
	if err := e.EnsurePath(""); err == nil {
		t.Fatal("expected error for empty path")
	}
}

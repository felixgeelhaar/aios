package sync

import "testing"

func TestStateTransitionsOnDriftAndStable(t *testing.T) {
	e := NewEngine()
	if e.CurrentState() != "clean" {
		t.Fatalf("expected clean, got %s", e.CurrentState())
	}

	_ = e.DetectDrift(map[string]string{"a": "1"}, map[string]string{"a": "2"})
	if e.CurrentState() != "drifted" {
		t.Fatalf("expected drifted, got %s", e.CurrentState())
	}

	e.MarkRepairing()
	if e.CurrentState() != "repairing" {
		t.Fatalf("expected repairing, got %s", e.CurrentState())
	}

	_ = e.DetectDrift(map[string]string{"a": "1"}, map[string]string{"a": "1"})
	if e.CurrentState() != "clean" {
		t.Fatalf("expected clean again, got %s", e.CurrentState())
	}
}

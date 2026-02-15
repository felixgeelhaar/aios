package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/felixgeelhaar/aios/internal/observability"
)

func TestSnapshotStore_AppendAndLoadAll(t *testing.T) {
	store := fileSnapshotStore{}
	path := filepath.Join(t.TempDir(), "state", "analytics-history.json")
	if err := store.Append(path, observability.NewSnapshot(map[string]float64{
		"tracked_projects": 1,
		"healthy_links":    1,
	})); err != nil {
		t.Fatalf("append 1 failed: %v", err)
	}
	if err := store.Append(path, observability.NewSnapshot(map[string]float64{
		"tracked_projects": 3,
		"healthy_links":    2,
	})); err != nil {
		t.Fatalf("append 2 failed: %v", err)
	}
	history, err := store.LoadAll(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected two snapshots, got %d", len(history))
	}
	trend := observability.BuildTrend(history)
	if trend["delta_tracked_projects"] != float64(2) {
		t.Fatalf("unexpected trend: %#v", trend)
	}
}

// AC4: Snapshots persist locally and survive reload.
func TestSnapshotStore_Persistence(t *testing.T) {
	store := fileSnapshotStore{}
	path := filepath.Join(t.TempDir(), "analytics", "snapshots.json")
	if err := store.Append(path, observability.NewSnapshot(map[string]float64{
		"tracked_projects": 5,
		"healthy_links":    4,
		"skill_executions": 100,
	})); err != nil {
		t.Fatalf("append: %v", err)
	}

	loaded, err := store.LoadAll(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(loaded))
	}
	if loaded[0].Metrics["skill_executions"] != 100 {
		t.Fatalf("expected 100 skill_executions, got %f", loaded[0].Metrics["skill_executions"])
	}
}

// AC4: Load from non-existent file returns empty slice (not error).
func TestSnapshotStore_NonExistentReturnsEmpty(t *testing.T) {
	store := fileSnapshotStore{}
	snapshots, err := store.LoadAll(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if len(snapshots) != 0 {
		t.Fatalf("expected empty slice, got %d snapshots", len(snapshots))
	}
}

// AC6: Snapshots are machine-readable JSON.
func TestSnapshotStore_JSONExportable(t *testing.T) {
	store := fileSnapshotStore{}
	path := filepath.Join(t.TempDir(), "export.json")
	if err := store.Append(path, observability.NewSnapshot(map[string]float64{
		"tracked_projects": 10,
		"skill_executions": 500,
	})); err != nil {
		t.Fatalf("append: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var parsed []observability.Snapshot
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("exported data is not valid JSON: %v", err)
	}
	if len(parsed) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(parsed))
	}
	if parsed[0].RecordedAt == "" {
		t.Fatal("snapshot missing recorded_at timestamp")
	}
}

// AC7: Drift metrics can be tracked in snapshots.
func TestSnapshotStore_DriftMetrics(t *testing.T) {
	store := fileSnapshotStore{}
	path := filepath.Join(t.TempDir(), "drift.json")
	if err := store.Append(path, observability.NewSnapshot(map[string]float64{
		"drift_incidents":     3,
		"drift_auto_resolved": 2,
	})); err != nil {
		t.Fatalf("append: %v", err)
	}
	history, err := store.LoadAll(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if history[0].Metrics["drift_incidents"] != 3 {
		t.Fatalf("expected 3 drift incidents, got %f", history[0].Metrics["drift_incidents"])
	}
	if history[0].Metrics["drift_auto_resolved"] != 2 {
		t.Fatalf("expected 2 auto-resolved, got %f", history[0].Metrics["drift_auto_resolved"])
	}
}

package observability

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAppendLoadAndTrend(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "analytics-history.json")
	if err := AppendSnapshot(path, map[string]float64{
		"tracked_projects": 1,
		"healthy_links":    1,
	}); err != nil {
		t.Fatalf("append 1 failed: %v", err)
	}
	if err := AppendSnapshot(path, map[string]float64{
		"tracked_projects": 3,
		"healthy_links":    2,
	}); err != nil {
		t.Fatalf("append 2 failed: %v", err)
	}
	history, err := LoadSnapshots(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected two snapshots, got %d", len(history))
	}
	trend := BuildTrend(history)
	if trend["delta_tracked_projects"] != float64(2) {
		t.Fatalf("unexpected trend: %#v", trend)
	}
}

// AC4: Snapshots persist locally and survive reload.
func TestSnapshotPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "analytics", "snapshots.json")
	metrics := map[string]float64{
		"tracked_projects": 5,
		"healthy_links":    4,
		"skill_executions": 100,
	}
	if err := AppendSnapshot(path, metrics); err != nil {
		t.Fatalf("append: %v", err)
	}

	loaded, err := LoadSnapshots(path)
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
func TestLoadSnapshots_NonExistentReturnsEmpty(t *testing.T) {
	snapshots, err := LoadSnapshots(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if len(snapshots) != 0 {
		t.Fatalf("expected empty slice, got %d snapshots", len(snapshots))
	}
}

// AC3: Adoption metrics tracked across projects.
func TestSnapshotTracksAdoptionMetrics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "adoption.json")
	if err := AppendSnapshot(path, map[string]float64{
		"tracked_projects": 2,
		"healthy_links":    2,
	}); err != nil {
		t.Fatalf("append 1: %v", err)
	}
	if err := AppendSnapshot(path, map[string]float64{
		"tracked_projects": 5,
		"healthy_links":    4,
	}); err != nil {
		t.Fatalf("append 2: %v", err)
	}

	history, err := LoadSnapshots(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	trend := BuildTrend(history)
	// Adoption grew by 3 projects.
	delta := trend["delta_tracked_projects"].(float64)
	if delta != 3 {
		t.Fatalf("expected adoption delta of 3, got %f", delta)
	}
}

// AC5: BuildTrend supports trend reporting over time.
func TestBuildTrend_MultiplePoints(t *testing.T) {
	history := []Snapshot{
		{RecordedAt: "2026-02-01T00:00:00Z", Metrics: map[string]float64{"tracked_projects": 1, "healthy_links": 1}},
		{RecordedAt: "2026-02-07T00:00:00Z", Metrics: map[string]float64{"tracked_projects": 3, "healthy_links": 2}},
		{RecordedAt: "2026-02-13T00:00:00Z", Metrics: map[string]float64{"tracked_projects": 5, "healthy_links": 5}},
	}
	trend := BuildTrend(history)
	if trend["points"] != 3 {
		t.Fatalf("expected 3 points, got %v", trend["points"])
	}
	if trend["delta_tracked_projects"] != float64(4) {
		t.Fatalf("expected delta of 4 projects, got %v", trend["delta_tracked_projects"])
	}
	if trend["delta_healthy_links"] != float64(4) {
		t.Fatalf("expected delta of 4 healthy links, got %v", trend["delta_healthy_links"])
	}
}

// AC5: BuildTrend handles empty history.
func TestBuildTrend_EmptyHistory(t *testing.T) {
	trend := BuildTrend(nil)
	if trend["points"] != 0 {
		t.Fatalf("expected 0 points, got %v", trend["points"])
	}
	if trend["delta_tracked_projects"] != 0.0 {
		t.Fatalf("expected zero delta, got %v", trend["delta_tracked_projects"])
	}
}

// AC5: BuildTrend with single point has zero delta.
func TestBuildTrend_SinglePoint(t *testing.T) {
	history := []Snapshot{
		{RecordedAt: "2026-02-13T00:00:00Z", Metrics: map[string]float64{"tracked_projects": 10}},
	}
	trend := BuildTrend(history)
	if trend["points"] != 1 {
		t.Fatalf("expected 1 point, got %v", trend["points"])
	}
	if trend["delta_tracked_projects"] != float64(0) {
		t.Fatalf("expected zero delta for single point, got %v", trend["delta_tracked_projects"])
	}
}

// AC6: Snapshots are machine-readable JSON.
func TestSnapshotIsJSONExportable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "export.json")
	if err := AppendSnapshot(path, map[string]float64{
		"tracked_projects": 10,
		"skill_executions": 500,
	}); err != nil {
		t.Fatalf("append: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var parsed []Snapshot
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

// AC6: BuildTrend output is JSON-serializable.
func TestBuildTrendOutputIsJSONSerializable(t *testing.T) {
	history := []Snapshot{
		{RecordedAt: "2026-02-01T00:00:00Z", Metrics: map[string]float64{"tracked_projects": 1}},
		{RecordedAt: "2026-02-13T00:00:00Z", Metrics: map[string]float64{"tracked_projects": 5}},
	}
	trend := BuildTrend(history)
	data, err := json.Marshal(trend)
	if err != nil {
		t.Fatalf("trend not JSON-serializable: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty JSON output")
	}
}

// AC7: Drift metrics can be tracked in snapshots.
func TestSnapshotTracksDriftMetrics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drift.json")
	if err := AppendSnapshot(path, map[string]float64{
		"drift_incidents":     3,
		"drift_auto_resolved": 2,
	}); err != nil {
		t.Fatalf("append: %v", err)
	}
	history, err := LoadSnapshots(path)
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

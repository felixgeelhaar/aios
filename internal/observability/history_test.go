package observability

import (
	"encoding/json"
	"testing"
)

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

func TestNewSnapshot(t *testing.T) {
	metrics := map[string]float64{"tracked_projects": 5}
	s := NewSnapshot(metrics)
	if s.RecordedAt == "" {
		t.Fatal("expected non-empty RecordedAt")
	}
	if s.Metrics["tracked_projects"] != 5 {
		t.Fatalf("expected 5 tracked_projects, got %f", s.Metrics["tracked_projects"])
	}
}

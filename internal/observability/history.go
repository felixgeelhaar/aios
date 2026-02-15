package observability

import (
	"time"
)

type Snapshot struct {
	RecordedAt string             `json:"recorded_at"`
	Metrics    map[string]float64 `json:"metrics"`
}

// SnapshotStore abstracts persistence of observability snapshots.
type SnapshotStore interface {
	Append(path string, snapshot Snapshot) error
	LoadAll(path string) ([]Snapshot, error)
}

// NewSnapshot constructs a Snapshot with the current timestamp.
func NewSnapshot(metrics map[string]float64) Snapshot {
	return Snapshot{
		RecordedAt: time.Now().UTC().Format(time.RFC3339),
		Metrics:    metrics,
	}
}

func BuildTrend(history []Snapshot) map[string]any {
	out := map[string]any{
		"points": len(history),
	}
	if len(history) == 0 {
		out["delta_tracked_projects"] = 0.0
		out["delta_healthy_links"] = 0.0
		return out
	}
	latest := history[len(history)-1]
	out["latest"] = latest
	out["delta_tracked_projects"] = metricDelta(history, "tracked_projects")
	out["delta_healthy_links"] = metricDelta(history, "healthy_links")
	return out
}

func metricDelta(history []Snapshot, key string) float64 {
	if len(history) < 2 {
		return 0
	}
	first := history[0].Metrics[key]
	last := history[len(history)-1].Metrics[key]
	return last - first
}

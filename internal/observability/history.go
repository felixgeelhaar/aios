package observability

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Snapshot struct {
	RecordedAt string             `json:"recorded_at"`
	Metrics    map[string]float64 `json:"metrics"`
}

func AppendSnapshot(path string, metrics map[string]float64) error {
	history, err := LoadSnapshots(path)
	if err != nil {
		return err
	}
	history = append(history, Snapshot{
		RecordedAt: time.Now().UTC().Format(time.RFC3339),
		Metrics:    metrics,
	})
	body, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o600)
}

func LoadSnapshots(path string) ([]Snapshot, error) {
	path = filepath.Clean(path)
	// #nosec G304 -- path is managed by runtime workspace configuration.
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Snapshot{}, nil
		}
		return nil, err
	}
	if len(body) == 0 {
		return []Snapshot{}, nil
	}
	var history []Snapshot
	if err := json.Unmarshal(body, &history); err != nil {
		return nil, err
	}
	return history, nil
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

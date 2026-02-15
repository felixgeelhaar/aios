package core

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/felixgeelhaar/aios/internal/observability"
)

// fileSnapshotStore implements observability.SnapshotStore using the local
// filesystem for reading and writing analytics snapshots.
type fileSnapshotStore struct{}

var _ observability.SnapshotStore = fileSnapshotStore{}

func (fileSnapshotStore) Append(path string, snapshot observability.Snapshot) error {
	history, err := (fileSnapshotStore{}).LoadAll(path)
	if err != nil {
		return err
	}
	history = append(history, snapshot)
	body, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o600)
}

func (fileSnapshotStore) LoadAll(path string) ([]observability.Snapshot, error) {
	path = filepath.Clean(path)
	// #nosec G304 -- path is managed by runtime workspace configuration.
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []observability.Snapshot{}, nil
		}
		return nil, err
	}
	if len(body) == 0 {
		return []observability.Snapshot{}, nil
	}
	var history []observability.Snapshot
	if err := json.Unmarshal(body, &history); err != nil {
		return nil, err
	}
	return history, nil
}

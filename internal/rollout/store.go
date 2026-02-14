package rollout

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct {
	path string
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Save(p Plan) error {
	if s.path == "" {
		return fmt.Errorf("path is required")
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o750); err != nil {
		return fmt.Errorf("create rollout dir: %w", err)
	}
	body, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal rollout plan: %w", err)
	}
	if err := os.WriteFile(s.path, body, 0o600); err != nil {
		return fmt.Errorf("write rollout plan: %w", err)
	}
	return nil
}

func (s *Store) Load() (Plan, error) {
	if s.path == "" {
		return Plan{}, fmt.Errorf("path is required")
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return Plan{}, fmt.Errorf("read rollout plan: %w", err)
	}
	var p Plan
	if err := json.Unmarshal(data, &p); err != nil {
		return Plan{}, fmt.Errorf("parse rollout plan: %w", err)
	}
	return p, nil
}

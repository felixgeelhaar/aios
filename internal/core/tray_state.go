package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/felixgeelhaar/aios/internal/agents"
)

type TrayState struct {
	UpdatedAt   string          `json:"updated_at"`
	Skills      []string        `json:"skills"`
	Connections map[string]bool `json:"connections"`
}

func trayStatePath(workspace string) string {
	return filepath.Join(workspace, "tray", "state.json")
}

func ReadTrayState(workspace string) (TrayState, error) {
	path := filepath.Clean(trayStatePath(workspace))
	// #nosec G304 -- path is derived from configured workspace directory.
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return TrayState{
				UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				Skills:      []string{},
				Connections: map[string]bool{"google_drive": false},
			}, nil
		}
		return TrayState{}, err
	}
	var s TrayState
	if err := json.Unmarshal(body, &s); err != nil {
		return TrayState{}, err
	}
	if s.Connections == nil {
		s.Connections = map[string]bool{"google_drive": false}
	}
	if _, ok := s.Connections["google_drive"]; !ok {
		s.Connections["google_drive"] = false
	}
	return s, nil
}

func WriteTrayState(workspace string, state TrayState) error {
	path := trayStatePath(workspace)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	body, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o600)
}

func RefreshTrayState(cfg Config, googleDriveOverride *bool) (TrayState, error) {
	state, err := ReadTrayState(cfg.WorkspaceDir)
	if err != nil {
		return TrayState{}, err
	}
	if googleDriveOverride != nil {
		state.Connections["google_drive"] = *googleDriveOverride
	}
	skills, err := collectInstalledSkills(cfg)
	if err != nil {
		return TrayState{}, err
	}
	state.Skills = skills
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := WriteTrayState(cfg.WorkspaceDir, state); err != nil {
		return TrayState{}, err
	}
	return state, nil
}

func collectInstalledSkills(cfg Config) ([]string, error) {
	allAgents, err := agents.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("loading agents: %w", err)
	}
	si := agents.NewSkillInstaller(allAgents)
	return si.CollectInstalledSkills(cfg.ProjectDir)
}

package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/felixgeelhaar/aios/internal/agents"
	"github.com/felixgeelhaar/aios/internal/domain/agentregistry"
)

func RestoreClientConfigs(cfg Config, backupDir string) error {
	if backupDir == "" {
		return fmt.Errorf("backup directory is required")
	}

	allAgents, err := agents.LoadAll()
	if err != nil {
		return fmt.Errorf("loading agents: %w", err)
	}

	// Restore canonical skills directory.
	canonicalDst := filepath.Join(cfg.ProjectDir, agentregistry.CanonicalSkillsDir)
	if err := copyDir(filepath.Join(backupDir, "skills"), canonicalDst); err != nil {
		return err
	}

	// Restore each non-universal agent's skill directory.
	for _, agent := range agentregistry.FilterNonUniversal(allAgents) {
		src := filepath.Join(backupDir, agent.Name)
		dst := filepath.Join(cfg.ProjectDir, agent.SkillsDir)
		if err := copyDir(src, dst); err != nil {
			return err
		}
	}

	return nil
}

func LatestBackupDir(workspace string) (string, error) {
	root := filepath.Join(workspace, "backups")
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	latest := ""
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if e.Name() > latest {
			latest = e.Name()
		}
	}
	if latest == "" {
		return "", fmt.Errorf("no backups found")
	}
	return filepath.Join(root, latest), nil
}

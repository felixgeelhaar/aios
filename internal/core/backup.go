package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/felixgeelhaar/aios/internal/agents"
	"github.com/felixgeelhaar/aios/internal/domain/agentregistry"
)

func BackupClientConfigs(cfg Config) (string, error) {
	stamp := time.Now().UTC().Format("20060102T150405Z")
	root := filepath.Join(cfg.WorkspaceDir, "backups", stamp)
	if err := os.MkdirAll(root, 0o750); err != nil {
		return "", err
	}

	allAgents, err := agents.LoadAll()
	if err != nil {
		return "", fmt.Errorf("loading agents: %w", err)
	}

	// Backup canonical skills directory.
	canonicalSrc := filepath.Join(cfg.ProjectDir, agentregistry.CanonicalSkillsDir)
	if err := copyDir(canonicalSrc, filepath.Join(root, "skills")); err != nil {
		return "", err
	}

	// Backup each non-universal agent's skill directory.
	for _, agent := range agentregistry.FilterNonUniversal(allAgents) {
		src := filepath.Join(cfg.ProjectDir, agent.SkillsDir)
		if err := copyDir(src, filepath.Join(root, agent.Name)); err != nil {
			return "", err
		}
	}

	return root, nil
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o750); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		s := filepath.Join(src, e.Name())
		d := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(s, d); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(s, d); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	// #nosec G304 -- src path originates from traversing existing config directories.
	in, err := os.Open(filepath.Clean(src))
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(filepath.Clean(dst), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	if err := out.Sync(); err != nil {
		return fmt.Errorf("sync %s: %w", dst, err)
	}
	return nil
}

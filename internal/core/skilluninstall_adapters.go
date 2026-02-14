package core

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/felixgeelhaar/aios/internal/agents"
	domain "github.com/felixgeelhaar/aios/internal/domain/skilluninstall"
	"github.com/felixgeelhaar/aios/internal/skill"
)

type uninstallSkillIDResolverAdapter struct{}

func (uninstallSkillIDResolverAdapter) ResolveSkillID(skillDir string) (string, error) {
	spec, err := skill.LoadSkillSpec(filepath.Join(skillDir, "skill.yaml"))
	if err != nil {
		return "", err
	}
	return spec.ID, nil
}

type clientUninstallerAdapter struct {
	cfg Config
}

func (a clientUninstallerAdapter) UninstallAcrossClients(_ context.Context, skillID string) error {
	allAgents, err := agents.LoadAll()
	if err != nil {
		return fmt.Errorf("loading agents: %w", err)
	}
	si := agents.NewSkillInstaller(allAgents)
	return si.UninstallSkill(skillID, a.cfg.ProjectDir)
}

var _ domain.SkillIDResolver = uninstallSkillIDResolverAdapter{}
var _ domain.ClientUninstaller = clientUninstallerAdapter{}

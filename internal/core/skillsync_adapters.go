package core

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/felixgeelhaar/aios/internal/agents"
	domain "github.com/felixgeelhaar/aios/internal/domain/skillsync"
	"github.com/felixgeelhaar/aios/internal/skill"
)

type skillSpecResolverAdapter struct{}

func (skillSpecResolverAdapter) ResolveSkillID(skillDir string) (string, error) {
	specPath := filepath.Join(skillDir, "skill.yaml")
	spec, err := skill.LoadSkillSpec(specPath)
	if err != nil {
		return "", err
	}
	if err := skill.ValidateSkillSpec(skillDir, spec); err != nil {
		return "", err
	}
	return spec.ID, nil
}

type clientInstallerAdapter struct {
	cfg Config
}

func (a clientInstallerAdapter) InstallSkillAcrossClients(_ context.Context, skillID string, skillDir string) error {
	allAgents, err := agents.LoadAll()
	if err != nil {
		return fmt.Errorf("loading agents: %w", err)
	}
	// Compose rich SKILL.md from skill.yaml + prompt.md.
	skillContent, _ := skill.LoadAndBuildSkillMd(skillDir)
	si := agents.NewSkillInstaller(allAgents)
	_, installErr := si.InstallSkill(skillID, agents.InstallOptions{
		ProjectDir:   a.cfg.ProjectDir,
		SkillContent: skillContent,
	})
	return installErr
}

var _ domain.SkillSpecResolver = skillSpecResolverAdapter{}
var _ domain.ClientInstaller = clientInstallerAdapter{}

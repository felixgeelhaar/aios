package core

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/felixgeelhaar/aios/internal/agents"
	domain "github.com/felixgeelhaar/aios/internal/domain/syncplan"
	"github.com/felixgeelhaar/aios/internal/skill"
)

type syncPlanSkillResolverAdapter struct{}

func (syncPlanSkillResolverAdapter) ResolveSkillID(skillDir string) (string, error) {
	spec, err := skill.LoadSkillSpec(filepath.Join(skillDir, "skill.yaml"))
	if err != nil {
		return "", err
	}
	if err := skill.ValidateSkillSpec(skillDir, spec); err != nil {
		return "", err
	}
	return spec.ID, nil
}

type syncPlanWriteTargetPlannerAdapter struct {
	cfg Config
}

func (a syncPlanWriteTargetPlannerAdapter) PlanWriteTargets(_ context.Context, skillID string) ([]string, error) {
	allAgents, err := agents.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("loading agents: %w", err)
	}
	si := agents.NewSkillInstaller(allAgents)
	return si.PlanWriteTargets(skillID, a.cfg.ProjectDir), nil
}

var _ domain.SkillIDResolver = syncPlanSkillResolverAdapter{}
var _ domain.WriteTargetPlanner = syncPlanWriteTargetPlannerAdapter{}

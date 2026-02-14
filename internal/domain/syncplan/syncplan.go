package syncplan

import (
	"context"
	"fmt"
	"strings"
)

var ErrSkillDirRequired = fmt.Errorf("skill-dir is required")

type BuildSyncPlanCommand struct {
	SkillDir string
}

type BuildSyncPlanResult struct {
	SkillID string   `json:"skill_id"`
	Writes  []string `json:"writes"`
}

type SkillIDResolver interface {
	ResolveSkillID(skillDir string) (string, error)
}

type WriteTargetPlanner interface {
	PlanWriteTargets(ctx context.Context, skillID string) ([]string, error)
}

func (c BuildSyncPlanCommand) Normalized() BuildSyncPlanCommand {
	return BuildSyncPlanCommand{SkillDir: strings.TrimSpace(c.SkillDir)}
}

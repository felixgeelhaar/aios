package syncplan

import (
	"context"

	domain "github.com/felixgeelhaar/aios/internal/domain/syncplan"
)

type Service struct {
	resolver domain.SkillIDResolver
	planner  domain.WriteTargetPlanner
}

func NewService(resolver domain.SkillIDResolver, planner domain.WriteTargetPlanner) Service {
	return Service{
		resolver: resolver,
		planner:  planner,
	}
}

func (s Service) BuildSyncPlan(ctx context.Context, command domain.BuildSyncPlanCommand) (domain.BuildSyncPlanResult, error) {
	cmd := command.Normalized()
	if cmd.SkillDir == "" {
		return domain.BuildSyncPlanResult{}, domain.ErrSkillDirRequired
	}
	skillID, err := s.resolver.ResolveSkillID(cmd.SkillDir)
	if err != nil {
		return domain.BuildSyncPlanResult{}, err
	}
	writes, err := s.planner.PlanWriteTargets(ctx, skillID)
	if err != nil {
		return domain.BuildSyncPlanResult{}, err
	}
	return domain.BuildSyncPlanResult{
		SkillID: skillID,
		Writes:  writes,
	}, nil
}

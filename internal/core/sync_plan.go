package core

import (
	"context"

	applicationsyncplan "github.com/felixgeelhaar/aios/internal/application/syncplan"
	domainsyncplan "github.com/felixgeelhaar/aios/internal/domain/syncplan"
)

type SyncPlan struct {
	SkillID string
	Writes  []string
}

func BuildSyncPlan(cfg Config, skillDir string) (SyncPlan, error) {
	service := applicationsyncplan.NewService(
		syncPlanSkillResolverAdapter{},
		syncPlanWriteTargetPlannerAdapter{cfg: cfg},
	)
	result, err := service.BuildSyncPlan(context.Background(), domainsyncplan.BuildSyncPlanCommand{SkillDir: skillDir})
	if err != nil {
		return SyncPlan{}, err
	}
	return SyncPlan{
		SkillID: result.SkillID,
		Writes:  result.Writes,
	}, nil
}

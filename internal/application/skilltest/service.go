package skilltest

import (
	"context"

	domain "github.com/felixgeelhaar/aios/internal/domain/skilltest"
)

type Service struct {
	runner domain.FixtureRunner
}

func NewService(runner domain.FixtureRunner) Service {
	return Service{runner: runner}
}

func (s Service) TestSkill(ctx context.Context, command domain.TestSkillCommand) (domain.TestSkillResult, error) {
	cmd := command.Normalized()
	if cmd.SkillDir == "" {
		return domain.TestSkillResult{}, domain.ErrSkillDirRequired
	}
	results, err := s.runner.Run(ctx, cmd.SkillDir)
	if err != nil {
		return domain.TestSkillResult{}, err
	}
	failed := 0
	for _, r := range results {
		if !r.Passed {
			failed++
		}
	}
	return domain.TestSkillResult{
		Results: results,
		Failed:  failed,
	}, nil
}

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
	if err := cmd.Validate(); err != nil {
		return domain.TestSkillResult{}, err
	}
	results, err := s.runner.Run(ctx, cmd.SkillDir)
	if err != nil {
		return domain.TestSkillResult{}, err
	}
	return domain.NewTestSkillResult(results), nil
}

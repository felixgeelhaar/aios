package core

import (
	"context"

	domain "github.com/felixgeelhaar/aios/internal/domain/skilltest"
	"github.com/felixgeelhaar/aios/internal/skill"
)

type fixtureRunnerAdapter struct{}

func (fixtureRunnerAdapter) Run(_ context.Context, skillDir string) ([]domain.FixtureResult, error) {
	results, err := skill.RunFixtureSuite(skillDir)
	if err != nil {
		return nil, err
	}
	out := make([]domain.FixtureResult, 0, len(results))
	for _, r := range results {
		out = append(out, domain.FixtureResult{
			Name:   r.Name,
			Passed: r.Passed,
			Error:  r.Error,
		})
	}
	return out, nil
}

var _ domain.FixtureRunner = fixtureRunnerAdapter{}

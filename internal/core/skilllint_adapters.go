package core

import (
	"context"

	domain "github.com/felixgeelhaar/aios/internal/domain/skilllint"
	"github.com/felixgeelhaar/aios/internal/skill"
)

type skillLinterAdapter struct{}

func (skillLinterAdapter) Lint(_ context.Context, skillDir string) (domain.LintSkillResult, error) {
	res, err := skill.LintSkillDir(skillDir)
	if err != nil {
		return domain.LintSkillResult{}, err
	}
	return domain.LintSkillResult{
		Valid:  res.Valid,
		Issues: res.Issues,
	}, nil
}

var _ domain.SkillLinter = skillLinterAdapter{}

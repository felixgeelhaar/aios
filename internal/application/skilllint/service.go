package skilllint

import (
	"context"

	domain "github.com/felixgeelhaar/aios/internal/domain/skilllint"
)

type Service struct {
	linter domain.SkillLinter
}

func NewService(linter domain.SkillLinter) Service {
	return Service{linter: linter}
}

func (s Service) LintSkill(ctx context.Context, command domain.LintSkillCommand) (domain.LintSkillResult, error) {
	cmd := command.Normalized()
	if cmd.SkillDir == "" {
		return domain.LintSkillResult{}, domain.ErrSkillDirRequired
	}
	return s.linter.Lint(ctx, cmd.SkillDir)
}

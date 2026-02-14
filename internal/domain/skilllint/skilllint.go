package skilllint

import (
	"context"
	"fmt"
	"strings"
)

var ErrSkillDirRequired = fmt.Errorf("skill-dir is required")

type LintSkillCommand struct {
	SkillDir string
}

type LintSkillResult struct {
	Valid  bool
	Issues []string
}

type SkillLinter interface {
	Lint(ctx context.Context, skillDir string) (LintSkillResult, error)
}

func (c LintSkillCommand) Normalized() LintSkillCommand {
	return LintSkillCommand{SkillDir: strings.TrimSpace(c.SkillDir)}
}

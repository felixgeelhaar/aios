package skilltest

import (
	"context"
	"fmt"
	"strings"
)

var ErrSkillDirRequired = fmt.Errorf("skill-dir is required")

type TestSkillCommand struct {
	SkillDir string
}

type FixtureResult struct {
	Name   string
	Passed bool
	Error  string
}

type TestSkillResult struct {
	Results []FixtureResult
	Failed  int
}

type FixtureRunner interface {
	Run(ctx context.Context, skillDir string) ([]FixtureResult, error)
}

func (c TestSkillCommand) Normalized() TestSkillCommand {
	return TestSkillCommand{SkillDir: strings.TrimSpace(c.SkillDir)}
}

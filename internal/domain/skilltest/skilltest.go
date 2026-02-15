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

// Validate checks that the command has all required fields.
func (c TestSkillCommand) Validate() error {
	if c.SkillDir == "" {
		return ErrSkillDirRequired
	}
	return nil
}

// NewTestSkillResult constructs a TestSkillResult from fixture results,
// computing the failed count from the results.
func NewTestSkillResult(results []FixtureResult) TestSkillResult {
	failed := 0
	for _, r := range results {
		if !r.Passed {
			failed++
		}
	}
	return TestSkillResult{
		Results: results,
		Failed:  failed,
	}
}

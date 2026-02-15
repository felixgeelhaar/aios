package skilltest_test

import (
	"testing"

	"github.com/felixgeelhaar/aios/internal/domain/skilltest"
)

func TestValidate_EmptySkillDir(t *testing.T) {
	cmd := skilltest.TestSkillCommand{}.Normalized()
	if err := cmd.Validate(); err != skilltest.ErrSkillDirRequired {
		t.Errorf("expected ErrSkillDirRequired, got %v", err)
	}
}

func TestValidate_WhitespaceSkillDir(t *testing.T) {
	cmd := skilltest.TestSkillCommand{SkillDir: "   "}.Normalized()
	if err := cmd.Validate(); err != skilltest.ErrSkillDirRequired {
		t.Errorf("expected ErrSkillDirRequired, got %v", err)
	}
}

func TestValidate_ValidSkillDir(t *testing.T) {
	cmd := skilltest.TestSkillCommand{SkillDir: "/path/to/skill"}.Normalized()
	if err := cmd.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewTestSkillResult_AllPassed(t *testing.T) {
	results := []skilltest.FixtureResult{
		{Name: "test1", Passed: true},
		{Name: "test2", Passed: true},
	}
	result := skilltest.NewTestSkillResult(results)
	if result.Failed != 0 {
		t.Errorf("expected 0 failures, got %d", result.Failed)
	}
	if len(result.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(result.Results))
	}
}

func TestNewTestSkillResult_SomeFailed(t *testing.T) {
	results := []skilltest.FixtureResult{
		{Name: "test1", Passed: true},
		{Name: "test2", Passed: false, Error: "assertion failed"},
		{Name: "test3", Passed: false, Error: "timeout"},
	}
	result := skilltest.NewTestSkillResult(results)
	if result.Failed != 2 {
		t.Errorf("expected 2 failures, got %d", result.Failed)
	}
}

func TestNewTestSkillResult_Empty(t *testing.T) {
	result := skilltest.NewTestSkillResult(nil)
	if result.Failed != 0 {
		t.Errorf("expected 0 failures, got %d", result.Failed)
	}
	if result.Results != nil {
		t.Errorf("expected nil results, got %v", result.Results)
	}
}

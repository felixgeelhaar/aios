package skilltest

import (
	"context"
	"errors"
	"testing"

	domain "github.com/felixgeelhaar/aios/internal/domain/skilltest"
)

type fakeFixtureRunner struct {
	results []domain.FixtureResult
	err     error
}

func (f fakeFixtureRunner) Run(context.Context, string) ([]domain.FixtureResult, error) {
	return f.results, f.err
}

func TestServiceTestSkill(t *testing.T) {
	svc := NewService(fakeFixtureRunner{
		results: []domain.FixtureResult{
			{Name: "fixture_01.json", Passed: true},
			{Name: "fixture_02.json", Passed: false},
		},
	})
	res, err := svc.TestSkill(context.Background(), domain.TestSkillCommand{SkillDir: "/tmp/skill"})
	if err != nil {
		t.Fatalf("test-skill failed: %v", err)
	}
	if res.Failed != 1 {
		t.Fatalf("unexpected failed count: %d", res.Failed)
	}
}

func TestServiceTestSkillRequiresSkillDir(t *testing.T) {
	svc := NewService(fakeFixtureRunner{})
	_, err := svc.TestSkill(context.Background(), domain.TestSkillCommand{})
	if !errors.Is(err, domain.ErrSkillDirRequired) {
		t.Fatalf("expected skill-dir required error, got %v", err)
	}
}

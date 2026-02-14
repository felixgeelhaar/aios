package skilllint

import (
	"context"
	"errors"
	"testing"

	domain "github.com/felixgeelhaar/aios/internal/domain/skilllint"
)

type fakeLinter struct {
	result domain.LintSkillResult
	err    error
}

func (f fakeLinter) Lint(context.Context, string) (domain.LintSkillResult, error) {
	return f.result, f.err
}

func TestServiceLintSkill(t *testing.T) {
	svc := NewService(fakeLinter{result: domain.LintSkillResult{Valid: true}})
	res, err := svc.LintSkill(context.Background(), domain.LintSkillCommand{SkillDir: "/tmp/skill"})
	if err != nil {
		t.Fatalf("lint failed: %v", err)
	}
	if !res.Valid {
		t.Fatalf("unexpected invalid result: %#v", res)
	}
}

func TestServiceLintSkillRequiresSkillDir(t *testing.T) {
	svc := NewService(fakeLinter{})
	_, err := svc.LintSkill(context.Background(), domain.LintSkillCommand{})
	if !errors.Is(err, domain.ErrSkillDirRequired) {
		t.Fatalf("expected skill-dir required error, got %v", err)
	}
}

package skilllint_test

import (
	"testing"

	"github.com/felixgeelhaar/aios/internal/domain/skilllint"
)

func TestValidate_EmptySkillDir(t *testing.T) {
	cmd := skilllint.LintSkillCommand{}.Normalized()
	if err := cmd.Validate(); err != skilllint.ErrSkillDirRequired {
		t.Errorf("expected ErrSkillDirRequired, got %v", err)
	}
}

func TestValidate_WhitespaceSkillDir(t *testing.T) {
	cmd := skilllint.LintSkillCommand{SkillDir: "   "}.Normalized()
	if err := cmd.Validate(); err != skilllint.ErrSkillDirRequired {
		t.Errorf("expected ErrSkillDirRequired, got %v", err)
	}
}

func TestValidate_ValidSkillDir(t *testing.T) {
	cmd := skilllint.LintSkillCommand{SkillDir: "/path/to/skill"}.Normalized()
	if err := cmd.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

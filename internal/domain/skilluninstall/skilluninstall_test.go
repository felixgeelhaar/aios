package skilluninstall_test

import (
	"testing"

	"github.com/felixgeelhaar/aios/internal/domain/skilluninstall"
)

func TestValidate_EmptySkillDir(t *testing.T) {
	cmd := skilluninstall.UninstallSkillCommand{}.Normalized()
	if err := cmd.Validate(); err != skilluninstall.ErrSkillDirRequired {
		t.Errorf("expected ErrSkillDirRequired, got %v", err)
	}
}

func TestValidate_WhitespaceSkillDir(t *testing.T) {
	cmd := skilluninstall.UninstallSkillCommand{SkillDir: "   "}.Normalized()
	if err := cmd.Validate(); err != skilluninstall.ErrSkillDirRequired {
		t.Errorf("expected ErrSkillDirRequired, got %v", err)
	}
}

func TestValidate_ValidSkillDir(t *testing.T) {
	cmd := skilluninstall.UninstallSkillCommand{SkillDir: "/path/to/skill"}.Normalized()
	if err := cmd.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

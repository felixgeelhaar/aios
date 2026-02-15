package skillsync_test

import (
	"testing"

	"github.com/felixgeelhaar/aios/internal/domain/skillsync"
)

func TestValidate_EmptySkillDir(t *testing.T) {
	cmd := skillsync.SyncSkillCommand{}.Normalized()
	if err := cmd.Validate(); err != skillsync.ErrSkillDirRequired {
		t.Errorf("expected ErrSkillDirRequired, got %v", err)
	}
}

func TestValidate_WhitespaceSkillDir(t *testing.T) {
	cmd := skillsync.SyncSkillCommand{SkillDir: "   "}.Normalized()
	if err := cmd.Validate(); err != skillsync.ErrSkillDirRequired {
		t.Errorf("expected ErrSkillDirRequired, got %v", err)
	}
}

func TestValidate_ValidSkillDir(t *testing.T) {
	cmd := skillsync.SyncSkillCommand{SkillDir: "/path/to/skill"}.Normalized()
	if err := cmd.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

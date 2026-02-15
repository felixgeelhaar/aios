package syncplan_test

import (
	"testing"

	"github.com/felixgeelhaar/aios/internal/domain/syncplan"
)

func TestValidate_EmptySkillDir(t *testing.T) {
	cmd := syncplan.BuildSyncPlanCommand{}.Normalized()
	if err := cmd.Validate(); err != syncplan.ErrSkillDirRequired {
		t.Errorf("expected ErrSkillDirRequired, got %v", err)
	}
}

func TestValidate_WhitespaceSkillDir(t *testing.T) {
	cmd := syncplan.BuildSyncPlanCommand{SkillDir: "   "}.Normalized()
	if err := cmd.Validate(); err != syncplan.ErrSkillDirRequired {
		t.Errorf("expected ErrSkillDirRequired, got %v", err)
	}
}

func TestValidate_ValidSkillDir(t *testing.T) {
	cmd := syncplan.BuildSyncPlanCommand{SkillDir: "/path/to/skill"}.Normalized()
	if err := cmd.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

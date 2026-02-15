package skillsync

import (
	"context"
	"fmt"
	"strings"
)

var ErrSkillDirRequired = fmt.Errorf("skill-dir is required")

type SyncSkillCommand struct {
	SkillDir string
}

type SkillSpecResolver interface {
	ResolveSkillID(skillDir string) (string, error)
}

type ClientInstaller interface {
	InstallSkillAcrossClients(ctx context.Context, skillID string, skillDir string) error
}

func (c SyncSkillCommand) Normalized() SyncSkillCommand {
	return SyncSkillCommand{SkillDir: strings.TrimSpace(c.SkillDir)}
}

// Validate checks that the command has all required fields.
func (c SyncSkillCommand) Validate() error {
	if c.SkillDir == "" {
		return ErrSkillDirRequired
	}
	return nil
}

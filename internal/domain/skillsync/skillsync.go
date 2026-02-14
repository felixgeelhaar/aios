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
	InstallSkillAcrossClients(ctx context.Context, skillID string) error
}

func (c SyncSkillCommand) Normalized() SyncSkillCommand {
	return SyncSkillCommand{SkillDir: strings.TrimSpace(c.SkillDir)}
}

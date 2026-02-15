package skilluninstall

import (
	"context"
	"fmt"
	"strings"
)

var ErrSkillDirRequired = fmt.Errorf("skill-dir is required")

type UninstallSkillCommand struct {
	SkillDir string
}

type SkillIDResolver interface {
	ResolveSkillID(skillDir string) (string, error)
}

type ClientUninstaller interface {
	UninstallAcrossClients(ctx context.Context, skillID string) error
}

func (c UninstallSkillCommand) Normalized() UninstallSkillCommand {
	return UninstallSkillCommand{SkillDir: strings.TrimSpace(c.SkillDir)}
}

// Validate checks that the command has all required fields.
func (c UninstallSkillCommand) Validate() error {
	if c.SkillDir == "" {
		return ErrSkillDirRequired
	}
	return nil
}

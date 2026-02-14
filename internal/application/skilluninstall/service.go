package skilluninstall

import (
	"context"

	domain "github.com/felixgeelhaar/aios/internal/domain/skilluninstall"
)

type Service struct {
	resolver    domain.SkillIDResolver
	uninstaller domain.ClientUninstaller
}

func NewService(resolver domain.SkillIDResolver, uninstaller domain.ClientUninstaller) Service {
	return Service{
		resolver:    resolver,
		uninstaller: uninstaller,
	}
}

func (s Service) UninstallSkill(ctx context.Context, command domain.UninstallSkillCommand) (string, error) {
	cmd := command.Normalized()
	if cmd.SkillDir == "" {
		return "", domain.ErrSkillDirRequired
	}
	skillID, err := s.resolver.ResolveSkillID(cmd.SkillDir)
	if err != nil {
		return "", err
	}
	if err := s.uninstaller.UninstallAcrossClients(ctx, skillID); err != nil {
		return "", err
	}
	return skillID, nil
}

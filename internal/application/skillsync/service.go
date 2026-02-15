package skillsync

import (
	"context"

	domain "github.com/felixgeelhaar/aios/internal/domain/skillsync"
)

type Service struct {
	resolver  domain.SkillSpecResolver
	installer domain.ClientInstaller
}

func NewService(resolver domain.SkillSpecResolver, installer domain.ClientInstaller) Service {
	return Service{
		resolver:  resolver,
		installer: installer,
	}
}

func (s Service) SyncSkill(ctx context.Context, command domain.SyncSkillCommand) (string, error) {
	cmd := command.Normalized()
	if err := cmd.Validate(); err != nil {
		return "", err
	}
	skillID, err := s.resolver.ResolveSkillID(cmd.SkillDir)
	if err != nil {
		return "", err
	}
	if err := s.installer.InstallSkillAcrossClients(ctx, skillID); err != nil {
		return "", err
	}
	return skillID, nil
}

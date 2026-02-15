package skillpackage

import (
	"context"
	"path/filepath"

	domain "github.com/felixgeelhaar/aios/internal/domain/skillpackage"
)

type Service struct {
	resolver domain.SkillMetadataResolver
	packager domain.SkillPackager
}

func NewService(resolver domain.SkillMetadataResolver, packager domain.SkillPackager) Service {
	return Service{
		resolver: resolver,
		packager: packager,
	}
}

func (s Service) PackageSkill(_ context.Context, command domain.PackageSkillCommand) (domain.PackageSkillResult, error) {
	cmd := command.Normalized()
	if err := cmd.Validate(); err != nil {
		return domain.PackageSkillResult{}, err
	}

	id, version, err := s.resolver.ResolveIDAndVersion(cmd.SkillDir)
	if err != nil {
		return domain.PackageSkillResult{}, err
	}
	artifactPath := filepath.Join(filepath.Dir(cmd.SkillDir), domain.ArtifactName(id, version))
	if err := s.packager.Package(cmd.SkillDir, artifactPath); err != nil {
		return domain.PackageSkillResult{}, err
	}
	return domain.PackageSkillResult{ArtifactPath: artifactPath}, nil
}

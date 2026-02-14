package core

import (
	"path/filepath"

	domain "github.com/felixgeelhaar/aios/internal/domain/skillpackage"
	"github.com/felixgeelhaar/aios/internal/skill"
)

type skillMetadataResolverAdapter struct{}

func (skillMetadataResolverAdapter) ResolveIDAndVersion(skillDir string) (string, string, error) {
	spec, err := skill.LoadSkillSpec(filepath.Join(skillDir, "skill.yaml"))
	if err != nil {
		return "", "", err
	}
	return spec.ID, spec.Version, nil
}

type skillPackagerAdapter struct{}

func (skillPackagerAdapter) Package(skillDir string, outputPath string) error {
	return skill.PackageSkill(skillDir, outputPath)
}

var _ domain.SkillMetadataResolver = skillMetadataResolverAdapter{}
var _ domain.SkillPackager = skillPackagerAdapter{}

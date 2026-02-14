package skillpackage

import (
	"fmt"
	"strings"
)

var ErrSkillDirRequired = fmt.Errorf("skill-dir is required")

type PackageSkillCommand struct {
	SkillDir string
}

type PackageSkillResult struct {
	ArtifactPath string
}

type SkillMetadataResolver interface {
	ResolveIDAndVersion(skillDir string) (id string, version string, err error)
}

type SkillPackager interface {
	Package(skillDir string, outputPath string) error
}

func (c PackageSkillCommand) Normalized() PackageSkillCommand {
	return PackageSkillCommand{SkillDir: strings.TrimSpace(c.SkillDir)}
}

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

// Validate checks that the command has all required fields.
func (c PackageSkillCommand) Validate() error {
	if c.SkillDir == "" {
		return ErrSkillDirRequired
	}
	return nil
}

// ArtifactName returns the conventional artifact filename for a skill package.
func ArtifactName(id, version string) string {
	return id + "-" + version + ".zip"
}

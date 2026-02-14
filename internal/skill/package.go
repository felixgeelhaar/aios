package skill

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func PackageSkill(skillDir, outputZip string) error {
	spec, err := LoadSkillSpec(filepath.Join(skillDir, "skill.yaml"))
	if err != nil {
		return err
	}
	if err := ValidateSkillSpec(skillDir, spec); err != nil {
		return err
	}
	if outputZip == "" {
		outputZip = filepath.Join(filepath.Dir(skillDir), spec.ID+"-"+spec.Version+".zip")
	}

	out, err := os.OpenFile(filepath.Clean(outputZip), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create zip: %w", err)
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()

	err = filepath.WalkDir(skillDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(skillDir, path)
		if err != nil {
			return err
		}
		f, err := zw.Create(rel)
		if err != nil {
			return err
		}
		// #nosec G304 -- path originates from filepath.WalkDir on skillDir.
		src, err := os.Open(filepath.Clean(path))
		if err != nil {
			return err
		}
		defer src.Close()
		_, err = io.Copy(f, src)
		return err
	})
	if err != nil {
		return fmt.Errorf("package skill: %w", err)
	}
	return nil
}

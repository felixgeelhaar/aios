package skill

import (
	"os"
	"testing"
)

// FuzzValidateSkillSpecVersion fuzzes the semver validation in skill specs
func FuzzValidateSkillSpecVersion(f *testing.F) {
	// Valid semver seeds
	f.Add("1.0.0")
	f.Add("0.1.0")
	f.Add("2.3.4-alpha.1")
	f.Add("1.0.0-beta+build.123")
	f.Add("10.20.30")

	f.Fuzz(func(t *testing.T, version string) {
		// Just validate the version field using the regex
		_ = semverRe.MatchString(version)
	})
}

// FuzzLoadSkillSpecYAML fuzzes YAML parsing of skill specs
func FuzzLoadSkillSpecYAML(f *testing.F) {
	// Valid YAML seeds
	f.Add([]byte(`id: test-skill
version: 1.0.0
inputs:
  schema: schema.input.json
outputs:
  schema: schema.output.json
`))
	f.Add([]byte(`id: another-skill
name: Another Skill
version: 0.2.0-alpha
inputs:
  schema: input.json
outputs:
  schema: output.json
`))
	f.Add([]byte(`id: minimal
version: 1.0.0
`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Create temp file with fuzzed content
		tmpDir := t.TempDir()
		path := tmpDir + "/skill.yaml"

		if err := writeFile(path, data); err != nil {
			t.Skip("Failed to write test file")
		}

		spec, err := LoadSkillSpec(path)
		if err != nil {
			// Expected for invalid YAML
			return
		}

		// If we successfully parsed, verify basic structure
		if spec.ID != "" {
			// Valid spec loaded
			_ = spec.Version
			_ = spec.Name
		}
	})
}

func writeFile(path string, data []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

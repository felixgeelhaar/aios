package skill

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

type SkillSpec struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Inputs  struct {
		Schema string `yaml:"schema"`
	} `yaml:"inputs"`
	Outputs struct {
		Schema string `yaml:"schema"`
	} `yaml:"outputs"`
}

func LoadSkillSpec(path string) (SkillSpec, error) {
	var spec SkillSpec
	path = filepath.Clean(path)
	// #nosec G304 -- path is cleaned and provided by trusted CLI/MCP flow.
	data, err := os.ReadFile(path)
	if err != nil {
		return spec, fmt.Errorf("read skill spec: %w", err)
	}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return spec, fmt.Errorf("parse skill spec: %w", err)
	}
	return spec, nil
}

// semverRe matches basic semantic versions: MAJOR.MINOR.PATCH with optional
// pre-release suffix (e.g. "1.0.0-alpha.1"). Build metadata (+build) is also
// accepted per the semver spec.
var semverRe = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)` +
	`(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
	`(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

func ValidateSkillSpec(baseDir string, spec SkillSpec) error {
	if spec.ID == "" {
		return fmt.Errorf("id is required")
	}
	if spec.Version == "" {
		return fmt.Errorf("version is required")
	}
	if !semverRe.MatchString(spec.Version) {
		return fmt.Errorf("version %q is not valid semver", spec.Version)
	}
	if spec.Inputs.Schema == "" || spec.Outputs.Schema == "" {
		return fmt.Errorf("input/output schemas are required")
	}
	if err := ValidateJSONSchema(filepath.Join(baseDir, spec.Inputs.Schema)); err != nil {
		return fmt.Errorf("invalid input schema: %w", err)
	}
	if err := ValidateJSONSchema(filepath.Join(baseDir, spec.Outputs.Schema)); err != nil {
		return fmt.Errorf("invalid output schema: %w", err)
	}
	return nil
}

func ValidateJSONSchema(path string) error {
	path = filepath.Clean(path)
	// #nosec G304 -- path is cleaned and resolved from validated skill spec.
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse schema: %w", err)
	}
	if t, ok := raw["type"].(string); ok && t != "object" {
		return fmt.Errorf("schema type must be object")
	}
	if _, ok := raw["properties"]; !ok {
		return fmt.Errorf("schema missing properties")
	}
	return nil
}

package skill

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type SkillSpec struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Inputs      struct {
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

// BuildSkillMd composes a SKILL.md from a spec and prompt body. The result
// follows the Claude Code SKILL.md frontmatter format with name, description,
// and allowed-tools fields followed by the prompt content.
func BuildSkillMd(spec SkillSpec, promptBody string) string {
	name := spec.Name
	if name == "" {
		name = spec.ID
	}
	desc := spec.Description
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "name: %s\n", name)
	fmt.Fprintf(&b, "description: %s\n", desc)
	b.WriteString("allowed-tools: Read, Grep, Glob\n")
	b.WriteString("---\n")
	if promptBody != "" {
		b.WriteString("\n")
		b.WriteString(strings.TrimRight(promptBody, "\n"))
		b.WriteString("\n")
	}
	return b.String()
}

// LoadAndBuildSkillMd reads skill.yaml and prompt.md from skillDir and returns
// the composed SKILL.md content.
func LoadAndBuildSkillMd(skillDir string) (string, error) {
	spec, err := LoadSkillSpec(filepath.Join(skillDir, "skill.yaml"))
	if err != nil {
		return "", err
	}
	promptPath := filepath.Join(filepath.Clean(skillDir), "prompt.md")
	// #nosec G304 -- path is derived from validated skill directory.
	promptData, err := os.ReadFile(promptPath)
	if err != nil {
		// prompt.md is optional; fall back to spec-only SKILL.md.
		return BuildSkillMd(spec, ""), nil
	}
	return BuildSkillMd(spec, string(promptData)), nil
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

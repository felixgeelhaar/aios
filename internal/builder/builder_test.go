package builder

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/felixgeelhaar/aios/internal/skill"
)

func TestBuildSkill(t *testing.T) {
	dir := t.TempDir()
	if err := BuildSkill(Spec{ID: "roadmap-reader", Version: "0.1.0", Dir: dir}); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "roadmap-reader", "skill.yaml")); err != nil {
		t.Fatalf("skill.yaml missing: %v", err)
	}
}

// AC1: Must scaffold complete directory structure.
func TestBuildSkillScaffoldsCompleteStructure(t *testing.T) {
	dir := t.TempDir()
	if err := BuildSkill(Spec{ID: "test-skill", Version: "1.0.0", Dir: dir}); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	root := filepath.Join(dir, "test-skill")
	for _, path := range []string{
		"skill.yaml",
		"prompt.md",
		"schema.input.json",
		"schema.output.json",
		"tests/fixture_01.json",
		"tests/expected_01.json",
	} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}
	// Verify tests/ is a directory.
	info, err := os.Stat(filepath.Join(root, "tests"))
	if err != nil {
		t.Fatalf("tests dir missing: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("tests should be a directory")
	}
}

// AC2: Generated skill must pass lint-skill without modification.
func TestBuildSkillPassesLint(t *testing.T) {
	dir := t.TempDir()
	if err := BuildSkill(Spec{ID: "lint-pass-skill", Version: "0.2.0", Dir: dir}); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	root := filepath.Join(dir, "lint-pass-skill")
	res, err := skill.LintSkillDir(root)
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}
	if !res.Valid {
		t.Fatalf("scaffolded skill should pass lint, got issues: %v", res.Issues)
	}
}

// AC4: Generated schemas must be valid JSON Schema.
func TestBuildSkillGeneratesValidSchemas(t *testing.T) {
	dir := t.TempDir()
	if err := BuildSkill(Spec{ID: "schema-skill", Version: "0.1.0", Dir: dir}); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	root := filepath.Join(dir, "schema-skill")

	for _, name := range []string{"schema.input.json", "schema.output.json"} {
		if err := skill.ValidateJSONSchema(filepath.Join(root, name)); err != nil {
			t.Fatalf("generated %s is not a valid JSON Schema: %v", name, err)
		}
		// Also verify it parses as valid JSON with type=object and properties.
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("%s is not valid JSON: %v", name, err)
		}
		if raw["type"] != "object" {
			t.Fatalf("%s type should be 'object', got %v", name, raw["type"])
		}
		if _, ok := raw["properties"]; !ok {
			t.Fatalf("%s missing 'properties' key", name)
		}
	}
}

// AC5: Must include at least one starter fixture/expected pair.
func TestBuildSkillIncludesFixturePair(t *testing.T) {
	dir := t.TempDir()
	if err := BuildSkill(Spec{ID: "fixture-skill", Version: "0.1.0", Dir: dir}); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	root := filepath.Join(dir, "fixture-skill")

	fixturePath := filepath.Join(root, "tests", "fixture_01.json")
	expectedPath := filepath.Join(root, "tests", "expected_01.json")

	// Both files must exist.
	for _, p := range []string{fixturePath, expectedPath} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected %s to exist: %v", p, err)
		}
	}

	// Both must be valid JSON.
	for _, p := range []string{fixturePath, expectedPath} {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("%s is not valid JSON: %v", p, err)
		}
		if len(raw) == 0 {
			t.Fatalf("%s should not be an empty object", p)
		}
	}
}

// AC: Builder rejects missing required fields.
func TestBuildSkillRejectsMissingFields(t *testing.T) {
	tests := []struct {
		name string
		spec Spec
	}{
		{"missing id", Spec{Version: "1.0.0", Dir: t.TempDir()}},
		{"missing version", Spec{ID: "x", Dir: t.TempDir()}},
		{"missing dir", Spec{ID: "x", Version: "1.0.0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BuildSkill(tt.spec); err == nil {
				t.Fatal("expected error for missing field")
			}
		})
	}
}

// AC2+AC4: Generated skill.yaml loads and validates successfully.
func TestBuildSkillGeneratesValidSkillYAML(t *testing.T) {
	dir := t.TempDir()
	if err := BuildSkill(Spec{ID: "yaml-skill", Version: "1.2.3", Dir: dir}); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	root := filepath.Join(dir, "yaml-skill")
	spec, err := skill.LoadSkillSpec(filepath.Join(root, "skill.yaml"))
	if err != nil {
		t.Fatalf("load skill spec: %v", err)
	}
	if spec.ID != "yaml-skill" {
		t.Fatalf("expected id 'yaml-skill', got %q", spec.ID)
	}
	if spec.Version != "1.2.3" {
		t.Fatalf("expected version '1.2.3', got %q", spec.Version)
	}
	if spec.Inputs.Schema == "" || spec.Outputs.Schema == "" {
		t.Fatal("skill.yaml must reference input and output schemas")
	}
	if err := skill.ValidateSkillSpec(root, spec); err != nil {
		t.Fatalf("skill spec validation failed: %v", err)
	}
}

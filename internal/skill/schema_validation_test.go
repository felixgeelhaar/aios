package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSkillSpec(t *testing.T) {
	dir := t.TempDir()
	in := `{"type":"object","properties":{"query":{"type":"string"}}}`
	out := `{"type":"object","properties":{"summary":{"type":"string"}}}`
	if err := os.WriteFile(filepath.Join(dir, "schema.input.json"), []byte(in), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.output.json"), []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}
	yaml := "id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	spec, err := LoadSkillSpec(filepath.Join(dir, "skill.yaml"))
	if err != nil {
		t.Fatalf("load spec: %v", err)
	}
	if err := ValidateSkillSpec(dir, spec); err != nil {
		t.Fatalf("validate spec: %v", err)
	}
}

func TestValidateJSONSchemaRejectsNonObject(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "schema.json")
	if err := os.WriteFile(file, []byte(`{"type":"array","properties":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateJSONSchema(file); err == nil {
		t.Fatal("expected error")
	}
}

// AC2: Version must be valid semver — table-driven validation.
func TestSemverValidation(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"1.0.0", true},
		{"0.1.0", true},
		{"0.0.1", true},
		{"10.20.30", true},
		{"1.0.0-alpha", true},
		{"1.0.0-alpha.1", true},
		{"1.0.0-0.3.7", true},
		{"1.0.0-beta.11", true},
		{"1.0.0+build", true},
		{"1.0.0+build.123", true},
		{"1.0.0-alpha+build", true},
		{"1.0.0-rc.1+meta", true},
		// Invalid versions
		{"1.0", false},
		{"v1.0.0", false},
		{"abc", false},
		{"1.0.0.0", false},
		{"", false},
		{"1", false},
		{".1.0", false},
		{"01.0.0", false},
		{"1.02.0", false},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := semverRe.MatchString(tt.version)
			if got != tt.valid {
				t.Fatalf("semverRe.MatchString(%q) = %v, want %v", tt.version, got, tt.valid)
			}
		})
	}
}

// AC: Schema validation must reject malformed JSON before install.
func TestValidateJSONSchemaRejectsMalformedJSON(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{"not json at all", `this is not json`, "parse schema"},
		{"truncated json", `{"type":"object"`, "parse schema"},
		{"empty file", ``, "parse schema"},
		{"json array instead of object", `[1,2,3]`, "parse schema"},
		{"missing properties key", `{"type":"object"}`, "schema missing properties"},
		{"type is not object", `{"type":"array","properties":{}}`, "schema type must be object"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			file := filepath.Join(dir, "schema.json")
			if err := os.WriteFile(file, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			err := ValidateJSONSchema(file)
			if err == nil {
				t.Fatal("expected error for malformed schema")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestBuildSkillMd_WithDescriptionAndPrompt(t *testing.T) {
	spec := SkillSpec{ID: "my-skill", Description: "A helpful skill."}
	got := BuildSkillMd(spec, "# Instructions\n\nDo the thing.\n")
	if !strings.Contains(got, "name: my-skill") {
		t.Error("expected name from ID")
	}
	if !strings.Contains(got, "description: A helpful skill.") {
		t.Error("expected description")
	}
	if !strings.Contains(got, "# Instructions") {
		t.Error("expected prompt body")
	}
}

func TestBuildSkillMd_UsesNameOverID(t *testing.T) {
	spec := SkillSpec{ID: "my-skill", Name: "My Custom Skill", Description: "desc"}
	got := BuildSkillMd(spec, "")
	if !strings.Contains(got, "name: My Custom Skill") {
		t.Errorf("expected Name field used, got:\n%s", got)
	}
}

func TestBuildSkillMd_EmptyPrompt(t *testing.T) {
	spec := SkillSpec{ID: "bare", Description: "minimal"}
	got := BuildSkillMd(spec, "")
	if strings.Contains(got, "\n\n\n") {
		t.Error("expected no extra blank lines when prompt is empty")
	}
	if !strings.HasSuffix(got, "---\n") {
		t.Errorf("expected to end with frontmatter close, got:\n%s", got)
	}
}

func TestLoadAndBuildSkillMd_ReadsFilesFromDisk(t *testing.T) {
	dir := t.TempDir()
	yamlContent := "id: disk-skill\nversion: 1.0.0\ndescription: From disk.\ninputs:\n  schema: in.json\noutputs:\n  schema: out.json\n"
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte(yamlContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "prompt.md"), []byte("# Review\n\nCheck things.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadAndBuildSkillMd(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "description: From disk.") {
		t.Error("expected description from skill.yaml")
	}
	if !strings.Contains(got, "# Review") {
		t.Error("expected prompt.md content")
	}
}

func TestLoadAndBuildSkillMd_NoPromptMd(t *testing.T) {
	dir := t.TempDir()
	yamlContent := "id: no-prompt\nversion: 1.0.0\ndescription: No prompt.\ninputs:\n  schema: in.json\noutputs:\n  schema: out.json\n"
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte(yamlContent), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadAndBuildSkillMd(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "name: no-prompt") {
		t.Error("expected fallback name from ID")
	}
	if !strings.HasSuffix(got, "---\n") {
		t.Errorf("expected to end with frontmatter close when no prompt, got:\n%s", got)
	}
}

// AC1+AC2: ValidateSkillSpec rejects specs missing required fields.
func TestValidateSkillSpecRejectsMissingFields(t *testing.T) {
	// Helper to write valid schema files so we isolate spec-level validation.
	writeSchemas := func(t *testing.T, dir string) {
		t.Helper()
		schema := `{"type":"object","properties":{"k":{"type":"string"}}}`
		if err := os.WriteFile(filepath.Join(dir, "in.json"), []byte(schema), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "out.json"), []byte(schema), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name    string
		spec    SkillSpec
		wantErr string
	}{
		{
			name: "missing id",
			spec: SkillSpec{Version: "1.0.0", Inputs: struct {
				Schema string `yaml:"schema"`
			}{Schema: "in.json"}, Outputs: struct {
				Schema string `yaml:"schema"`
			}{Schema: "out.json"}},
			wantErr: "id is required",
		},
		{
			name: "missing version",
			spec: SkillSpec{ID: "skill-1", Inputs: struct {
				Schema string `yaml:"schema"`
			}{Schema: "in.json"}, Outputs: struct {
				Schema string `yaml:"schema"`
			}{Schema: "out.json"}},
			wantErr: "version is required",
		},
		{
			name: "invalid semver",
			spec: SkillSpec{ID: "skill-1", Version: "1.0", Inputs: struct {
				Schema string `yaml:"schema"`
			}{Schema: "in.json"}, Outputs: struct {
				Schema string `yaml:"schema"`
			}{Schema: "out.json"}},
			wantErr: "not valid semver",
		},
		{
			name: "missing input schema",
			spec: SkillSpec{ID: "skill-1", Version: "1.0.0", Outputs: struct {
				Schema string `yaml:"schema"`
			}{Schema: "out.json"}},
			wantErr: "schemas are required",
		},
		{
			name: "missing output schema",
			spec: SkillSpec{ID: "skill-1", Version: "1.0.0", Inputs: struct {
				Schema string `yaml:"schema"`
			}{Schema: "in.json"}},
			wantErr: "schemas are required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeSchemas(t, dir)
			err := ValidateSkillSpec(dir, tt.spec)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

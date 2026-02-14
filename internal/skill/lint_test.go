package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLintSkillDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "prompt.md"), []byte("# prompt"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := LintSkillDir(dir)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	if !res.Valid {
		t.Fatalf("unexpected issues: %#v", res.Issues)
	}
}

func TestLintSkillDirDetectsFixturePairMismatch(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "prompt.md"), []byte("# prompt"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tests", "fixture_01.json"), []byte(`{"q":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := LintSkillDir(dir)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	if res.Valid {
		t.Fatal("expected lint failure")
	}
}

func TestLintSkillDirRejectsEmbeddedCredentials(t *testing.T) {
	tests := []struct {
		name    string
		file    string // relative path within skill dir
		content string
		pattern string // expected substring in issue
	}{
		{
			name:    "api key in prompt",
			file:    "prompt.md",
			content: "Use this API_KEY= xxxx to authenticate",
			pattern: "embedded credential detected in prompt.md",
		},
		{
			name:    "client secret in skill.yaml",
			file:    "skill.yaml",
			content: "id: bad-skill\nversion: 0.1.0\n# client_secret= xxxx\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n",
			pattern: "embedded credential detected in skill.yaml",
		},
		{
			name:    "bearer token in prompt",
			file:    "prompt.md",
			content: "Authorization: Bearer xxxx",
			pattern: "embedded credential detected in prompt.md",
		},
		{
			name:    "password in schema",
			file:    "schema.input.json",
			content: `{"type":"object","properties":{"q":{"type":"string"}},"password= xxxx":"x"}`,
			pattern: "embedded credential detected in schema.input.json",
		},
		{
			name:    "private key in prompt",
			file:    "prompt.md",
			content: "private_key= xxxx",
			pattern: "embedded credential detected in prompt.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupValidSkillDir(t)

			targetPath := filepath.Join(dir, tt.file)
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(targetPath, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			// For fixtures, add matching expected file.
			if strings.HasPrefix(tt.file, "tests/fixture_") {
				suffix := strings.TrimPrefix(filepath.Base(tt.file), "fixture_")
				expectedPath := filepath.Join(dir, "tests", "expected_"+suffix)
				if err := os.WriteFile(expectedPath, []byte(`{"a":"ok"}`), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			res, err := LintSkillDir(dir)
			if err != nil {
				t.Fatalf("lint: %v", err)
			}
			if res.Valid {
				t.Fatal("expected lint failure for embedded credential")
			}
			found := false
			for _, issue := range res.Issues {
				if strings.Contains(issue, tt.pattern) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected issue containing %q, got: %v", tt.pattern, res.Issues)
			}
		})
	}
}

func TestLintSkillDirCleanSkillPassesCredentialScan(t *testing.T) {
	dir := setupValidSkillDir(t)
	// Write prompt that mentions credentials conceptually but doesn't embed any.
	if err := os.WriteFile(filepath.Join(dir, "prompt.md"), []byte("# Prompt\nThis skill requires Google Drive access.\nThe connector provides authentication at runtime.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := LintSkillDir(dir)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	if !res.Valid {
		t.Fatalf("clean skill should pass lint, got issues: %v", res.Issues)
	}
}

// AC: lint-skill must detect missing prompt.md.
func TestLintSkillDirDetectsMissingPrompt(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte("id: test-skill\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// No prompt.md written.

	res, err := LintSkillDir(dir)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	if res.Valid {
		t.Fatal("expected lint failure for missing prompt.md")
	}
	found := false
	for _, issue := range res.Issues {
		if strings.Contains(issue, "missing prompt.md") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'missing prompt.md' issue, got: %v", res.Issues)
	}
}

// AC: lint-skill must detect missing tests directory.
func TestLintSkillDirDetectsMissingTestsDir(t *testing.T) {
	dir := t.TempDir()
	// No tests/ directory created.
	if err := os.WriteFile(filepath.Join(dir, "prompt.md"), []byte("# prompt"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte("id: test-skill\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := LintSkillDir(dir)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	if res.Valid {
		t.Fatal("expected lint failure for missing tests directory")
	}
	found := false
	for _, issue := range res.Issues {
		if strings.Contains(issue, "missing tests directory") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'missing tests directory' issue, got: %v", res.Issues)
	}
}

// AC: lint-skill must detect missing schema files referenced in skill.yaml.
func TestLintSkillDirDetectsMissingSchemaFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "prompt.md"), []byte("# prompt"), 0o644); err != nil {
		t.Fatal(err)
	}
	// skill.yaml references schema files that don't exist.
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte("id: test-skill\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// schema.input.json and schema.output.json intentionally NOT created.

	res, err := LintSkillDir(dir)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}
	if res.Valid {
		t.Fatal("expected lint failure for missing schema files")
	}
	found := false
	for _, issue := range res.Issues {
		if strings.Contains(issue, "invalid input schema") || strings.Contains(issue, "read schema") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected schema validation issue, got: %v", res.Issues)
	}
}

// setupValidSkillDir creates a minimal valid skill directory for test isolation.
func setupValidSkillDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "prompt.md"), []byte("# prompt"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte("id: test-skill\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

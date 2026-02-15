package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProgressivePromptBaseOnly(t *testing.T) {
	dir := t.TempDir()
	promptPath := filepath.Join(dir, "prompt.md")
	content := `# Base instructions
You are a helpful assistant.
`
	if err := os.WriteFile(promptPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	prompt, err := LoadProgressivePrompt(promptPath)
	if err != nil {
		t.Fatal(err)
	}

	if !contains(prompt.Base(), "helpful assistant") {
		t.Errorf("unexpected base: %q", prompt.Base())
	}
	if len(prompt.Sections) != 0 {
		t.Errorf("expected no sections, got %d", len(prompt.Sections))
	}
}

func TestProgressivePromptWithSections(t *testing.T) {
	dir := t.TempDir()
	promptPath := filepath.Join(dir, "prompt.md")
	content := `# Base instructions
You are a helpful assistant.

# @section code
Write clean, testable code.

# @section *security
Always validate input.
`
	if err := os.WriteFile(promptPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	prompt, err := LoadProgressivePrompt(promptPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(prompt.Sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(prompt.Sections))
	}

	if prompt.Sections[0].Trigger != "code" {
		t.Errorf("expected trigger 'code', got %q", prompt.Sections[0].Trigger)
	}

	if !prompt.Sections[1].Required {
		t.Error("expected second section to be required")
	}

	required := prompt.RequiredSections()
	if len(required) != 1 || required[0] != "security" {
		t.Errorf("expected ['security'], got %v", required)
	}
}

func TestProgressivePromptWithSection(t *testing.T) {
	dir := t.TempDir()
	promptPath := filepath.Join(dir, "prompt.md")
	content := `# Base
Be helpful.

# @section code
Write code.
`
	if err := os.WriteFile(promptPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	prompt, err := LoadProgressivePrompt(promptPath)
	if err != nil {
		t.Fatal(err)
	}

	withCode := prompt.WithSection("code")
	if !contains(withCode, "Be helpful") {
		t.Error("expected base content in withSection")
	}
	if !contains(withCode, "Write code") {
		t.Error("expected section content in withSection")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) &&
			(s[:len(substr)] == substr ||
				contains(s[1:], substr)))
}

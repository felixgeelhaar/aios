package builder

import (
	"fmt"
	"os"
	"path/filepath"
)

type Spec struct {
	ID      string
	Version string
	Dir     string
}

func BuildSkill(s Spec) error {
	if s.ID == "" || s.Version == "" || s.Dir == "" {
		return fmt.Errorf("id, version, and dir are required")
	}
	root := filepath.Join(s.Dir, s.ID)
	if err := os.MkdirAll(filepath.Join(root, "tests"), 0o750); err != nil {
		return err
	}

	skillYAML := "id: " + s.ID + "\nversion: " + s.Version + "\n" +
		"inputs:\n  schema: schema.input.json\n" +
		"outputs:\n  schema: schema.output.json\n"

	inputSchema := `{"type":"object","properties":{"query":{"type":"string","description":"Input query"}}}` + "\n"
	outputSchema := `{"type":"object","properties":{"result":{"type":"string","description":"Output result"}}}` + "\n"

	fixture := `{"query":"example"}` + "\n"
	expected := `{"status":"ok"}` + "\n"

	files := map[string]string{
		"skill.yaml":             skillYAML,
		"prompt.md":              "# Prompt\n",
		"schema.input.json":      inputSchema,
		"schema.output.json":     outputSchema,
		"tests/fixture_01.json":  fixture,
		"tests/expected_01.json": expected,
	}
	for rel, content := range files {
		if err := os.WriteFile(filepath.Join(root, rel), []byte(content), 0o600); err != nil {
			return err
		}
	}
	return nil
}

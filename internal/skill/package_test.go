package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackageSkill(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "roadmap-reader")
	if err := os.MkdirAll(filepath.Join(skillDir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	outZip := filepath.Join(root, "skill.zip")
	if err := PackageSkill(skillDir, outZip); err != nil {
		t.Fatalf("package failed: %v", err)
	}
	if _, err := os.Stat(outZip); err != nil {
		t.Fatalf("missing zip: %v", err)
	}
}

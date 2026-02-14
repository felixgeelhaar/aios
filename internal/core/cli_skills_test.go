package core

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCLISyncRequiresSkillDir(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	if err := cli.Run(context.Background(), "sync", "", "stdio", ":8080", "text"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCLISyncPlanRequiresSkillDir(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	if err := cli.Run(context.Background(), "sync-plan", "", "stdio", ":8080", "text"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCLIInitSkill(t *testing.T) {
	root := t.TempDir()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	dir := filepath.Join(root, "roadmap-reader")
	if err := cli.Run(context.Background(), "init-skill", dir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("init-skill failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "skill.yaml")); err != nil {
		t.Fatalf("missing skill scaffold: %v", err)
	}
}

func TestCLISyncValidatesAndInstalls(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_PROJECT_DIR", root)
	cfg := DefaultConfig()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, cfg)

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
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

	if err := cli.Run(context.Background(), "sync", skillDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("sync failed: %v", err)
	}
}

func TestCLITestSkillPasses(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultConfig()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, cfg)

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"query":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"status":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "tests", "fixture_01.json"), []byte(`{"query":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "tests", "expected_01.json"), []byte(`{"status":"ok"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := cli.Run(context.Background(), "test-skill", skillDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("test-skill failed: %v", err)
	}
}

func TestCLILintSkill(t *testing.T) {
	root := t.TempDir()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "prompt.md"), []byte("# p"), 0o644); err != nil {
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
	if err := cli.Run(context.Background(), "lint-skill", skillDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("lint-skill failed: %v", err)
	}
}

func TestCLIPackageSkill(t *testing.T) {
	root := t.TempDir()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	skillDir := filepath.Join(root, "skill")
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
	if err := cli.Run(context.Background(), "package-skill", skillDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("package-skill failed: %v", err)
	}
}

func TestCLIUninstallSkill(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_PROJECT_DIR", root)
	cfg := DefaultConfig()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, cfg)

	skillDir := filepath.Join(root, "skill")
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

	if err := cli.Run(context.Background(), "sync", skillDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	if err := cli.Run(context.Background(), "uninstall-skill", skillDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cfg.ProjectDir, ".agents", "skills", "roadmap-reader")); !os.IsNotExist(err) {
		t.Fatalf("expected canonical skill removed, stat err: %v", err)
	}
}

func TestCLISyncJSON(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_PROJECT_DIR", root)
	cfg := DefaultConfig()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, cfg)

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
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

	if err := cli.Run(context.Background(), "sync", skillDir, "stdio", ":8080", "json"); err != nil {
		t.Fatalf("sync json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["synced"] != true {
		t.Fatalf("expected synced=true, got: %#v", out)
	}
	if out["skill_id"] != "roadmap-reader" {
		t.Fatalf("unexpected skill_id: %#v", out)
	}
}

func TestCLITestSkillJSON(t *testing.T) {
	root := t.TempDir()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"query":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"status":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "tests", "fixture_01.json"), []byte(`{"query":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "tests", "expected_01.json"), []byte(`{"status":"ok"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := cli.Run(context.Background(), "test-skill", skillDir, "stdio", ":8080", "json"); err != nil {
		t.Fatalf("test-skill json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["results"] == nil {
		t.Fatalf("missing results in json output: %#v", out)
	}
}

func TestCLIInitSkillJSON(t *testing.T) {
	root := t.TempDir()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	dir := filepath.Join(root, "roadmap-reader")
	if err := cli.Run(context.Background(), "init-skill", dir, "stdio", ":8080", "json"); err != nil {
		t.Fatalf("init-skill json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["initialized"] != true {
		t.Fatalf("expected initialized=true, got: %#v", out)
	}
	if out["skill_dir"] != dir {
		t.Fatalf("unexpected skill_dir: %#v", out)
	}
}

func TestCLILintSkillJSON(t *testing.T) {
	root := t.TempDir()
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "prompt.md"), []byte("# p"), 0o644); err != nil {
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
	if err := cli.Run(context.Background(), "lint-skill", skillDir, "stdio", ":8080", "json"); err != nil {
		t.Fatalf("lint-skill json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["valid"] != true {
		t.Fatalf("expected valid=true, got: %#v", out)
	}
}

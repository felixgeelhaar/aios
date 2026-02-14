package core

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCLIDefaultMarketplaceAndMatrixFlow(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AIOS_WORKSPACE_DIR", root)
	t.Setenv("AIOS_PROJECT_DIR", projectDir)

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

	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	if err := cli.Run(context.Background(), "marketplace-publish", skillDir, "stdio", ":8080", "json"); err != nil {
		t.Fatalf("marketplace-publish failed: %v", err)
	}
	buf.Reset()
	if err := cli.Run(context.Background(), "marketplace-list", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("marketplace-list failed: %v", err)
	}
	buf.Reset()
	if err := cli.Run(context.Background(), "marketplace-install", "roadmap-reader", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("marketplace-install failed: %v", err)
	}
	buf.Reset()
	if err := cli.Run(context.Background(), "marketplace-matrix", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("marketplace-matrix failed: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if _, ok := out["matrix"]; !ok {
		t.Fatalf("missing matrix in output: %#v", out)
	}
}

func TestCLIDefaultAuditExportAndVerify(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AIOS_WORKSPACE_DIR", root)
	t.Setenv("AIOS_PROJECT_DIR", projectDir)

	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	if err := cli.Run(context.Background(), "audit-export", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("audit-export failed: %v", err)
	}
	var exportOut map[string]any
	if err := json.Unmarshal(buf.Bytes(), &exportOut); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	path, _ := exportOut["path"].(string)
	if path == "" {
		t.Fatalf("missing path in export output: %#v", exportOut)
	}

	buf.Reset()
	if err := cli.Run(context.Background(), "audit-verify", path, "stdio", ":8080", "json"); err != nil {
		t.Fatalf("audit-verify failed: %v", err)
	}
	var verifyOut map[string]any
	if err := json.Unmarshal(buf.Bytes(), &verifyOut); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if verifyOut["valid"] != true {
		t.Fatalf("expected valid=true, got: %#v", verifyOut)
	}
}

func TestCLIDefaultRuntimeExecutionReport(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AIOS_WORKSPACE_DIR", root)
	t.Setenv("AIOS_PROJECT_DIR", projectDir)

	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	if err := cli.Run(context.Background(), "runtime-execution-report", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("runtime-execution-report failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	path, _ := out["path"].(string)
	if path == "" {
		t.Fatalf("missing report path: %#v", out)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("missing report file: %v", err)
	}
}

func TestCLIDefaultTrayStatus(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "project")
	skillDir := filepath.Join(projectDir, ".agents", "skills", "roadmap-reader")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: roadmap-reader\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("AIOS_WORKSPACE_DIR", root)
	t.Setenv("AIOS_PROJECT_DIR", projectDir)

	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	if err := cli.Run(context.Background(), "tray-status", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("tray-status failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected tray-status output")
	}
}

package core

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIDefaultProjectInventoryTextFlow(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "project")
	t.Setenv("AIOS_WORKSPACE_DIR", root)
	t.Setenv("AIOS_PROJECT_DIR", projectDir)

	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	if err := cli.Run(context.Background(), "project-add", projectDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("project-add failed: %v", err)
	}
	if !strings.Contains(buf.String(), "project tracked") {
		t.Fatalf("unexpected project-add output: %q", buf.String())
	}
	buf.Reset()

	if err := cli.Run(context.Background(), "project-list", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("project-list failed: %v", err)
	}
	if !strings.Contains(buf.String(), projectDir) {
		t.Fatalf("expected project path in list output: %q", buf.String())
	}
	buf.Reset()

	if err := cli.Run(context.Background(), "project-inspect", projectDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("project-inspect failed: %v", err)
	}
	if !strings.Contains(buf.String(), "added_at") {
		t.Fatalf("unexpected project-inspect output: %q", buf.String())
	}
}

func TestCLIDefaultAnalyticsSummaryText(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "project")
	t.Setenv("AIOS_WORKSPACE_DIR", root)
	t.Setenv("AIOS_PROJECT_DIR", projectDir)

	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	if err := cli.Run(context.Background(), "project-add", projectDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("project-add failed: %v", err)
	}
	buf.Reset()

	if err := cli.Run(context.Background(), "analytics-summary", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("analytics-summary failed: %v", err)
	}
	if !strings.Contains(buf.String(), "tracked_projects") {
		t.Fatalf("unexpected analytics summary output: %q", buf.String())
	}
}

func TestCLIDefaultStatusText(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_WORKSPACE_DIR", root)

	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	if err := cli.Run(context.Background(), "status", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(buf.String(), "status:") {
		t.Fatalf("unexpected status output: %q", buf.String())
	}
}

func TestCLIDefaultListClientsText(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "project")
	t.Setenv("AIOS_WORKSPACE_DIR", root)
	t.Setenv("AIOS_PROJECT_DIR", projectDir)

	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())

	if err := cli.Run(context.Background(), "list-clients", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("list-clients failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected list-clients output")
	}
}

package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	domainprojectinventory "github.com/felixgeelhaar/aios/internal/domain/projectinventory"
	domainworkspace "github.com/felixgeelhaar/aios/internal/domain/workspaceorchestration"
	"github.com/felixgeelhaar/aios/internal/runtime"
	mcpg "github.com/felixgeelhaar/mcp-go"
)

func TestRunCLIRejectsUnknown(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	if err := cli.Run(context.Background(), "invalid", "", "stdio", ":8080", "text"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCLIStatusOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.Health = func() runtime.HealthReport {
		return runtime.HealthReport{Status: "ok", Ready: true, TokenStore: "keychain", Workspace: "/tmp/aios"}
	}
	if err := cli.Run(context.Background(), "status", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("status failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "status: ok") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestCLIServeMCPInvokesServe(t *testing.T) {
	buf := &bytes.Buffer{}
	called := false
	cli := DefaultCLI(buf, DefaultConfig())
	cli.ServeMCP = func(context.Context, *mcpg.Server, ...mcpg.ServeOption) error {
		called = true
		return errors.New("stop")
	}
	_ = cli.Run(context.Background(), "serve-mcp", "", "stdio", ":8080", "text")
	if !called {
		t.Fatal("expected ServeMCP to be called")
	}
}

func TestCLIHelp(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	if err := cli.Run(context.Background(), "help", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("help failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "commands:") {
		t.Fatalf("unexpected help output: %q", buf.String())
	}
	for _, cmd := range []string{
		"project-list",
		"project-add",
		"model-policy-packs",
		"analytics-summary",
		"analytics-record",
		"analytics-trend",
		"marketplace-publish",
		"marketplace-list",
		"marketplace-install",
		"marketplace-matrix",
		"audit-export",
		"audit-verify",
		"runtime-execution-report",
		"workspace-validate",
		"workspace-plan",
		"workspace-repair",
		"tui",
	} {
		if !strings.Contains(out, cmd) {
			t.Fatalf("help output missing %q: %q", cmd, out)
		}
	}
}

func TestCLIVersion(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.BuildInfo = func() BuildInfo {
		return BuildInfo{Version: "1.2.3", Commit: "abc123", BuildDate: "2026-02-13"}
	}
	if err := cli.Run(context.Background(), "version", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("version failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "version: 1.2.3") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestCLIStatusJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.Health = func() runtime.HealthReport {
		return runtime.HealthReport{Status: "ok", Ready: true, TokenStore: "keychain", Workspace: "/tmp/aios"}
	}
	if err := cli.Run(context.Background(), "status", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("status json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["status"] != "ok" {
		t.Fatalf("unexpected status: %#v", out)
	}
}

// AC1: Health check JSON must report all required fields: status, ready, sync, token_store, workspace.
func TestCLIStatusJSONContainsAllFields(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.Health = func() runtime.HealthReport {
		return runtime.HealthReport{Status: "ok", Ready: true, TokenStore: "keychain", Workspace: "/tmp/aios"}
	}
	cli.SyncState = func() string { return "clean" }
	if err := cli.Run(context.Background(), "status", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("status json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	for _, key := range []string{"status", "ready", "sync", "token_store", "workspace"} {
		if _, ok := out[key]; !ok {
			t.Fatalf("missing required field %q in status JSON", key)
		}
	}
}

func TestCLIDoctorJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.Doctor = func() DoctorReport {
		return DoctorReport{Overall: true, Checks: []DoctorCheck{{Name: "workspace_dir", OK: true, Detail: "/tmp"}}}
	}
	if err := cli.Run(context.Background(), "doctor", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("doctor json failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["overall"] != true {
		t.Fatalf("unexpected doctor result: %#v", out)
	}
}

// AC7: Doctor exit code must reflect health status (non-zero on failure).
func TestCLIDoctorFailReturnsError(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.Doctor = func() DoctorReport {
		return DoctorReport{Overall: false, Checks: []DoctorCheck{{Name: "workspace_dir", OK: false, Detail: "missing"}}}
	}
	err := cli.Run(context.Background(), "doctor", "", "stdio", ":8080", "text")
	if err == nil {
		t.Fatal("expected error for failed doctor checks")
	}
}

// AC7: Doctor JSON output on failure still returns JSON but Run returns error.
func TestCLIDoctorFailJSONReturnsError(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.Doctor = func() DoctorReport {
		return DoctorReport{Overall: false, Checks: []DoctorCheck{{Name: "workspace_dir", OK: false, Detail: "missing"}}}
	}
	err := cli.Run(context.Background(), "doctor", "", "stdio", ":8080", "json")
	if err == nil {
		t.Fatal("expected error for failed doctor checks in json mode")
	}
	// Should still have written JSON before returning error.
	var out map[string]any
	if jsonErr := json.Unmarshal(buf.Bytes(), &out); jsonErr != nil {
		t.Fatalf("expected json output even on failure: %v", jsonErr)
	}
	if out["overall"] != false {
		t.Fatalf("expected overall=false in json output, got: %#v", out)
	}
}

func TestCLIListClientsJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	if err := cli.Run(context.Background(), "list-clients", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("list-clients failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if _, ok := out["opencode"]; !ok {
		t.Fatalf("missing opencode entry: %#v", out)
	}
}

func TestCLIModelPolicyPacksJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	if err := cli.Run(context.Background(), "model-policy-packs", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("model-policy-packs failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	packs, ok := out["policy_packs"].([]any)
	if !ok || len(packs) == 0 {
		t.Fatalf("missing policy packs: %#v", out)
	}
}

func TestCLITUIQuit(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.In = strings.NewReader("q\n")
	cli.ListProjects = func(context.Context) ([]domainprojectinventory.Project, error) { return nil, nil }
	cli.ValidateWorkspace = func(context.Context) (domainworkspace.ValidationResult, error) {
		return domainworkspace.ValidationResult{Healthy: true}, nil
	}
	cli.RepairWorkspace = func(context.Context) (domainworkspace.RepairResult, error) {
		return domainworkspace.RepairResult{}, nil
	}

	if err := cli.Run(context.Background(), "tui", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("tui failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "AIOS Operations Console") {
		t.Fatalf("unexpected tui output: %q", out)
	}
}

func TestCLITUIProjectsAndValidate(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.In = strings.NewReader("q\n")
	cli.ListProjects = func(context.Context) ([]domainprojectinventory.Project, error) {
		return []domainprojectinventory.Project{
			{ID: "p1", Path: "/tmp/repo", AddedAt: "2026-02-13T00:00:00Z"},
		}, nil
	}
	cli.ValidateWorkspace = func(context.Context) (domainworkspace.ValidationResult, error) {
		return domainworkspace.ValidationResult{
			Healthy: true,
		}, nil
	}
	cli.RepairWorkspace = func(context.Context) (domainworkspace.RepairResult, error) {
		return domainworkspace.RepairResult{}, nil
	}

	if err := cli.Run(context.Background(), "tui", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("tui failed: %v", err)
	}
}

// AC6: tray-status surfaces connection state (google_drive: true/false).
func TestCLITrayStatusSurfacesConnectionState(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.TrayStatus = func() (TrayState, error) {
		return TrayState{
			UpdatedAt:   "2026-02-13T00:00:00Z",
			Skills:      []string{},
			Connections: map[string]bool{"google_drive": true},
		}, nil
	}

	if err := cli.Run(context.Background(), "tray-status", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("tray-status failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	conns, ok := out["connections"].(map[string]any)
	if !ok {
		t.Fatalf("expected connections map, got: %#v", out)
	}
	if conns["google_drive"] != true {
		t.Fatalf("expected google_drive=true in tray-status, got: %#v", conns)
	}
}

// AC6: tray-status text output shows connection status.
func TestCLITrayStatusTextShowsConnection(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.TrayStatus = func() (TrayState, error) {
		return TrayState{
			UpdatedAt:   "2026-02-13T00:00:00Z",
			Skills:      []string{},
			Connections: map[string]bool{"google_drive": false},
		}, nil
	}

	if err := cli.Run(context.Background(), "tray-status", "", "stdio", ":8080", "text"); err != nil {
		t.Fatalf("tray-status failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "google_drive") {
		t.Fatalf("tray-status text must show google_drive connection, got: %q", out)
	}
}

func TestCLITrayStatusJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, DefaultConfig())
	cli.TrayStatus = func() (TrayState, error) {
		return TrayState{
			UpdatedAt: "2026-02-13T00:00:00Z",
			Skills:    []string{"roadmap-reader"},
			Connections: map[string]bool{
				"google_drive": true,
			},
		}, nil
	}

	if err := cli.Run(context.Background(), "tray-status", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("tray-status failed: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["updated_at"] == "" {
		t.Fatalf("missing updated_at: %#v", out)
	}
}

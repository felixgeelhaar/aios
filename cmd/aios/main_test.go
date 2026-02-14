package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunReturnsExitCode2OnUnknownFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"status", "--unknown"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown flag") {
		t.Fatalf("expected flag parse error, got %q", stderr.String())
	}
}

func TestRunReturnsExitCode1OnUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"no-such-command"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("expected unknown command error, got %q", stderr.String())
	}
}

func TestRunReturnsExitCode0ForStatus(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"status"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%q", code, stderr.String())
	}
}

func TestRunLegacyModeSupportsCLICommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--mode", "cli", "--command", "status"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%q", code, stderr.String())
	}
}

func TestRunReturnsExitCode1OnUnsupportedMode(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--mode", "invalid"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unsupported mode") {
		t.Fatalf("expected unsupported mode error, got %q", stderr.String())
	}
}

func TestRunEmitsStructuredJSONError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"status", "--output", "json", "--unknown"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}

	var errResp map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &errResp); err != nil {
		t.Fatalf("expected structured JSON error on stderr, got %q: %v", stderr.String(), err)
	}
	if _, ok := errResp["error"]; !ok {
		t.Fatalf("JSON error missing 'error' field: %#v", errResp)
	}
	if errResp["command"] != "status" {
		t.Fatalf("JSON error missing correct 'command' field: %#v", errResp)
	}
}

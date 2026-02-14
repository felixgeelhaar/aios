package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunReturnsExitCode2OnFlagParseError(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"-unknown-flag"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("expected flag parse error in stderr, got %q", stderr.String())
	}
}

func TestRunReturnsExitCode1OnUnsupportedMode(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"-mode", "invalid"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unsupported mode") {
		t.Fatalf("expected unsupported mode error in stderr, got %q", stderr.String())
	}
}

func TestRunReturnsExitCode1OnUnknownCLICommand(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"-mode", "cli", "-command", "no-such-command"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown cli command") {
		t.Fatalf("expected unknown command error in stderr, got %q", stderr.String())
	}
}

func TestRunReturnsExitCode0ForCLIStatus(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"-mode", "cli", "-command", "status"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%q", code, stderr.String())
	}
}

func TestRunUsesSkillIDFallback(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"-mode", "cli", "-command", "sync", "-skill-id", "testdata/missing"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "read skill spec") {
		t.Fatalf("expected sync error using skill-id fallback, got %q", stderr.String())
	}
}

func TestRunEmitsStructuredJSONError(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"-mode", "cli", "-command", "no-such-command", "-output", "json"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}

	var errResp map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &errResp); err != nil {
		t.Fatalf("expected structured JSON error on stderr, got %q: %v", stderr.String(), err)
	}
	if _, ok := errResp["error"]; !ok {
		t.Fatalf("JSON error missing 'error' field: %#v", errResp)
	}
	if errResp["command"] != "no-such-command" {
		t.Fatalf("JSON error missing correct 'command' field: %#v", errResp)
	}
}

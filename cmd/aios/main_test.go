package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestMainCallsRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"status"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%q", code, stderr.String())
	}
}

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

func TestRunTrayMode(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--mode", "tray", "--command", "status"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%q", code, stderr.String())
	}
}

func TestRunLegacyWithSkillID(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--mode", "cli", "--command", "status", "--skill-id", "my-skill-id"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%q", code, stderr.String())
	}
}

func TestExtractOutputFlagLongForm(t *testing.T) {
	result := extractOutputFlag([]string{"--output", "json"})
	if result != "json" {
		t.Fatalf("expected 'json', got %q", result)
	}
}

func TestExtractOutputFlagShortForm(t *testing.T) {
	result := extractOutputFlag([]string{"--output=json"})
	if result != "json" {
		t.Fatalf("expected 'json', got %q", result)
	}
}

func TestExtractOutputFlagDefault(t *testing.T) {
	result := extractOutputFlag([]string{"status"})
	if result != "text" {
		t.Fatalf("expected 'text', got %q", result)
	}
}

func TestIsFlagParseErrorUnknownFlag(t *testing.T) {
	err := &testError{msg: "unknown flag: --unknown"}
	if !isFlagParseError(err) {
		t.Fatal("expected true for unknown flag error")
	}
}

func TestIsFlagParseErrorRequiresArg(t *testing.T) {
	err := &testError{msg: "--output requires an argument"}
	if !isFlagParseError(err) {
		t.Fatal("expected true for requires argument error")
	}
}

func TestIsFlagParseErrorNil(t *testing.T) {
	if isFlagParseError(nil) {
		t.Fatal("expected false for nil error")
	}
}

func TestIsFlagParseErrorOther(t *testing.T) {
	err := &testError{msg: "some other error"}
	if isFlagParseError(err) {
		t.Fatal("expected false for other error")
	}
}

func TestSkillsInitRequiresArg(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"skills", "init"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "skill-dir is required") {
		t.Fatalf("expected skill-dir error, got %q", stderr.String())
	}
}

func TestSkillsSyncRequiresArg(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"skills", "sync"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "skill-dir is required") {
		t.Fatalf("expected skill-dir error, got %q", stderr.String())
	}
}

func TestProjectAddRequiresArg(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"project", "add"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "path is required") {
		t.Fatalf("expected path error, got %q", stderr.String())
	}
}

func TestVersionCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "version:") {
		t.Fatalf("expected version output, got %q", stdout.String())
	}
}

func TestDoctorCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"doctor"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "doctor:") {
		t.Fatalf("expected doctor output, got %q", stdout.String())
	}
}

func TestListClientsCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"list-clients"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestModelPolicyPacksCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"model-policy-packs"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestWorkspaceValidateCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"workspace", "validate"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestBackupConfigsCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"backup-configs"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestTrayStatusCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"tray-status"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestProjectListCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"project", "list"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestMarketplaceListCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"marketplace", "list"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestAnalyticsSummaryCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"analytics", "summary"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestAuditExportCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"audit", "export"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestRuntimeExecutionReportCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"runtime", "execution-report"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestRestoreConfigsCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"restore-configs"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestExportStatusReportCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"export-status-report"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestSkillsPlanCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"skills", "plan"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1 for missing arg, got %d", code)
	}
}

func TestSkillsTestCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"skills", "test"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1 for missing arg, got %d", code)
	}
}

func TestSkillsLintCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"skills", "lint"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1 for missing arg, got %d", code)
	}
}

func TestSkillsPackageCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"skills", "package"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1 for missing arg, got %d", code)
	}
}

func TestSkillsUninstallCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"skills", "uninstall"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1 for missing arg, got %d", code)
	}
}

func TestMarketplacePublishCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"marketplace", "publish"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1 for missing arg, got %d", code)
	}
}

func TestMarketplaceInstallCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"marketplace", "install"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1 for missing arg, got %d", code)
	}
}

func TestMCPServeCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"mcp", "serve", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestProjectRemoveCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"project", "remove"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1 for missing arg, got %d", code)
	}
}

func TestProjectInspectCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"project", "inspect"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1 for missing arg, got %d", code)
	}
}

func TestWorkspacePlanCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"workspace", "plan"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestWorkspaceRepairCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"workspace", "repair"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestAnalyticsRecordCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"analytics", "record"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestAnalyticsTrendCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"analytics", "trend"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestAuditVerifyCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"audit", "verify"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1 for missing arg, got %d", code)
	}
}

func TestRootHelpCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestRootNoArgsShowsHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExportStatusReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.md")
	err := ExportStatusReport(path, BuildInfo{Version: "0.1.0", Commit: "dev", BuildDate: "today"}, DoctorReport{Overall: true}, map[string]any{"status": "ok"})
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("missing report: %v", err)
	}
}

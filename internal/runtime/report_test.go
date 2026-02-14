package runtime

import (
	"path/filepath"
	"testing"
)

func TestBuildAndWriteExecutionReport(t *testing.T) {
	report := BuildExecutionReport(ExecutionPlan{
		SkillID: "roadmap-reader",
		Version: "0.1.0",
		Model:   "gpt-4.1",
	}, "ok")
	if report.SkillID != "roadmap-reader" {
		t.Fatalf("unexpected report: %#v", report)
	}
	path := filepath.Join(t.TempDir(), "state", "execution-report.json")
	if err := WriteExecutionReport(path, report); err != nil {
		t.Fatalf("write report failed: %v", err)
	}
}

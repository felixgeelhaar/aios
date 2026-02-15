package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/felixgeelhaar/aios/internal/runtime"
)

func TestExecutionReportStore_WriteReport(t *testing.T) {
	store := fileExecutionReportStore{}
	report := runtime.BuildExecutionReport(runtime.ExecutionPlan{
		SkillID: "roadmap-reader",
		Version: "0.1.0",
		Model:   "gpt-4.1",
	}, "ok")
	path := filepath.Join(t.TempDir(), "state", "execution-report.json")
	if err := store.WriteReport(path, report); err != nil {
		t.Fatalf("write report failed: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var parsed runtime.ExecutionReport
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("report is not valid JSON: %v", err)
	}
	if parsed.SkillID != "roadmap-reader" {
		t.Fatalf("unexpected skill_id: %s", parsed.SkillID)
	}
	if parsed.ExecutionOutcome != "ok" {
		t.Fatalf("unexpected outcome: %s", parsed.ExecutionOutcome)
	}
}

package runtime

import (
	"testing"
)

func TestBuildExecutionReport(t *testing.T) {
	report := BuildExecutionReport(ExecutionPlan{
		SkillID: "roadmap-reader",
		Version: "0.1.0",
		Model:   "gpt-4.1",
	}, "ok")
	if report.SkillID != "roadmap-reader" {
		t.Fatalf("unexpected skill_id: %s", report.SkillID)
	}
	if report.Version != "0.1.0" {
		t.Fatalf("unexpected version: %s", report.Version)
	}
	if report.ExecutionOutcome != "ok" {
		t.Fatalf("unexpected outcome: %s", report.ExecutionOutcome)
	}
	if report.GeneratedAt == "" {
		t.Fatal("expected non-empty GeneratedAt")
	}
}

package runtime

import (
	"time"
)

type ExecutionReport struct {
	GeneratedAt      string `json:"generated_at"`
	SkillID          string `json:"skill_id"`
	Version          string `json:"version"`
	Model            string `json:"model"`
	PolicyTelemetry  any    `json:"policy_telemetry"`
	ExecutionOutcome string `json:"execution_outcome"`
}

// ExecutionReportStore abstracts persistence of runtime execution reports.
type ExecutionReportStore interface {
	WriteReport(path string, report ExecutionReport) error
}

func BuildExecutionReport(plan ExecutionPlan, outcome string) ExecutionReport {
	return ExecutionReport{
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
		SkillID:          plan.SkillID,
		Version:          plan.Version,
		Model:            plan.Model,
		PolicyTelemetry:  plan.PolicyTelemetry,
		ExecutionOutcome: outcome,
	}
}

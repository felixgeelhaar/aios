package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func WriteExecutionReport(path string, report ExecutionReport) error {
	body, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o600)
}

package workspaceorchestration

import "context"

type LinkStatus string

const (
	LinkStatusOK       LinkStatus = "ok"
	LinkStatusMissing  LinkStatus = "missing"
	LinkStatusBroken   LinkStatus = "broken"
	LinkStatusConflict LinkStatus = "conflict"
)

type ActionKind string

const (
	ActionCreate ActionKind = "create"
	ActionRepair ActionKind = "repair"
	ActionSkip   ActionKind = "skip"
)

type ProjectRef struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type LinkReport struct {
	ProjectID     string     `json:"project_id"`
	ProjectPath   string     `json:"project_path"`
	LinkPath      string     `json:"link_path"`
	Status        LinkStatus `json:"status"`
	CurrentTarget string     `json:"current_target,omitempty"`
}

type ValidationResult struct {
	Healthy bool         `json:"healthy"`
	Links   []LinkReport `json:"links"`
}

type PlanAction struct {
	Kind       ActionKind `json:"kind"`
	ProjectID  string     `json:"project_id"`
	LinkPath   string     `json:"link_path"`
	TargetPath string     `json:"target_path"`
	Reason     string     `json:"reason"`
}

type PlanResult struct {
	Actions []PlanAction `json:"actions"`
}

type RepairResult struct {
	Applied []PlanAction `json:"applied"`
	Skipped []PlanAction `json:"skipped"`
}

type ProjectSource interface {
	ListProjects(ctx context.Context) ([]ProjectRef, error)
}

type WorkspaceLinks interface {
	Inspect(projectID string, targetPath string) (LinkReport, error)
	Ensure(projectID string, targetPath string) error
}

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

// RecommendAction maps a link's validation status to the appropriate plan action.
func (lr LinkReport) RecommendAction() PlanAction {
	action := PlanAction{
		ProjectID:  lr.ProjectID,
		LinkPath:   lr.LinkPath,
		TargetPath: lr.ProjectPath,
	}
	switch lr.Status {
	case LinkStatusOK:
		action.Kind = ActionSkip
		action.Reason = "already healthy"
	case LinkStatusMissing:
		action.Kind = ActionCreate
		action.Reason = "link missing"
	case LinkStatusBroken:
		action.Kind = ActionRepair
		action.Reason = "link target mismatch"
	case LinkStatusConflict:
		action.Kind = ActionSkip
		action.Reason = "non-symlink conflict at link path"
	}
	return action
}

// IsApplicable returns true for actions that modify state (create, repair).
func (pa PlanAction) IsApplicable() bool {
	return pa.Kind == ActionCreate || pa.Kind == ActionRepair
}

// ComputeHealthy returns true when every link has status OK.
func ComputeHealthy(links []LinkReport) bool {
	for _, l := range links {
		if l.Status != LinkStatusOK {
			return false
		}
	}
	return true
}

type ProjectSource interface {
	ListProjects(ctx context.Context) ([]ProjectRef, error)
}

type WorkspaceLinks interface {
	Inspect(projectID string, targetPath string) (LinkReport, error)
	Ensure(projectID string, targetPath string) error
}

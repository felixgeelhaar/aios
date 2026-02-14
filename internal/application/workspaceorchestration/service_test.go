package workspaceorchestration

import (
	"context"
	"testing"

	domain "github.com/felixgeelhaar/aios/internal/domain/workspaceorchestration"
)

type fakeProjectSource struct {
	projects []domain.ProjectRef
}

func (f fakeProjectSource) ListProjects(context.Context) ([]domain.ProjectRef, error) {
	return f.projects, nil
}

type fakeWorkspaceLinks struct {
	inspect map[string]domain.LinkReport
	ensured []string
}

func (f *fakeWorkspaceLinks) Inspect(projectID string, _ string) (domain.LinkReport, error) {
	return f.inspect[projectID], nil
}

func (f *fakeWorkspaceLinks) Ensure(projectID string, _ string) error {
	f.ensured = append(f.ensured, projectID)
	return nil
}

func TestPlanAndRepair(t *testing.T) {
	source := fakeProjectSource{
		projects: []domain.ProjectRef{
			{ID: "p1", Path: "/repo1"},
			{ID: "p2", Path: "/repo2"},
			{ID: "p3", Path: "/repo3"},
		},
	}
	links := &fakeWorkspaceLinks{
		inspect: map[string]domain.LinkReport{
			"p1": {ProjectID: "p1", ProjectPath: "/repo1", LinkPath: "/links/p1", Status: domain.LinkStatusOK},
			"p2": {ProjectID: "p2", ProjectPath: "/repo2", LinkPath: "/links/p2", Status: domain.LinkStatusMissing},
			"p3": {ProjectID: "p3", ProjectPath: "/repo3", LinkPath: "/links/p3", Status: domain.LinkStatusBroken},
		},
	}
	svc := NewService(source, links)

	plan, err := svc.Plan(context.Background())
	if err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	if len(plan.Actions) != 3 {
		t.Fatalf("unexpected action count: %d", len(plan.Actions))
	}
	if plan.Actions[1].Kind != domain.ActionCreate {
		t.Fatalf("unexpected action kind: %s", plan.Actions[1].Kind)
	}
	if plan.Actions[2].Kind != domain.ActionRepair {
		t.Fatalf("unexpected action kind: %s", plan.Actions[2].Kind)
	}

	repair, err := svc.Repair(context.Background())
	if err != nil {
		t.Fatalf("repair failed: %v", err)
	}
	if len(repair.Applied) != 2 {
		t.Fatalf("expected 2 applied actions, got %d", len(repair.Applied))
	}
}

func TestPlanConflictIsSkippedAndNotApplied(t *testing.T) {
	source := fakeProjectSource{
		projects: []domain.ProjectRef{
			{ID: "p1", Path: "/repo1"},
		},
	}
	links := &fakeWorkspaceLinks{
		inspect: map[string]domain.LinkReport{
			"p1": {ProjectID: "p1", ProjectPath: "/repo1", LinkPath: "/links/p1", Status: domain.LinkStatusConflict},
		},
	}
	svc := NewService(source, links)

	validation, err := svc.Validate(context.Background())
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if validation.Healthy {
		t.Fatalf("expected unhealthy validation for conflict: %#v", validation)
	}

	plan, err := svc.Plan(context.Background())
	if err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	if len(plan.Actions) != 1 {
		t.Fatalf("expected one action, got %d", len(plan.Actions))
	}
	if plan.Actions[0].Kind != domain.ActionSkip {
		t.Fatalf("expected skip action, got %s", plan.Actions[0].Kind)
	}

	repair, err := svc.Repair(context.Background())
	if err != nil {
		t.Fatalf("repair failed: %v", err)
	}
	if len(repair.Applied) != 0 {
		t.Fatalf("expected no applied actions, got %d", len(repair.Applied))
	}
	if len(repair.Skipped) != 1 {
		t.Fatalf("expected one skipped action, got %d", len(repair.Skipped))
	}
}

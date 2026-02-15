package workspaceorchestration

import (
	"context"

	domain "github.com/felixgeelhaar/aios/internal/domain/workspaceorchestration"
)

type Service struct {
	source domain.ProjectSource
	links  domain.WorkspaceLinks
}

func NewService(source domain.ProjectSource, links domain.WorkspaceLinks) Service {
	return Service{
		source: source,
		links:  links,
	}
}

func (s Service) Validate(ctx context.Context) (domain.ValidationResult, error) {
	projects, err := s.source.ListProjects(ctx)
	if err != nil {
		return domain.ValidationResult{}, err
	}
	links := make([]domain.LinkReport, 0, len(projects))
	for _, p := range projects {
		link, err := s.links.Inspect(p.ID, p.Path)
		if err != nil {
			return domain.ValidationResult{}, err
		}
		links = append(links, link)
	}
	return domain.ValidationResult{
		Healthy: domain.ComputeHealthy(links),
		Links:   links,
	}, nil
}

func (s Service) Plan(ctx context.Context) (domain.PlanResult, error) {
	validation, err := s.Validate(ctx)
	if err != nil {
		return domain.PlanResult{}, err
	}
	actions := make([]domain.PlanAction, 0, len(validation.Links))
	for _, link := range validation.Links {
		actions = append(actions, link.RecommendAction())
	}
	return domain.PlanResult{Actions: actions}, nil
}

func (s Service) Repair(ctx context.Context) (domain.RepairResult, error) {
	plan, err := s.Plan(ctx)
	if err != nil {
		return domain.RepairResult{}, err
	}
	applied := make([]domain.PlanAction, 0, len(plan.Actions))
	skipped := make([]domain.PlanAction, 0, len(plan.Actions))
	for _, action := range plan.Actions {
		if !action.IsApplicable() {
			skipped = append(skipped, action)
			continue
		}
		if err := s.links.Ensure(action.ProjectID, action.TargetPath); err != nil {
			action.Reason = action.Reason + ": " + err.Error()
			skipped = append(skipped, action)
			continue
		}
		applied = append(applied, action)
	}
	return domain.RepairResult{
		Applied: applied,
		Skipped: skipped,
	}, nil
}

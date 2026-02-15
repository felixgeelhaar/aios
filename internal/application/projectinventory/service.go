package projectinventory

import (
	"context"
	"strings"
	"time"

	domain "github.com/felixgeelhaar/aios/internal/domain/projectinventory"
)

type Service struct {
	repo          domain.Repository
	canonicalizer domain.PathCanonicalizer
	now           func() time.Time
}

func NewService(repo domain.Repository, canonicalizer domain.PathCanonicalizer) Service {
	return Service{
		repo:          repo,
		canonicalizer: canonicalizer,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s Service) List(ctx context.Context) ([]domain.Project, error) {
	inv, err := s.repo.Load(ctx)
	if err != nil {
		return nil, err
	}
	return inv.SortedProjects(), nil
}

func (s Service) Track(ctx context.Context, path string) (domain.Project, error) {
	if strings.TrimSpace(path) == "" {
		return domain.Project{}, domain.ErrProjectPathRequired
	}
	canonicalPath, err := s.canonicalizer.Canonicalize(path)
	if err != nil {
		return domain.Project{}, err
	}

	inv, err := s.repo.Load(ctx)
	if err != nil {
		return domain.Project{}, err
	}

	project := domain.Project{
		ID:      domain.ProjectID(canonicalPath),
		Path:    canonicalPath,
		AddedAt: s.now().Format(time.RFC3339),
	}
	if !inv.Track(project) {
		if existing, ok := inv.FindBySelector(project.ID); ok {
			return existing, nil
		}
		return domain.Project{}, nil
	}
	if err := s.repo.Save(ctx, inv); err != nil {
		return domain.Project{}, err
	}
	return project, nil
}

func (s Service) Untrack(ctx context.Context, selector string) error {
	key := domain.NormalizeSelector(selector)
	if key == "" {
		return domain.ErrProjectSelectorRequired
	}

	inv, err := s.repo.Load(ctx)
	if err != nil {
		return err
	}
	if !inv.Untrack(key) {
		if canonicalPath, err := s.canonicalizer.Canonicalize(key); err == nil {
			if !inv.Untrack(canonicalPath) {
				return domain.ErrProjectNotFound
			}
		} else {
			return domain.ErrProjectNotFound
		}
	}
	return s.repo.Save(ctx, inv)
}

func (s Service) Inspect(ctx context.Context, selector string) (domain.Project, error) {
	key := domain.NormalizeSelector(selector)
	if key == "" {
		return domain.Project{}, domain.ErrProjectSelectorRequired
	}

	inv, err := s.repo.Load(ctx)
	if err != nil {
		return domain.Project{}, err
	}
	if p, ok := inv.FindBySelector(key); ok {
		return p, nil
	}
	if canonicalPath, err := s.canonicalizer.Canonicalize(key); err == nil {
		if p, ok := inv.FindBySelector(canonicalPath); ok {
			return p, nil
		}
	}
	return domain.Project{}, domain.ErrProjectNotFound
}

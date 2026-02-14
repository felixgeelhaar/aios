package projectinventory

import (
	"context"
	"sort"
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
	out := append([]domain.Project(nil), inv.Projects...)
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
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

	id := domain.ProjectID(canonicalPath)
	for _, p := range inv.Projects {
		if p.ID == id || p.Path == canonicalPath {
			return p, nil
		}
	}

	project := domain.Project{
		ID:      id,
		Path:    canonicalPath,
		AddedAt: s.now().Format(time.RFC3339),
	}
	inv.Projects = append(inv.Projects, project)
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
	filtered := make([]domain.Project, 0, len(inv.Projects))
	removed := false
	for _, p := range inv.Projects {
		if p.ID == key || p.Path == key {
			removed = true
			continue
		}
		filtered = append(filtered, p)
	}
	if !removed {
		if canonicalPath, err := s.canonicalizer.Canonicalize(key); err == nil {
			filtered = filtered[:0]
			for _, p := range inv.Projects {
				if p.ID == key || p.Path == canonicalPath {
					removed = true
					continue
				}
				filtered = append(filtered, p)
			}
		}
	}
	if !removed {
		return domain.ErrProjectNotFound
	}
	inv.Projects = filtered
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
	for _, p := range inv.Projects {
		if p.ID == key || p.Path == key {
			return p, nil
		}
	}
	if canonicalPath, err := s.canonicalizer.Canonicalize(key); err == nil {
		for _, p := range inv.Projects {
			if p.Path == canonicalPath {
				return p, nil
			}
		}
	}
	return domain.Project{}, domain.ErrProjectNotFound
}

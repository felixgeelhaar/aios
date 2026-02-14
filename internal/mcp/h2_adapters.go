package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	domainproject "github.com/felixgeelhaar/aios/internal/domain/projectinventory"
	domainworkspace "github.com/felixgeelhaar/aios/internal/domain/workspaceorchestration"
)

type mcpProjectInventoryRepository struct {
	workspaceDir string
}

type mcpProjectInventoryFile struct {
	Version   int                     `json:"version"`
	UpdatedAt string                  `json:"updated_at"`
	Projects  []domainproject.Project `json:"projects"`
}

func mcpProjectInventoryPath(workspaceDir string) string {
	return filepath.Join(workspaceDir, "projects", "inventory.json")
}

func (r mcpProjectInventoryRepository) Load(context.Context) (domainproject.Inventory, error) {
	path := filepath.Clean(mcpProjectInventoryPath(r.workspaceDir))
	// #nosec G304 -- path is constrained to workspace inventory location.
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return domainproject.Inventory{Projects: []domainproject.Project{}}, nil
		}
		return domainproject.Inventory{}, err
	}
	var inv mcpProjectInventoryFile
	if err := json.Unmarshal(body, &inv); err != nil {
		return domainproject.Inventory{}, err
	}
	return domainproject.Inventory{Projects: inv.Projects}, nil
}

func (r mcpProjectInventoryRepository) Save(_ context.Context, inventory domainproject.Inventory) error {
	path := mcpProjectInventoryPath(r.workspaceDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	projects := append([]domainproject.Project(nil), inventory.Projects...)
	sort.Slice(projects, func(i, j int) bool { return projects[i].Path < projects[j].Path })
	body, err := json.MarshalIndent(mcpProjectInventoryFile{
		Version:   1,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Projects:  projects,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o600)
}

type mcpPathCanonicalizer struct{}

func (mcpPathCanonicalizer) Canonicalize(path string) (string, error) {
	return filepath.Abs(filepath.Clean(path))
}

type mcpInventoryProjectSource struct {
	repo mcpProjectInventoryRepository
}

func (s mcpInventoryProjectSource) ListProjects(ctx context.Context) ([]domainworkspace.ProjectRef, error) {
	inv, err := s.repo.Load(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domainworkspace.ProjectRef, 0, len(inv.Projects))
	for _, p := range inv.Projects {
		out = append(out, domainworkspace.ProjectRef{ID: p.ID, Path: p.Path})
	}
	return out, nil
}

type mcpFilesystemWorkspaceLinks struct {
	workspaceDir string
}

func mcpWorkspaceLinksDir(workspaceDir string) string {
	return filepath.Join(workspaceDir, "projects", "links")
}

func mcpWorkspaceLinkPath(workspaceDir, projectID string) string {
	return filepath.Join(mcpWorkspaceLinksDir(workspaceDir), projectID)
}

func (f mcpFilesystemWorkspaceLinks) Inspect(projectID string, targetPath string) (domainworkspace.LinkReport, error) {
	linkPath := mcpWorkspaceLinkPath(f.workspaceDir, projectID)
	report := domainworkspace.LinkReport{
		ProjectID:   projectID,
		ProjectPath: targetPath,
		LinkPath:    linkPath,
	}
	info, err := os.Lstat(linkPath)
	if err != nil {
		if os.IsNotExist(err) {
			report.Status = domainworkspace.LinkStatusMissing
			return report, nil
		}
		return domainworkspace.LinkReport{}, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		report.Status = domainworkspace.LinkStatusConflict
		return report, nil
	}
	current, err := os.Readlink(linkPath)
	if err != nil {
		return domainworkspace.LinkReport{}, err
	}
	report.CurrentTarget = current
	if filepath.Clean(current) == filepath.Clean(targetPath) {
		report.Status = domainworkspace.LinkStatusOK
		return report, nil
	}
	report.Status = domainworkspace.LinkStatusBroken
	return report, nil
}

func (f mcpFilesystemWorkspaceLinks) Ensure(projectID string, targetPath string) error {
	if err := os.MkdirAll(mcpWorkspaceLinksDir(f.workspaceDir), 0o750); err != nil {
		return err
	}
	linkPath := mcpWorkspaceLinkPath(f.workspaceDir, projectID)
	info, err := os.Lstat(linkPath)
	if err == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("link path %s exists and is not a symlink", linkPath)
		}
		if err := os.Remove(linkPath); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.Symlink(targetPath, linkPath)
}

func mcpWorkspaceDir() string {
	if v := strings.TrimSpace(os.Getenv("AIOS_WORKSPACE_DIR")); v != "" {
		return v
	}
	return filepath.Join(".", ".aios")
}

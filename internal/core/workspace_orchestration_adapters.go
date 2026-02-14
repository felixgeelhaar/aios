package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	domainworkspace "github.com/felixgeelhaar/aios/internal/domain/workspaceorchestration"
)

type inventoryProjectSource struct {
	repo fileProjectInventoryRepository
}

func (s inventoryProjectSource) ListProjects(ctx context.Context) ([]domainworkspace.ProjectRef, error) {
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

type filesystemWorkspaceLinks struct {
	workspaceDir string
}

func workspaceLinksDir(workspaceDir string) string {
	return filepath.Join(workspaceDir, "projects", "links")
}

func workspaceLinkPath(workspaceDir string, projectID string) string {
	return filepath.Join(workspaceLinksDir(workspaceDir), projectID)
}

func (f filesystemWorkspaceLinks) Inspect(projectID string, targetPath string) (domainworkspace.LinkReport, error) {
	linkPath := workspaceLinkPath(f.workspaceDir, projectID)
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

func (f filesystemWorkspaceLinks) Ensure(projectID string, targetPath string) error {
	if err := os.MkdirAll(workspaceLinksDir(f.workspaceDir), 0o750); err != nil {
		return err
	}
	linkPath := workspaceLinkPath(f.workspaceDir, projectID)
	info, err := os.Lstat(linkPath)
	if err == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("link path %s exists and is not a symlink", linkPath)
		}
		if removeErr := os.Remove(linkPath); removeErr != nil {
			return removeErr
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.Symlink(targetPath, linkPath)
}

var _ domainworkspace.ProjectSource = inventoryProjectSource{}
var _ domainworkspace.WorkspaceLinks = filesystemWorkspaceLinks{}

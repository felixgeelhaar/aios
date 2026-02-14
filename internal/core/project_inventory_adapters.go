package core

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	domain "github.com/felixgeelhaar/aios/internal/domain/projectinventory"
)

type fileProjectInventoryRepository struct {
	workspaceDir string
}

type projectInventoryFile struct {
	Version   int              `json:"version"`
	UpdatedAt string           `json:"updated_at"`
	Projects  []domain.Project `json:"projects"`
}

func inventoryFilePath(workspaceDir string) string {
	return filepath.Join(workspaceDir, "projects", "inventory.json")
}

func (r fileProjectInventoryRepository) Load(context.Context) (domain.Inventory, error) {
	path := filepath.Clean(inventoryFilePath(r.workspaceDir))
	// #nosec G304 -- path is constrained to workspace inventory location.
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.Inventory{Projects: []domain.Project{}}, nil
		}
		return domain.Inventory{}, err
	}
	var file projectInventoryFile
	if err := json.Unmarshal(body, &file); err != nil {
		return domain.Inventory{}, err
	}
	return domain.Inventory{Projects: file.Projects}, nil
}

func (r fileProjectInventoryRepository) Save(_ context.Context, inventory domain.Inventory) error {
	path := inventoryFilePath(r.workspaceDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	projects := append([]domain.Project(nil), inventory.Projects...)
	sort.Slice(projects, func(i, j int) bool { return projects[i].Path < projects[j].Path })
	file := projectInventoryFile{
		Version:   1,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Projects:  projects,
	}
	body, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o600)
}

type absPathCanonicalizer struct{}

func (absPathCanonicalizer) Canonicalize(path string) (string, error) {
	return filepath.Abs(filepath.Clean(path))
}

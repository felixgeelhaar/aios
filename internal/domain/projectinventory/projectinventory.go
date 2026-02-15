package projectinventory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

var ErrProjectPathRequired = fmt.Errorf("project path is required")
var ErrProjectSelectorRequired = fmt.Errorf("project selector is required")
var ErrProjectNotFound = fmt.Errorf("project not found")

type Project struct {
	ID      string `json:"id"`
	Path    string `json:"path"`
	AddedAt string `json:"added_at"`
}

type Inventory struct {
	Projects []Project `json:"projects"`
}

type Repository interface {
	Load(ctx context.Context) (Inventory, error)
	Save(ctx context.Context, inventory Inventory) error
}

type PathCanonicalizer interface {
	Canonicalize(path string) (string, error)
}

func NormalizeSelector(selector string) string {
	return strings.TrimSpace(selector)
}

func ProjectID(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:])
}

// FindBySelector looks up a project by ID or path.
func (inv Inventory) FindBySelector(selector string) (Project, bool) {
	for _, p := range inv.Projects {
		if p.ID == selector || p.Path == selector {
			return p, true
		}
	}
	return Project{}, false
}

// Track appends a project if no duplicate (by ID or path) exists.
// Returns true if the project was added, false if already tracked.
func (inv *Inventory) Track(project Project) bool {
	for _, p := range inv.Projects {
		if p.ID == project.ID || p.Path == project.Path {
			return false
		}
	}
	inv.Projects = append(inv.Projects, project)
	return true
}

// Untrack removes a project by ID or path.
// Returns true if a project was removed, false if not found.
func (inv *Inventory) Untrack(selector string) bool {
	for i, p := range inv.Projects {
		if p.ID == selector || p.Path == selector {
			inv.Projects = append(inv.Projects[:i], inv.Projects[i+1:]...)
			return true
		}
	}
	return false
}

// SortedProjects returns a copy of the projects sorted by path.
func (inv Inventory) SortedProjects() []Project {
	out := append([]Project(nil), inv.Projects...)
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

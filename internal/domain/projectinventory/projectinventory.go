package projectinventory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

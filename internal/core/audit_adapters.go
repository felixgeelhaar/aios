package core

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/felixgeelhaar/aios/internal/governance"
)

// fileAuditBundleStore implements governance.AuditBundleStore using the local
// filesystem for reading and writing signed audit bundles.
type fileAuditBundleStore struct{}

var _ governance.AuditBundleStore = fileAuditBundleStore{}

func (fileAuditBundleStore) WriteBundle(path string, bundle governance.AuditBundle) error {
	body, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o600)
}

func (fileAuditBundleStore) LoadBundle(path string) (governance.AuditBundle, error) {
	path = filepath.Clean(path)
	// #nosec G304 -- path is provided by explicit audit export/verify command input.
	body, err := os.ReadFile(path)
	if err != nil {
		return governance.AuditBundle{}, err
	}
	var bundle governance.AuditBundle
	if err := json.Unmarshal(body, &bundle); err != nil {
		return governance.AuditBundle{}, err
	}
	return bundle, nil
}

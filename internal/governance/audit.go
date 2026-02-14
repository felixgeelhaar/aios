package governance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type AuditRecord struct {
	Category  string         `json:"category"`
	Decision  string         `json:"decision"`
	Actor     string         `json:"actor"`
	Timestamp string         `json:"timestamp"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type AuditBundle struct {
	GeneratedAt string        `json:"generated_at"`
	Signature   string        `json:"signature"`
	Records     []AuditRecord `json:"records"`
}

func BuildBundle(records []AuditRecord) (AuditBundle, error) {
	payload, err := json.Marshal(records)
	if err != nil {
		return AuditBundle{}, err
	}
	sum := sha256.Sum256(payload)
	return AuditBundle{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Signature:   hex.EncodeToString(sum[:]),
		Records:     records,
	}, nil
}

func WriteBundle(path string, bundle AuditBundle) error {
	body, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o600)
}

func LoadBundle(path string) (AuditBundle, error) {
	path = filepath.Clean(path)
	// #nosec G304 -- path is provided by explicit audit export/verify command input.
	body, err := os.ReadFile(path)
	if err != nil {
		return AuditBundle{}, err
	}
	var bundle AuditBundle
	if err := json.Unmarshal(body, &bundle); err != nil {
		return AuditBundle{}, err
	}
	return bundle, nil
}

func VerifyBundle(bundle AuditBundle) error {
	payload, err := json.Marshal(bundle.Records)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(payload)
	expected := hex.EncodeToString(sum[:])
	if bundle.Signature != expected {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

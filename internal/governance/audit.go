package governance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

// AuditBundleStore abstracts export and import of signed audit bundles.
type AuditBundleStore interface {
	WriteBundle(path string, bundle AuditBundle) error
	LoadBundle(path string) (AuditBundle, error)
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

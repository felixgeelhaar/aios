package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/felixgeelhaar/aios/internal/governance"
)

func TestAuditBundleStore_RoundTrip(t *testing.T) {
	store := fileAuditBundleStore{}
	bundle, err := governance.BuildBundle([]governance.AuditRecord{
		{Category: "policy", Decision: "allow", Actor: "system", Timestamp: "2026-02-13T00:00:00Z"},
	})
	if err != nil {
		t.Fatalf("build bundle failed: %v", err)
	}
	path := filepath.Join(t.TempDir(), "audit", "bundle.json")
	if err := store.WriteBundle(path, bundle); err != nil {
		t.Fatalf("write bundle failed: %v", err)
	}
	loaded, err := store.LoadBundle(path)
	if err != nil {
		t.Fatalf("load bundle failed: %v", err)
	}
	if err := governance.VerifyBundle(loaded); err != nil {
		t.Fatalf("verify bundle failed: %v", err)
	}
	if len(loaded.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(loaded.Records))
	}
}

// AC5: Exported audit bundle is valid JSON and includes signature.
func TestAuditBundleStore_ExportedBundleIsValidJSON(t *testing.T) {
	store := fileAuditBundleStore{}
	records := []governance.AuditRecord{
		{Category: "publish", Decision: "allow", Actor: "admin", Timestamp: "2026-02-13T00:00:00Z"},
	}
	bundle, err := governance.BuildBundle(records)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	path := filepath.Join(t.TempDir(), "export.json")
	if err := store.WriteBundle(path, bundle); err != nil {
		t.Fatalf("write: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var parsed governance.AuditBundle
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("exported bundle is not valid JSON: %v", err)
	}
	if parsed.Signature == "" {
		t.Fatal("exported bundle missing signature")
	}
	if parsed.GeneratedAt == "" {
		t.Fatal("exported bundle missing generated_at timestamp")
	}
}

// AC5/AC6: Round-trip export -> load -> verify with multiple records.
func TestAuditBundleStore_MultiRecordRoundTrip(t *testing.T) {
	store := fileAuditBundleStore{}
	records := []governance.AuditRecord{
		{Category: "publish", Decision: "allow", Actor: "admin", Timestamp: "2026-02-13T10:00:00Z"},
		{Category: "install", Decision: "allow", Actor: "user", Timestamp: "2026-02-13T10:05:00Z"},
	}
	bundle, err := governance.BuildBundle(records)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	path := filepath.Join(t.TempDir(), "audit", "roundtrip.json")
	if err := store.WriteBundle(path, bundle); err != nil {
		t.Fatalf("write: %v", err)
	}
	loaded, err := store.LoadBundle(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := governance.VerifyBundle(loaded); err != nil {
		t.Fatalf("round-trip verify failed: %v", err)
	}
	if len(loaded.Records) != 2 {
		t.Fatalf("expected 2 records after round-trip, got %d", len(loaded.Records))
	}
}

// AC4: Metadata is preserved through bundle lifecycle.
func TestAuditBundleStore_MetadataPreserved(t *testing.T) {
	store := fileAuditBundleStore{}
	records := []governance.AuditRecord{
		{
			Category:  "publish",
			Decision:  "allow",
			Actor:     "admin",
			Timestamp: "2026-02-13T00:00:00Z",
			Metadata:  map[string]any{"skill_id": "test-skill", "version": "1.0.0", "scopes": "drive"},
		},
	}
	bundle, err := governance.BuildBundle(records)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	path := filepath.Join(t.TempDir(), "meta.json")
	if err := store.WriteBundle(path, bundle); err != nil {
		t.Fatalf("write: %v", err)
	}
	loaded, err := store.LoadBundle(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	meta := loaded.Records[0].Metadata
	if meta["skill_id"] != "test-skill" {
		t.Fatalf("metadata skill_id not preserved: %v", meta)
	}
	if meta["version"] != "1.0.0" {
		t.Fatalf("metadata version not preserved: %v", meta)
	}
}

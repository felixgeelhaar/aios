package governance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildAndWriteBundle(t *testing.T) {
	bundle, err := BuildBundle([]AuditRecord{
		{Category: "policy", Decision: "allow", Actor: "system", Timestamp: "2026-02-13T00:00:00Z"},
	})
	if err != nil {
		t.Fatalf("build bundle failed: %v", err)
	}
	if bundle.Signature == "" {
		t.Fatal("expected signature")
	}
	path := filepath.Join(t.TempDir(), "audit", "bundle.json")
	if err := WriteBundle(path, bundle); err != nil {
		t.Fatalf("write bundle failed: %v", err)
	}
	loaded, err := LoadBundle(path)
	if err != nil {
		t.Fatalf("load bundle failed: %v", err)
	}
	if err := VerifyBundle(loaded); err != nil {
		t.Fatalf("verify bundle failed: %v", err)
	}
}

func TestVerifyBundleDetectsTamper(t *testing.T) {
	bundle, err := BuildBundle([]AuditRecord{
		{Category: "policy", Decision: "allow", Actor: "system", Timestamp: "2026-02-13T00:00:00Z"},
	})
	if err != nil {
		t.Fatalf("build bundle failed: %v", err)
	}
	bundle.Records[0].Decision = "deny"
	if err := VerifyBundle(bundle); err == nil {
		t.Fatal("expected signature verification failure")
	}
}

// AC4: Audit records capture publish, install, and policy decisions.
func TestAuditRecordCategories(t *testing.T) {
	records := []AuditRecord{
		{Category: "publish", Decision: "allow", Actor: "admin-user", Timestamp: "2026-02-13T00:00:00Z", Metadata: map[string]any{"skill_id": "analytics-skill", "version": "1.0.0"}},
		{Category: "install", Decision: "allow", Actor: "team-lead", Timestamp: "2026-02-13T00:01:00Z", Metadata: map[string]any{"skill_id": "analytics-skill"}},
		{Category: "policy", Decision: "deny", Actor: "system", Timestamp: "2026-02-13T00:02:00Z", Metadata: map[string]any{"reason": "scope violation"}},
	}
	bundle, err := BuildBundle(records)
	if err != nil {
		t.Fatalf("build bundle: %v", err)
	}
	if len(bundle.Records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(bundle.Records))
	}
	// Verify all decision categories are preserved.
	categories := map[string]bool{}
	for _, r := range bundle.Records {
		categories[r.Category] = true
	}
	for _, c := range []string{"publish", "install", "policy"} {
		if !categories[c] {
			t.Fatalf("missing audit category %q", c)
		}
	}
}

// AC5: Exported audit bundle is valid JSON and includes signature.
func TestExportedBundleIsValidJSON(t *testing.T) {
	records := []AuditRecord{
		{Category: "publish", Decision: "allow", Actor: "admin", Timestamp: "2026-02-13T00:00:00Z"},
	}
	bundle, err := BuildBundle(records)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	path := filepath.Join(t.TempDir(), "export.json")
	if err := WriteBundle(path, bundle); err != nil {
		t.Fatalf("write: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var parsed AuditBundle
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

// AC5/AC6: Round-trip export → load → verify.
func TestBundleRoundTrip(t *testing.T) {
	records := []AuditRecord{
		{Category: "publish", Decision: "allow", Actor: "admin", Timestamp: "2026-02-13T10:00:00Z"},
		{Category: "install", Decision: "allow", Actor: "user", Timestamp: "2026-02-13T10:05:00Z"},
	}
	bundle, err := BuildBundle(records)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	path := filepath.Join(t.TempDir(), "audit", "roundtrip.json")
	if err := WriteBundle(path, bundle); err != nil {
		t.Fatalf("write: %v", err)
	}
	loaded, err := LoadBundle(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := VerifyBundle(loaded); err != nil {
		t.Fatalf("round-trip verify failed: %v", err)
	}
	if len(loaded.Records) != 2 {
		t.Fatalf("expected 2 records after round-trip, got %d", len(loaded.Records))
	}
}

// AC5: Bundle with empty records still produces valid signed bundle.
func TestBuildBundleEmptyRecords(t *testing.T) {
	bundle, err := BuildBundle([]AuditRecord{})
	if err != nil {
		t.Fatalf("build with empty records: %v", err)
	}
	if bundle.Signature == "" {
		t.Fatal("expected signature even for empty records")
	}
	if err := VerifyBundle(bundle); err != nil {
		t.Fatalf("empty bundle should verify: %v", err)
	}
}

// AC4: Metadata is preserved through bundle lifecycle.
func TestAuditRecordMetadataPreserved(t *testing.T) {
	records := []AuditRecord{
		{
			Category:  "publish",
			Decision:  "allow",
			Actor:     "admin",
			Timestamp: "2026-02-13T00:00:00Z",
			Metadata:  map[string]any{"skill_id": "test-skill", "version": "1.0.0", "scopes": "drive"},
		},
	}
	bundle, err := BuildBundle(records)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	path := filepath.Join(t.TempDir(), "meta.json")
	if err := WriteBundle(path, bundle); err != nil {
		t.Fatalf("write: %v", err)
	}
	loaded, err := LoadBundle(path)
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

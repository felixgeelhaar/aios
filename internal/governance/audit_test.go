package governance

import (
	"testing"
)

func TestBuildBundle(t *testing.T) {
	bundle, err := BuildBundle([]AuditRecord{
		{Category: "policy", Decision: "allow", Actor: "system", Timestamp: "2026-02-13T00:00:00Z"},
	})
	if err != nil {
		t.Fatalf("build bundle failed: %v", err)
	}
	if bundle.Signature == "" {
		t.Fatal("expected signature")
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

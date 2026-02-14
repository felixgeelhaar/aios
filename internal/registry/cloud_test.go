package registry

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPublishAndListVersions(t *testing.T) {
	r := NewCloudRegistry()
	if err := r.Publish(SkillVersion{
		ID:                "roadmap-reader",
		Version:           "0.1.0",
		CompatibleClients: []string{"opencode"},
	}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	if len(r.Versions("roadmap-reader")) != 1 {
		t.Fatal("expected 1 version")
	}
}

func TestRegistryPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "registry", "cloud.json")
	r, err := NewCloudRegistryWithPath(path)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	if err := r.Publish(SkillVersion{
		ID:                "roadmap-reader",
		Version:           "0.1.0",
		CompatibleClients: []string{"opencode", "cursor"},
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	loaded, err := NewCloudRegistryWithPath(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(loaded.Versions("roadmap-reader")) != 1 {
		t.Fatal("expected persisted version")
	}
}

func TestPublishRejectsMissingCompatibility(t *testing.T) {
	r := NewCloudRegistry()
	err := r.Publish(SkillVersion{ID: "roadmap-reader", Version: "0.1.0"})
	if err == nil {
		t.Fatal("expected compatibility validation error")
	}
}

func TestPublishRejectsBadgeWithoutEvidence(t *testing.T) {
	r := NewCloudRegistry()
	err := r.Publish(SkillVersion{
		ID:                "roadmap-reader",
		Version:           "0.1.0",
		CompatibleClients: []string{"opencode"},
		BadgeRequested:    true,
	})
	if err == nil {
		t.Fatal("expected badge evidence validation error")
	}
}

// AC1: Must support skill publishing with version metadata.
func TestPublishStoresVersionMetadata(t *testing.T) {
	r := NewCloudRegistry()
	sv := SkillVersion{
		ID:                "analytics-skill",
		Version:           "2.3.1",
		CompatibleClients: []string{"opencode", "cursor", "windsurf"},
	}
	if err := r.Publish(sv); err != nil {
		t.Fatalf("publish: %v", err)
	}

	versions := r.Versions("analytics-skill")
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	if versions[0] != "2.3.1" {
		t.Fatalf("expected version '2.3.1', got %q", versions[0])
	}

	// Verify it appears in full listing.
	listing := r.List()
	if _, ok := listing["analytics-skill"]; !ok {
		t.Fatal("skill missing from List()")
	}
}

// AC1: Publish rejects empty ID or version.
func TestPublishRejectsEmptyIDOrVersion(t *testing.T) {
	tests := []struct {
		name string
		sv   SkillVersion
		want string
	}{
		{"empty id", SkillVersion{Version: "1.0.0", CompatibleClients: []string{"opencode"}}, "id and version are required"},
		{"empty version", SkillVersion{ID: "x", CompatibleClients: []string{"opencode"}}, "id and version are required"},
	}
	r := NewCloudRegistry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.Publish(tt.sv)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %q", tt.want, err.Error())
			}
		})
	}
}

// AC2: Publish enforces compatibility contract (client allowlist acts as signing gate).
func TestPublishRejectsUnsupportedClient(t *testing.T) {
	r := NewCloudRegistry()
	err := r.Publish(SkillVersion{
		ID:                "bad-skill",
		Version:           "1.0.0",
		CompatibleClients: []string{"unknown_client"},
	})
	if err == nil {
		t.Fatal("expected error for unsupported client")
	}
	if !strings.Contains(err.Error(), "unsupported compatible client") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AC6: Multiple versions can be published per skill (rollback data exists).
func TestPublishMultipleVersions(t *testing.T) {
	r := NewCloudRegistry()
	for _, v := range []string{"1.0.0", "1.1.0", "1.2.0"} {
		if err := r.Publish(SkillVersion{
			ID:                "multi-version-skill",
			Version:           v,
			CompatibleClients: []string{"opencode"},
		}); err != nil {
			t.Fatalf("publish %s: %v", v, err)
		}
	}
	versions := r.Versions("multi-version-skill")
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}
	// Versions should be in publish order (rollback would use earlier entry).
	if versions[0] != "1.0.0" || versions[1] != "1.1.0" || versions[2] != "1.2.0" {
		t.Fatalf("version order unexpected: %v", versions)
	}
}

// AC8: Registry state persists across instances (multiple skills).
func TestRegistryPersistsMultipleSkills(t *testing.T) {
	path := filepath.Join(t.TempDir(), "registry", "cloud.json")
	r, err := NewCloudRegistryWithPath(path)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	for _, id := range []string{"skill-a", "skill-b", "skill-c"} {
		if err := r.Publish(SkillVersion{
			ID:                id,
			Version:           "0.1.0",
			CompatibleClients: []string{"opencode"},
		}); err != nil {
			t.Fatalf("publish %s: %v", id, err)
		}
	}

	loaded, err := NewCloudRegistryWithPath(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	listing := loaded.List()
	if len(listing) != 3 {
		t.Fatalf("expected 3 skills after reload, got %d", len(listing))
	}
	for _, id := range []string{"skill-a", "skill-b", "skill-c"} {
		if _, ok := listing[id]; !ok {
			t.Fatalf("missing skill %q after reload", id)
		}
	}
}

// AC8: List returns defensive copy — mutations do not affect registry.
func TestListReturnsDefensiveCopy(t *testing.T) {
	r := NewCloudRegistry()
	if err := r.Publish(SkillVersion{
		ID:                "safe-skill",
		Version:           "1.0.0",
		CompatibleClients: []string{"opencode"},
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	listing := r.List()
	// Mutate the returned copy.
	listing["safe-skill"] = append(listing["safe-skill"], "hacked")
	listing["injected"] = []string{"bad"}

	// Original registry must be unaffected.
	if len(r.Versions("safe-skill")) != 1 {
		t.Fatal("defensive copy failed: original mutated")
	}
	if len(r.Versions("injected")) != 0 {
		t.Fatal("defensive copy failed: injected key appeared in registry")
	}
}

// AC1 (Marketplace Ecosystem): Private org registries via separate persistent paths.
func TestPrivateOrgRegistriesIsolated(t *testing.T) {
	dir := t.TempDir()
	orgAPath := filepath.Join(dir, "org-a", "registry.json")
	orgBPath := filepath.Join(dir, "org-b", "registry.json")

	orgA, err := NewCloudRegistryWithPath(orgAPath)
	if err != nil {
		t.Fatalf("org-a init: %v", err)
	}
	orgB, err := NewCloudRegistryWithPath(orgBPath)
	if err != nil {
		t.Fatalf("org-b init: %v", err)
	}

	if err := orgA.Publish(SkillVersion{
		ID: "org-a-skill", Version: "1.0.0", CompatibleClients: []string{"opencode"},
	}); err != nil {
		t.Fatalf("orgA publish: %v", err)
	}
	if err := orgB.Publish(SkillVersion{
		ID: "org-b-skill", Version: "1.0.0", CompatibleClients: []string{"cursor"},
	}); err != nil {
		t.Fatalf("orgB publish: %v", err)
	}

	// Verify isolation: org-a should not see org-b's skills.
	if len(orgA.Versions("org-b-skill")) != 0 {
		t.Fatal("org-a registry should not contain org-b's skill")
	}
	if len(orgB.Versions("org-a-skill")) != 0 {
		t.Fatal("org-b registry should not contain org-a's skill")
	}

	// Each org has exactly 1 skill.
	if len(orgA.List()) != 1 {
		t.Fatalf("expected 1 skill in org-a, got %d", len(orgA.List()))
	}
	if len(orgB.List()) != 1 {
		t.Fatalf("expected 1 skill in org-b, got %d", len(orgB.List()))
	}
}

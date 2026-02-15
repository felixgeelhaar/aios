package marketplace

import "testing"

func TestAddListing(t *testing.T) {
	c := NewCatalog()
	if err := c.Add(Listing{
		SkillID:           "roadmap-reader",
		Version:           "0.1.0",
		Verified:          true,
		Publisher:         "internal",
		CompatibleClients: []string{"opencode", "cursor"},
		BadgeEvidence:     "fixtures-pass + signature",
	}); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if len(c.All()) != 1 {
		t.Fatal("expected one listing")
	}
}

func TestAddListingRejectsMissingCompatibility(t *testing.T) {
	c := NewCatalog()
	if err := c.Add(Listing{SkillID: "roadmap-reader", Version: "0.1.0"}); err == nil {
		t.Fatal("expected compatibility validation error")
	}
}

func TestAddListingRejectsVerifiedWithoutBadgeEvidence(t *testing.T) {
	c := NewCatalog()
	err := c.Add(Listing{
		SkillID:           "roadmap-reader",
		Version:           "0.1.0",
		Verified:          true,
		CompatibleClients: []string{"opencode"},
	})
	if err == nil {
		t.Fatal("expected badge evidence validation error")
	}
}

// AC3: Install contract rejects unsupported client identifiers.
func TestAddListingRejectsUnsupportedClient(t *testing.T) {
	c := NewCatalog()
	err := c.Add(Listing{
		SkillID:           "bad-client-skill",
		Version:           "1.0.0",
		CompatibleClients: []string{"unknown_editor"},
	})
	if err == nil {
		t.Fatal("expected error for unsupported client")
	}
}

// AC1/AC5: Catalog stores multiple listings with compatibility metadata.
func TestCatalogMultipleListings(t *testing.T) {
	c := NewCatalog()
	listings := []Listing{
		{SkillID: "skill-a", Version: "1.0.0", CompatibleClients: []string{"opencode"}},
		{SkillID: "skill-b", Version: "2.0.0", CompatibleClients: []string{"cursor", "windsurf"}},
	}
	for _, l := range listings {
		if err := c.Add(l); err != nil {
			t.Fatalf("add %s: %v", l.SkillID, err)
		}
	}
	all := c.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 listings, got %d", len(all))
	}
	// Verify compatibility metadata is preserved.
	if all[1].CompatibleClients[0] != "cursor" {
		t.Fatalf("expected cursor, got %q", all[1].CompatibleClients[0])
	}
}

// AC8: All() returns defensive copy — mutations do not affect catalog.
func TestCatalogAllReturnsDefensiveCopy(t *testing.T) {
	c := NewCatalog()
	if err := c.Add(Listing{
		SkillID:           "safe-skill",
		Version:           "1.0.0",
		CompatibleClients: []string{"opencode"},
	}); err != nil {
		t.Fatalf("add: %v", err)
	}

	all := c.All()
	// Mutate the returned slice.
	all[0].SkillID = "hacked"
	_ = append(all, Listing{SkillID: "injected"})

	// Original catalog must be unaffected.
	fresh := c.All()
	if len(fresh) != 1 {
		t.Fatalf("expected 1 listing, got %d", len(fresh))
	}
	if fresh[0].SkillID != "safe-skill" {
		t.Fatalf("defensive copy failed: skill ID mutated to %q", fresh[0].SkillID)
	}
}

// AC4: Unverified listing can be added without badge evidence.
func TestAddUnverifiedListingWithoutBadge(t *testing.T) {
	c := NewCatalog()
	err := c.Add(Listing{
		SkillID:           "community-skill",
		Version:           "0.5.0",
		Verified:          false,
		CompatibleClients: []string{"windsurf"},
	})
	if err != nil {
		t.Fatalf("expected no error for unverified listing, got: %v", err)
	}
}

// AC1: Add rejects empty skill ID or version.
func TestAddListingRejectsEmptyIDOrVersion(t *testing.T) {
	tests := []struct {
		name string
		l    Listing
	}{
		{"empty id", Listing{Version: "1.0.0", CompatibleClients: []string{"opencode"}}},
		{"empty version", Listing{SkillID: "x", CompatibleClients: []string{"opencode"}}},
	}
	c := NewCatalog()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := c.Add(tt.l); err == nil {
				t.Fatal("expected error for missing id/version")
			}
		})
	}
}

// AC7 (Marketplace Ecosystem): External developer can publish a verified skill.
func TestExternalDeveloperPublishesVerifiedSkill(t *testing.T) {
	c := NewCatalog()
	err := c.Add(Listing{
		SkillID:           "ext-analytics-tool",
		Version:           "1.0.0",
		Verified:          true,
		Publisher:         "external-developer",
		CompatibleClients: []string{"opencode", "cursor"},
		BadgeEvidence:     "fixtures-pass + signature + review",
	})
	if err != nil {
		t.Fatalf("external developer publish should succeed: %v", err)
	}
	all := c.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 listing, got %d", len(all))
	}
	if all[0].Publisher != "external-developer" {
		t.Fatalf("expected publisher 'external-developer', got %q", all[0].Publisher)
	}
	if !all[0].Verified {
		t.Fatal("expected verified=true")
	}
}

// AC8 (Marketplace Ecosystem): Org installs marketplace skill safely with contract validation.
func TestOrgInstallsSkillWithContractValidation(t *testing.T) {
	c := NewCatalog()
	// Install a skill that passes all contract checks.
	err := c.Add(Listing{
		SkillID:           "marketplace-data-analyzer",
		Version:           "2.1.0",
		Verified:          true,
		Publisher:         "marketplace-vendor",
		CompatibleClients: []string{"opencode", "cursor", "windsurf"},
		BadgeEvidence:     "compatibility-verified",
	})
	if err != nil {
		t.Fatalf("org install should succeed with valid contract: %v", err)
	}

	// Try installing a skill with unsupported client — should fail.
	err = c.Add(Listing{
		SkillID:           "unsafe-skill",
		Version:           "1.0.0",
		Publisher:         "unknown-vendor",
		CompatibleClients: []string{"unsupported_editor"},
	})
	if err == nil {
		t.Fatal("expected contract validation to reject unsupported client")
	}
}

// AC5 (Marketplace Ecosystem): Compatibility metadata preserved in listing.
func TestListingPreservesCompatibilityMetadata(t *testing.T) {
	c := NewCatalog()
	clients := []string{"opencode", "cursor", "windsurf"}
	err := c.Add(Listing{
		SkillID:           "multi-client-skill",
		Version:           "3.0.0",
		CompatibleClients: clients,
		Publisher:         "acme-corp",
	})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	all := c.All()
	if len(all[0].CompatibleClients) != 3 {
		t.Fatalf("expected 3 compatible clients, got %d", len(all[0].CompatibleClients))
	}
	for i, expected := range clients {
		if all[0].CompatibleClients[i] != expected {
			t.Fatalf("client[%d]: expected %q, got %q", i, expected, all[0].CompatibleClients[i])
		}
	}
}

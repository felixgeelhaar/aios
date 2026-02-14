package governance

import (
	"strings"
	"testing"
)

func TestEnforcePublish(t *testing.T) {
	err := EnforcePublish(PublishRequest{
		Role:      RolePublisher,
		SkillID:   "roadmap-reader",
		Version:   "0.1.0",
		ScopeList: []string{"files.readonly"},
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestEnforcePublishRejectsViewer(t *testing.T) {
	err := EnforcePublish(PublishRequest{
		Role:      RoleViewer,
		SkillID:   "roadmap-reader",
		Version:   "0.1.0",
		ScopeList: []string{"files.readonly"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

// AC7: Enforce publish checks role, id, version, and scope together.
func TestEnforcePublish_RejectsEmptyFields(t *testing.T) {
	tests := []struct {
		name string
		req  PublishRequest
		want string
	}{
		{"empty skill id", PublishRequest{Role: RoleAdmin, Version: "1.0.0"}, "skill id and version are required"},
		{"empty version", PublishRequest{Role: RoleAdmin, SkillID: "x"}, "skill id and version are required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnforcePublish(tt.req)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %q", tt.want, err.Error())
			}
		})
	}
}

// AC2/AC7: Enforce publish rejects disallowed scopes even for admin role.
func TestEnforcePublish_RejectsDisallowedScope(t *testing.T) {
	err := EnforcePublish(PublishRequest{
		Role:      RoleAdmin,
		SkillID:   "admin-skill",
		Version:   "1.0.0",
		ScopeList: []string{"files.readonly", "full_access"},
	})
	if err == nil {
		t.Fatal("expected error for disallowed scope")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AC7: Admin can publish with multiple allowed scopes.
func TestEnforcePublish_AdminMultipleScopes(t *testing.T) {
	err := EnforcePublish(PublishRequest{
		Role:      RoleAdmin,
		SkillID:   "multi-scope-skill",
		Version:   "2.0.0",
		ScopeList: []string{"drive", "files.readonly"},
	})
	if err != nil {
		t.Fatalf("expected admin with valid scopes to succeed, got: %v", err)
	}
}

// AC7: Publisher with no scopes passes (scopes are optional).
func TestEnforcePublish_NoScopesAllowed(t *testing.T) {
	err := EnforcePublish(PublishRequest{
		Role:    RolePublisher,
		SkillID: "basic-skill",
		Version: "1.0.0",
	})
	if err != nil {
		t.Fatalf("expected success with no scopes, got: %v", err)
	}
}

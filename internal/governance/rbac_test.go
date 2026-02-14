package governance

import "testing"

func TestCanPublish(t *testing.T) {
	if !CanPublish(RoleAdmin) {
		t.Fatal("admin should be able to publish")
	}
	if CanPublish(RoleViewer) {
		t.Fatal("viewer should not be able to publish")
	}
}

// AC1: Publisher role must be able to publish.
func TestCanPublish_PublisherAllowed(t *testing.T) {
	if !CanPublish(RolePublisher) {
		t.Fatal("publisher role should be able to publish")
	}
}

// AC1: All three roles have distinct publish permissions.
func TestCanPublish_AllRoles(t *testing.T) {
	tests := []struct {
		role Role
		want bool
	}{
		{RoleAdmin, true},
		{RolePublisher, true},
		{RoleViewer, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			got := CanPublish(tt.role)
			if got != tt.want {
				t.Fatalf("CanPublish(%q) = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

// AC2: Allowed scopes pass validation.
func TestValidateConnectorScope_Allowed(t *testing.T) {
	for _, scope := range []string{"drive", "files.readonly"} {
		t.Run(scope, func(t *testing.T) {
			if err := ValidateConnectorScope(scope); err != nil {
				t.Fatalf("expected scope %q to be allowed, got: %v", scope, err)
			}
		})
	}
}

// AC2: Empty scope is rejected.
func TestValidateConnectorScope_RejectsEmpty(t *testing.T) {
	if err := ValidateConnectorScope(""); err == nil {
		t.Fatal("expected error for empty scope")
	}
}

// AC2: Unknown scopes are restricted.
func TestValidateConnectorScope_RejectsUnknown(t *testing.T) {
	rejected := []string{"admin", "write", "delete", "calendar", "full_access"}
	for _, scope := range rejected {
		t.Run(scope, func(t *testing.T) {
			if err := ValidateConnectorScope(scope); err == nil {
				t.Fatalf("expected scope %q to be rejected", scope)
			}
		})
	}
}

package governance

import "fmt"

type Role string

const (
	RoleViewer    Role = "viewer"
	RolePublisher Role = "publisher"
	RoleAdmin     Role = "admin"
)

func CanPublish(role Role) bool {
	return role == RolePublisher || role == RoleAdmin
}

func ValidateConnectorScope(scope string) error {
	if scope == "" {
		return fmt.Errorf("scope is required")
	}
	if scope == "drive" || scope == "files.readonly" {
		return nil
	}
	return fmt.Errorf("scope %q not allowed", scope)
}

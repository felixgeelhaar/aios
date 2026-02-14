package governance

import "fmt"

type PublishRequest struct {
	Role      Role
	SkillID   string
	Version   string
	ScopeList []string
}

func EnforcePublish(req PublishRequest) error {
	if req.SkillID == "" || req.Version == "" {
		return fmt.Errorf("skill id and version are required")
	}
	if !CanPublish(req.Role) {
		return fmt.Errorf("role %q cannot publish", req.Role)
	}
	for _, scope := range req.ScopeList {
		if err := ValidateConnectorScope(scope); err != nil {
			return err
		}
	}
	return nil
}

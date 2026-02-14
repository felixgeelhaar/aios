package runtime

import "time"

type HealthReport struct {
	Status     string
	Ready      bool
	CheckedAt  time.Time
	TokenStore string
	Workspace  string
}

func (r *Runtime) Health() HealthReport {
	typeName := "memory"
	if r.UsesSecureTokenStore() {
		typeName = "keychain"
	}
	return HealthReport{
		Status:     "ok",
		Ready:      r.workspace != "",
		CheckedAt:  time.Now().UTC(),
		TokenStore: typeName,
		Workspace:  r.workspace,
	}
}

package skill

import "fmt"

type Artifact struct {
	ID           string
	Name         string
	Version      string
	PromptPath   string
	InputSchema  string
	OutputSchema string
	Guardrails   []string
}

func (a Artifact) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("id is required")
	}
	if a.Version == "" {
		return fmt.Errorf("version is required")
	}
	if a.InputSchema == "" || a.OutputSchema == "" {
		return fmt.Errorf("schemas are required")
	}
	return nil
}

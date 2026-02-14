package skill

import "fmt"

type Executor struct{}

func NewExecutor() *Executor { return &Executor{} }

func (e *Executor) Execute(a Artifact, input map[string]any) (map[string]any, error) {
	if err := a.Validate(); err != nil {
		return nil, err
	}
	if len(input) == 0 {
		return nil, fmt.Errorf("input is required")
	}
	return map[string]any{
		"skill_id": a.ID,
		"status":   "ok",
	}, nil
}

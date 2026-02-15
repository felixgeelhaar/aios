package skill

import "fmt"

// HandlerFunc processes a skill artifact with input and returns output.
type HandlerFunc func(artifact Artifact, input map[string]any) (map[string]any, error)

// Executor dispatches skill execution to registered handlers.
type Executor struct {
	handlers map[string]HandlerFunc
}

// NewExecutor creates an Executor with an empty handler registry.
func NewExecutor() *Executor {
	return &Executor{handlers: make(map[string]HandlerFunc)}
}

// RegisterHandler associates a handler with a skill ID.
func (e *Executor) RegisterHandler(skillID string, handler HandlerFunc) {
	e.handlers[skillID] = handler
}

// Execute validates the artifact, dispatches to a registered handler if one
// exists, or falls back to the default stub response.
func (e *Executor) Execute(a Artifact, input map[string]any) (map[string]any, error) {
	if err := a.Validate(); err != nil {
		return nil, err
	}
	if len(input) == 0 {
		return nil, fmt.Errorf("input is required")
	}
	if handler, ok := e.handlers[a.ID]; ok {
		return handler(a, input)
	}
	return map[string]any{
		"skill_id": a.ID,
		"status":   "ok",
	}, nil
}

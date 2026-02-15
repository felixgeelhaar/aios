package skill

import (
	"fmt"
	"testing"
)

func TestArtifactValidate(t *testing.T) {
	a := Artifact{ID: "roadmap-reader", Version: "0.1.0", InputSchema: "in.json", OutputSchema: "out.json"}
	if err := a.Validate(); err != nil {
		t.Fatalf("expected valid artifact, got %v", err)
	}
}

// AC1: Artifact must include required fields — table-driven rejection tests.
func TestArtifactValidateRejectsMissingFields(t *testing.T) {
	tests := []struct {
		name     string
		artifact Artifact
		wantErr  string
	}{
		{
			name:     "missing id",
			artifact: Artifact{Version: "1.0.0", InputSchema: "in.json", OutputSchema: "out.json"},
			wantErr:  "id is required",
		},
		{
			name:     "missing version",
			artifact: Artifact{ID: "skill-1", InputSchema: "in.json", OutputSchema: "out.json"},
			wantErr:  "version is required",
		},
		{
			name:     "missing input schema",
			artifact: Artifact{ID: "skill-1", Version: "1.0.0", OutputSchema: "out.json"},
			wantErr:  "schemas are required",
		},
		{
			name:     "missing output schema",
			artifact: Artifact{ID: "skill-1", Version: "1.0.0", InputSchema: "in.json"},
			wantErr:  "schemas are required",
		},
		{
			name:     "missing both schemas",
			artifact: Artifact{ID: "skill-1", Version: "1.0.0"},
			wantErr:  "schemas are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.artifact.Validate()
			if err == nil {
				t.Fatal("expected validation error")
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

// AC1: Valid artifact with all required fields and optional guardrails.
func TestArtifactValidateWithGuardrails(t *testing.T) {
	a := Artifact{
		ID:           "roadmap-reader",
		Version:      "1.0.0",
		InputSchema:  "schema.input.json",
		OutputSchema: "schema.output.json",
		Guardrails:   []string{"no-pii", "content-filter"},
	}
	if err := a.Validate(); err != nil {
		t.Fatalf("expected valid artifact with guardrails, got %v", err)
	}
}

// AC3: Executor defaults to suggest mode — stub returns status "ok".
func TestExecutorExecute(t *testing.T) {
	a := Artifact{ID: "roadmap-reader", Version: "0.1.0", InputSchema: "in.json", OutputSchema: "out.json"}
	e := NewExecutor()
	out, err := e.Execute(a, map[string]any{"query": "x"})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if out["status"] != "ok" {
		t.Fatalf("unexpected output: %#v", out)
	}
}

// AC3: Executor validates artifact before execution.
func TestExecutorRejectsInvalidArtifact(t *testing.T) {
	a := Artifact{} // missing all required fields
	e := NewExecutor()
	_, err := e.Execute(a, map[string]any{"query": "x"})
	if err == nil {
		t.Fatal("expected error for invalid artifact")
	}
}

// AC3: Executor rejects empty input.
func TestExecutorRejectsEmptyInput(t *testing.T) {
	a := Artifact{ID: "skill-1", Version: "1.0.0", InputSchema: "in.json", OutputSchema: "out.json"}
	e := NewExecutor()
	_, err := e.Execute(a, nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestExecutor_RegisteredHandler(t *testing.T) {
	a := Artifact{ID: "custom-skill", Version: "1.0.0", InputSchema: "in.json", OutputSchema: "out.json"}
	e := NewExecutor()
	e.RegisterHandler("custom-skill", func(_ Artifact, input map[string]any) (map[string]any, error) {
		return map[string]any{"echo": input["query"], "custom": true}, nil
	})
	out, err := e.Execute(a, map[string]any{"query": "hello"})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if out["echo"] != "hello" {
		t.Fatalf("expected echo=hello, got %v", out["echo"])
	}
	if out["custom"] != true {
		t.Fatalf("expected custom=true, got %v", out["custom"])
	}
}

func TestExecutor_FallbackWhenNoHandler(t *testing.T) {
	a := Artifact{ID: "unregistered", Version: "1.0.0", InputSchema: "in.json", OutputSchema: "out.json"}
	e := NewExecutor()
	e.RegisterHandler("other-skill", func(_ Artifact, _ map[string]any) (map[string]any, error) {
		return map[string]any{"should": "not reach"}, nil
	})
	out, err := e.Execute(a, map[string]any{"query": "x"})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if out["status"] != "ok" {
		t.Fatalf("expected fallback status=ok, got %v", out["status"])
	}
	if out["skill_id"] != "unregistered" {
		t.Fatalf("expected skill_id=unregistered, got %v", out["skill_id"])
	}
}

func TestExecutor_HandlerError(t *testing.T) {
	a := Artifact{ID: "failing-skill", Version: "1.0.0", InputSchema: "in.json", OutputSchema: "out.json"}
	e := NewExecutor()
	e.RegisterHandler("failing-skill", func(_ Artifact, _ map[string]any) (map[string]any, error) {
		return nil, fmt.Errorf("handler failed")
	})
	_, err := e.Execute(a, map[string]any{"query": "x"})
	if err == nil {
		t.Fatal("expected error from handler")
	}
	if err.Error() != "handler failed" {
		t.Fatalf("expected 'handler failed', got %q", err.Error())
	}
}

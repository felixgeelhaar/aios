package skill

import "testing"

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

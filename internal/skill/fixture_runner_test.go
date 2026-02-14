package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunFixtureSuite(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.input.json"), []byte(`{"type":"object","properties":{"query":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.output.json"), []byte(`{"type":"object","properties":{"skill_id":{"type":"string"},"status":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tests", "fixture_01.json"), []byte(`{"query":"hello"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tests", "expected_01.json"), []byte(`{"status":"ok"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := RunFixtureSuite(dir)
	if err != nil {
		t.Fatalf("run suite: %v", err)
	}
	if len(results) != 1 || !results[0].Passed {
		t.Fatalf("unexpected result: %#v", results)
	}
}

// AC: fixture runner must report FAIL with "output mismatch" when expected
// values do not match executor output.
func TestRunFixtureSuiteReportsMismatch(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte("id: test-skill\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.output.json"), []byte(`{"type":"object","properties":{"status":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// Fixture provides valid input; expected claims status="wrong" but
	// executor stub always returns status="ok".
	if err := os.WriteFile(filepath.Join(dir, "tests", "fixture_01.json"), []byte(`{"q":"hello"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tests", "expected_01.json"), []byte(`{"status":"wrong"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := RunFixtureSuite(dir)
	if err != nil {
		t.Fatalf("run suite: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Fatal("expected fixture to fail due to output mismatch")
	}
	if results[0].Error != "output mismatch" {
		t.Fatalf("expected error 'output mismatch', got %q", results[0].Error)
	}
}

// AC: fixture runner must report each fixture independently when multiple
// fixtures are present — some pass, some fail.
func TestRunFixtureSuiteMultipleFixturesIndependent(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte("id: multi-skill\nversion: 1.0.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.output.json"), []byte(`{"type":"object","properties":{"status":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// fixture_01: expects status="ok" — matches executor stub → PASS
	if err := os.WriteFile(filepath.Join(dir, "tests", "fixture_01.json"), []byte(`{"q":"hello"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tests", "expected_01.json"), []byte(`{"status":"ok"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// fixture_02: expects status="fail" — does NOT match executor stub → FAIL
	if err := os.WriteFile(filepath.Join(dir, "tests", "fixture_02.json"), []byte(`{"q":"world"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tests", "expected_02.json"), []byte(`{"status":"fail"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := RunFixtureSuite(dir)
	if err != nil {
		t.Fatalf("run suite: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Build a map by fixture name for order-independent assertion.
	byName := map[string]FixtureResult{}
	for _, r := range results {
		byName[r.Name] = r
	}

	r1, ok := byName["fixture_01.json"]
	if !ok {
		t.Fatal("missing result for fixture_01.json")
	}
	if !r1.Passed {
		t.Fatalf("fixture_01 should pass, got error: %q", r1.Error)
	}

	r2, ok := byName["fixture_02.json"]
	if !ok {
		t.Fatal("missing result for fixture_02.json")
	}
	if r2.Passed {
		t.Fatal("fixture_02 should fail due to output mismatch")
	}
	if r2.Error != "output mismatch" {
		t.Fatalf("expected 'output mismatch' error for fixture_02, got %q", r2.Error)
	}
}

// AC: fixture result must include fixture name for structured reporting.
func TestFixtureResultIncludesName(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte("id: named-skill\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.output.json"), []byte(`{"type":"object","properties":{"status":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tests", "fixture_alpha.json"), []byte(`{"q":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tests", "expected_alpha.json"), []byte(`{"status":"ok"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := RunFixtureSuite(dir)
	if err != nil {
		t.Fatalf("run suite: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "fixture_alpha.json" {
		t.Fatalf("expected fixture name 'fixture_alpha.json', got %q", results[0].Name)
	}
}

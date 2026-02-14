package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/felixgeelhaar/aios/internal/agents"
	agentregistry "github.com/felixgeelhaar/aios/internal/domain/agentregistry"
)

func TestBuildSyncPlan(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultConfig()
	cfg.ProjectDir = root

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"q":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"a":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildSyncPlan(cfg, skillDir)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	if plan.SkillID != "roadmap-reader" || len(plan.Writes) == 0 {
		t.Fatalf("unexpected plan: %#v", plan)
	}
}

// AC5: sync-plan must not create any files in agent skill directories.
func TestSyncPlanDoesNotMutateAgentDirs(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultConfig()
	cfg.ProjectDir = root

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: safe-skill\nversion: 0.1.0\ninputs:\n  schema: in.json\noutputs:\n  schema: out.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "in.json"), []byte(`{"type":"object","properties":{"x":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "out.json"), []byte(`{"type":"object","properties":{"y":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := BuildSyncPlan(cfg, skillDir)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	// Verify canonical skills directory was not created.
	canonicalDir := filepath.Join(root, agentregistry.CanonicalSkillsDir)
	if _, err := os.Stat(canonicalDir); err == nil {
		t.Fatalf("sync-plan must not create skills directory %q", canonicalDir)
	}
}

// AC8: Plan returns one write target per non-universal agent plus one canonical path.
func TestSyncPlanCoversAllAgents(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultConfig()
	cfg.ProjectDir = root

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: all-agents\nversion: 0.1.0\ninputs:\n  schema: in.json\noutputs:\n  schema: out.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "in.json"), []byte(`{"type":"object","properties":{"x":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "out.json"), []byte(`{"type":"object","properties":{"y":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildSyncPlan(cfg, skillDir)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	allAgents, loadErr := agents.LoadAll()
	if loadErr != nil {
		t.Fatalf("load agents: %v", loadErr)
	}
	nonUniversal := agentregistry.FilterNonUniversal(allAgents)
	// 1 canonical + N non-universal
	expectedWrites := 1 + len(nonUniversal)
	if len(plan.Writes) != expectedWrites {
		t.Fatalf("expected %d write targets (1 canonical + %d non-universal), got %d: %v",
			expectedWrites, len(nonUniversal), len(plan.Writes), plan.Writes)
	}

	// Verify canonical path is present.
	hasCanonical := false
	for _, w := range plan.Writes {
		if strings.Contains(w, agentregistry.CanonicalSkillsDir) {
			hasCanonical = true
			break
		}
	}
	if !hasCanonical {
		t.Fatalf("sync plan missing canonical write target in %v", plan.Writes)
	}
}

// AC2 + AC3: sync-plan (and sync) must reject invalid JSON schemas.
func TestSyncPlanRejectsInvalidSchema(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultConfig()
	cfg.ProjectDir = root

	skillDir := filepath.Join(root, "skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: bad-schema\nversion: 0.1.0\ninputs:\n  schema: in.json\noutputs:\n  schema: out.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Input schema with type=array instead of type=object — should be rejected.
	if err := os.WriteFile(filepath.Join(skillDir, "in.json"), []byte(`{"type":"array","items":{"type":"string"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "out.json"), []byte(`{"type":"object","properties":{"y":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := BuildSyncPlan(cfg, skillDir)
	if err == nil {
		t.Fatal("expected schema validation error for non-object type")
	}
	if !strings.Contains(err.Error(), "type") {
		t.Fatalf("error should mention schema type issue: %v", err)
	}
}

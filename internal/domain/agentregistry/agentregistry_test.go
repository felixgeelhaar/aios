package agentregistry_test

import (
	"testing"

	"github.com/felixgeelhaar/aios/internal/domain/agentregistry"
)

func TestAgentDefinition_Validate_RequiresName(t *testing.T) {
	agent := agentregistry.AgentDefinition{
		DisplayName: "Test",
		SkillsDir:   ".test/skills",
	}
	if err := agent.Validate(); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestAgentDefinition_Validate_RequiresDisplayName(t *testing.T) {
	agent := agentregistry.AgentDefinition{
		Name:      "test",
		SkillsDir: ".test/skills",
	}
	if err := agent.Validate(); err == nil {
		t.Error("expected error for empty display name")
	}
}

func TestAgentDefinition_Validate_RequiresSkillsDir(t *testing.T) {
	agent := agentregistry.AgentDefinition{
		Name:        "test",
		DisplayName: "Test",
	}
	if err := agent.Validate(); err == nil {
		t.Error("expected error for empty skills dir")
	}
}

func TestAgentDefinition_Validate_UniversalMustUseCanonicalDir(t *testing.T) {
	agent := agentregistry.AgentDefinition{
		Name:        "test",
		DisplayName: "Test",
		SkillsDir:   ".test/skills",
		Universal:   true,
	}
	if err := agent.Validate(); err == nil {
		t.Error("expected error for universal agent with non-canonical skills dir")
	}
}

func TestAgentDefinition_Validate_UniversalWithCanonicalDirSucceeds(t *testing.T) {
	agent := agentregistry.AgentDefinition{
		Name:        "opencode",
		DisplayName: "OpenCode",
		SkillsDir:   agentregistry.CanonicalSkillsDir,
		Universal:   true,
	}
	if err := agent.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAgentDefinition_Validate_NonUniversalSucceeds(t *testing.T) {
	agent := agentregistry.AgentDefinition{
		Name:        "cursor",
		DisplayName: "Cursor",
		SkillsDir:   ".cursor/skills",
		Universal:   false,
	}
	if err := agent.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAgentDefinition_IsUniversal(t *testing.T) {
	tests := []struct {
		name     string
		agent    agentregistry.AgentDefinition
		expected bool
	}{
		{
			name:     "universal agent",
			agent:    agentregistry.AgentDefinition{Universal: true},
			expected: true,
		},
		{
			name:     "non-universal agent",
			agent:    agentregistry.AgentDefinition{Universal: false},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.agent.IsUniversal(); got != tt.expected {
				t.Errorf("IsUniversal() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func testAgents() []agentregistry.AgentDefinition {
	return []agentregistry.AgentDefinition{
		{Name: "opencode", DisplayName: "OpenCode", SkillsDir: agentregistry.CanonicalSkillsDir, AltSkillsDirs: []string{".opencode/skills"}, Universal: true},
		{Name: "codex", DisplayName: "Codex", SkillsDir: agentregistry.CanonicalSkillsDir, Universal: true},
		{Name: "cursor", DisplayName: "Cursor", SkillsDir: ".cursor/skills", Universal: false},
		{Name: "claude-code", DisplayName: "Claude Code", SkillsDir: ".claude/skills", Universal: false},
		{Name: "windsurf", DisplayName: "Windsurf", SkillsDir: ".windsurf/skills", Universal: false},
	}
}

func TestFilterUniversal(t *testing.T) {
	agents := testAgents()
	universal := agentregistry.FilterUniversal(agents)
	if len(universal) != 2 {
		t.Fatalf("expected 2 universal agents, got %d", len(universal))
	}
	for _, a := range universal {
		if !a.Universal {
			t.Errorf("expected universal agent, got %q", a.Name)
		}
	}
}

func TestFilterNonUniversal(t *testing.T) {
	agents := testAgents()
	nonUniversal := agentregistry.FilterNonUniversal(agents)
	if len(nonUniversal) != 3 {
		t.Fatalf("expected 3 non-universal agents, got %d", len(nonUniversal))
	}
	for _, a := range nonUniversal {
		if a.Universal {
			t.Errorf("expected non-universal agent, got %q", a.Name)
		}
	}
}

func TestFilterUniversal_EmptyInput(t *testing.T) {
	result := agentregistry.FilterUniversal(nil)
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestFilterNonUniversal_EmptyInput(t *testing.T) {
	result := agentregistry.FilterNonUniversal(nil)
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestResolveByNames_ValidNames(t *testing.T) {
	agents := testAgents()
	resolved, err := agentregistry.ResolveByNames(agents, []string{"cursor", "opencode"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resolved) != 2 {
		t.Fatalf("expected 2 resolved agents, got %d", len(resolved))
	}
	if resolved[0].Name != "cursor" {
		t.Errorf("expected cursor first, got %q", resolved[0].Name)
	}
	if resolved[1].Name != "opencode" {
		t.Errorf("expected opencode second, got %q", resolved[1].Name)
	}
}

func TestResolveByNames_UnknownName(t *testing.T) {
	agents := testAgents()
	_, err := agentregistry.ResolveByNames(agents, []string{"nonexistent"})
	if err == nil {
		t.Error("expected error for unknown agent name")
	}
}

func TestResolveByNames_EmptyNames(t *testing.T) {
	agents := testAgents()
	resolved, err := agentregistry.ResolveByNames(agents, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != nil {
		t.Errorf("expected nil for empty names, got %v", resolved)
	}
}

func TestAllSkillsDirs_IncludesCanonicalFirst(t *testing.T) {
	agents := testAgents()
	dirs := agentregistry.AllSkillsDirs(agents)
	if len(dirs) == 0 {
		t.Fatal("expected at least one directory")
	}
	if dirs[0] != agentregistry.CanonicalSkillsDir {
		t.Errorf("expected canonical dir first, got %q", dirs[0])
	}
}

func TestAllSkillsDirs_Deduplicates(t *testing.T) {
	agents := testAgents()
	dirs := agentregistry.AllSkillsDirs(agents)
	seen := make(map[string]bool)
	for _, d := range dirs {
		if seen[d] {
			t.Errorf("duplicate directory: %q", d)
		}
		seen[d] = true
	}
}

func TestAllSkillsDirs_IncludesAltDirs(t *testing.T) {
	agents := testAgents()
	dirs := agentregistry.AllSkillsDirs(agents)
	found := false
	for _, d := range dirs {
		if d == ".opencode/skills" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected .opencode/skills in alt dirs")
	}
}

func TestAllSkillsDirs_IncludesAllAgentDirs(t *testing.T) {
	agents := testAgents()
	dirs := agentregistry.AllSkillsDirs(agents)
	// Should have: .agents/skills, .opencode/skills, .cursor/skills, .claude/skills, .windsurf/skills
	if len(dirs) != 5 {
		t.Errorf("expected 5 unique directories, got %d: %v", len(dirs), dirs)
	}
}

func TestAllSkillsDirs_EmptyInput(t *testing.T) {
	dirs := agentregistry.AllSkillsDirs(nil)
	if len(dirs) != 1 {
		t.Fatalf("expected 1 directory (canonical), got %d", len(dirs))
	}
	if dirs[0] != agentregistry.CanonicalSkillsDir {
		t.Errorf("expected canonical dir, got %q", dirs[0])
	}
}

func TestCanonicalSkillsDir_Value(t *testing.T) {
	if agentregistry.CanonicalSkillsDir != ".agents/skills" {
		t.Errorf("expected .agents/skills, got %q", agentregistry.CanonicalSkillsDir)
	}
}

package agents

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/felixgeelhaar/aios/internal/domain/agentregistry"
)

func TestLoadAll_ReturnsNineAgents(t *testing.T) {
	agents, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}
	if len(agents) != 9 {
		t.Errorf("expected 9 agents, got %d", len(agents))
	}
}

func TestLoadAll_AllAgentsValid(t *testing.T) {
	agents, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}
	for _, a := range agents {
		if err := a.Validate(); err != nil {
			t.Errorf("agent %q failed validation: %v", a.Name, err)
		}
	}
}

func TestLoadAll_UniversalAgentsUseCanonicalDir(t *testing.T) {
	agents, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}
	for _, a := range agents {
		if a.Universal && a.SkillsDir != agentregistry.CanonicalSkillsDir {
			t.Errorf("universal agent %q has skills dir %q, expected %q",
				a.Name, a.SkillsDir, agentregistry.CanonicalSkillsDir)
		}
	}
}

func TestLoadAll_ExpectedUniversalAgents(t *testing.T) {
	agents, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}
	expected := map[string]bool{
		"opencode":       true,
		"codex":          true,
		"gemini-cli":     true,
		"github-copilot": true,
	}
	for _, a := range agents {
		if a.Universal && !expected[a.Name] {
			t.Errorf("unexpected universal agent: %q", a.Name)
		}
		if !a.Universal && expected[a.Name] {
			t.Errorf("expected %q to be universal", a.Name)
		}
	}
}

func TestLoadAll_ExpectedNonUniversalAgents(t *testing.T) {
	agents, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}
	expected := map[string]bool{
		"claude-code": true,
		"cursor":      true,
		"goose":       true,
		"windsurf":    true,
		"cline":       true,
	}
	for _, a := range agents {
		if !a.Universal && !expected[a.Name] {
			t.Errorf("unexpected non-universal agent: %q", a.Name)
		}
	}
}

func TestLoadAll_AllAgentsHaveDetectPaths(t *testing.T) {
	agents, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}
	for _, a := range agents {
		if len(a.DetectPaths) == 0 {
			t.Errorf("agent %q has no detect paths", a.Name)
		}
	}
}

func TestLoadAll_NoDuplicateNames(t *testing.T) {
	agents, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}
	seen := make(map[string]bool)
	for _, a := range agents {
		if seen[a.Name] {
			t.Errorf("duplicate agent name: %q", a.Name)
		}
		seen[a.Name] = true
	}
}

func TestLoadFrom_InvalidJSON(t *testing.T) {
	_, err := loadFrom([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadFrom_InvalidAgent(t *testing.T) {
	data, _ := json.Marshal([]agentJSON{
		{Name: "", DisplayName: "Test", SkillsDir: ".test"},
	})
	_, err := loadFrom(data)
	if err == nil {
		t.Error("expected error for invalid agent (empty name)")
	}
}

func TestDetectInstalled_EmptyInput(t *testing.T) {
	result := DetectInstalled(nil)
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestDetectInFolder_FindsProjectSkillDir(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".cursor", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}

	agents := []agentregistry.AgentDefinition{
		{Name: "cursor", DisplayName: "Cursor", SkillsDir: ".cursor/skills"},
		{Name: "opencode", DisplayName: "OpenCode", SkillsDir: agentregistry.CanonicalSkillsDir, Universal: true},
	}

	detected := DetectInFolder(agents, tmp)
	if len(detected) != 1 {
		t.Fatalf("expected 1 detected agent, got %d", len(detected))
	}
	if detected[0].Name != "cursor" {
		t.Errorf("expected cursor, got %q", detected[0].Name)
	}
}

func TestDetectInFolder_FindsAltSkillDir(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".opencode", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}

	agents := []agentregistry.AgentDefinition{
		{Name: "opencode", DisplayName: "OpenCode", SkillsDir: agentregistry.CanonicalSkillsDir, AltSkillsDirs: []string{".opencode/skills"}, Universal: true},
	}

	detected := DetectInFolder(agents, tmp)
	if len(detected) != 1 {
		t.Fatalf("expected 1 detected agent, got %d", len(detected))
	}
}

func TestResolveProjectSkillsDir(t *testing.T) {
	agent := agentregistry.AgentDefinition{
		Name:      "cursor",
		SkillsDir: ".cursor/skills",
	}
	result := ResolveProjectSkillsDir(agent, "/home/user/project")
	expected := filepath.Join("/home/user/project", ".cursor/skills")
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpandPath_Tilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	result := expandPath("~/test")
	expected := filepath.Join(home, "test")
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpandPath_TildeOnly(t *testing.T) {
	home, _ := os.UserHomeDir()
	result := expandPath("~")
	if result != home {
		t.Errorf("expected %q, got %q", home, result)
	}
}

func TestExpandPath_XDGConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")
	result := expandPath("$XDG_CONFIG/opencode")
	expected := filepath.Join("/custom/config", "opencode")
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpandPath_XDGConfigDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home, _ := os.UserHomeDir()
	result := expandPath("$XDG_CONFIG/opencode")
	expected := filepath.Join(home, ".config", "opencode")
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpandPath_EnvVar(t *testing.T) {
	t.Setenv("CODEX_HOME", "/opt/codex")
	result := expandPath("$CODEX_HOME/skills")
	if result != filepath.Join("/opt/codex", "skills") {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestExpandPath_NoExpansionNeeded(t *testing.T) {
	result := expandPath("/absolute/path")
	if result != "/absolute/path" {
		t.Errorf("expected unchanged path, got %q", result)
	}
}

func TestResolveGlobalSkillsDir_ExpandsTilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	agent := agentregistry.AgentDefinition{
		Name:            "test",
		DisplayName:     "Test",
		SkillsDir:       ".test/skills",
		GlobalSkillsDir: "~/.test/skills",
		DetectPaths:     []string{"~/test"},
	}
	result := ResolveGlobalSkillsDir(agent)
	expected := filepath.Join(home, ".test", "skills")
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestResolveGlobalSkillsDir_ExpandsEnvVar(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom")
	agent := agentregistry.AgentDefinition{
		Name:            "test",
		DisplayName:     "Test",
		SkillsDir:       ".test/skills",
		GlobalSkillsDir: "$XDG_CONFIG/test/skills",
		DetectPaths:     []string{"/tmp"},
	}
	result := ResolveGlobalSkillsDir(agent)
	expected := filepath.Join("/custom", "test", "skills")
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestIsDetected_WithExistingDir(t *testing.T) {
	tmp := t.TempDir()
	agent := agentregistry.AgentDefinition{
		Name:        "test",
		DisplayName: "Test",
		SkillsDir:   ".test/skills",
		DetectPaths: []string{tmp},
	}
	if !isDetected(agent) {
		t.Error("expected agent to be detected when detect path exists")
	}
}

func TestIsDetected_WithNonExistentDir(t *testing.T) {
	agent := agentregistry.AgentDefinition{
		Name:        "test",
		DisplayName: "Test",
		SkillsDir:   ".test/skills",
		DetectPaths: []string{"/nonexistent/path/that/does/not/exist"},
	}
	if isDetected(agent) {
		t.Error("expected agent not to be detected when detect path does not exist")
	}
}

func TestIsDetected_EmptyDetectPaths(t *testing.T) {
	agent := agentregistry.AgentDefinition{
		Name:        "test",
		DisplayName: "Test",
		SkillsDir:   ".test/skills",
	}
	if isDetected(agent) {
		t.Error("expected agent not to be detected with no detect paths")
	}
}

func TestDetectInstalled_FindsAgentsWithExistingPaths(t *testing.T) {
	tmp := t.TempDir()
	agents := []agentregistry.AgentDefinition{
		{Name: "found", DisplayName: "Found", SkillsDir: ".found/skills", DetectPaths: []string{tmp}},
		{Name: "missing", DisplayName: "Missing", SkillsDir: ".missing/skills", DetectPaths: []string{"/does/not/exist"}},
	}
	detected := DetectInstalled(agents)
	if len(detected) != 1 {
		t.Fatalf("expected 1 detected, got %d", len(detected))
	}
	if detected[0].Name != "found" {
		t.Errorf("expected 'found', got %q", detected[0].Name)
	}
}

func TestDetectInFolder_FallsBackToGlobalDetect(t *testing.T) {
	tmp := t.TempDir()
	detectDir := t.TempDir() // This directory exists, so global detect succeeds.

	agents := []agentregistry.AgentDefinition{
		{Name: "global-agent", DisplayName: "Global", SkillsDir: ".nonexistent/skills", DetectPaths: []string{detectDir}},
	}

	detected := DetectInFolder(agents, tmp)
	if len(detected) != 1 {
		t.Fatalf("expected 1 detected via global fallback, got %d", len(detected))
	}
	if detected[0].Name != "global-agent" {
		t.Errorf("expected 'global-agent', got %q", detected[0].Name)
	}
}

func TestDetectInFolder_NoMatch(t *testing.T) {
	tmp := t.TempDir()
	agents := []agentregistry.AgentDefinition{
		{Name: "none", DisplayName: "None", SkillsDir: ".nowhere/skills", DetectPaths: []string{"/does/not/exist"}},
	}
	detected := DetectInFolder(agents, tmp)
	if len(detected) != 0 {
		t.Errorf("expected 0 detected, got %d", len(detected))
	}
}

func TestDirExists_True(t *testing.T) {
	tmp := t.TempDir()
	if !dirExists(tmp) {
		t.Error("expected true for existing directory")
	}
}

func TestDirExists_False(t *testing.T) {
	if dirExists("/nonexistent/path/for/testing") {
		t.Error("expected false for nonexistent path")
	}
}

func TestDirExists_FileNotDir(t *testing.T) {
	f := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if dirExists(f) {
		t.Error("expected false for regular file")
	}
}

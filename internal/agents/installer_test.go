package agents

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/felixgeelhaar/aios/internal/domain/agentregistry"
)

func testAgentDefs() []agentregistry.AgentDefinition {
	return []agentregistry.AgentDefinition{
		{Name: "opencode", DisplayName: "OpenCode", SkillsDir: agentregistry.CanonicalSkillsDir, Universal: true},
		{Name: "codex", DisplayName: "Codex", SkillsDir: agentregistry.CanonicalSkillsDir, Universal: true},
		{Name: "cursor", DisplayName: "Cursor", SkillsDir: ".cursor/skills", Universal: false},
		{Name: "claude-code", DisplayName: "Claude Code", SkillsDir: ".claude/skills", Universal: false},
	}
}

func TestInstallSkill_CreatesCanonicalDir(t *testing.T) {
	tmp := t.TempDir()
	si := NewSkillInstaller(testAgentDefs())

	result, err := si.InstallSkill("test-skill", InstallOptions{ProjectDir: tmp})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	canonicalDir := filepath.Join(tmp, agentregistry.CanonicalSkillsDir, "test-skill")
	if _, err := os.Stat(canonicalDir); os.IsNotExist(err) {
		t.Error("canonical directory was not created")
	}
	if result.CanonicalPath != canonicalDir {
		t.Errorf("expected canonical path %q, got %q", canonicalDir, result.CanonicalPath)
	}
}

func TestInstallSkill_WritesSkillMarker(t *testing.T) {
	tmp := t.TempDir()
	si := NewSkillInstaller(testAgentDefs())

	_, err := si.InstallSkill("test-skill", InstallOptions{ProjectDir: tmp})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	markerPath := filepath.Join(tmp, agentregistry.CanonicalSkillsDir, "test-skill", "SKILL.md")
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Error("SKILL.md was not created")
	}
}

func TestInstallSkill_CreatesSymlinksForNonUniversal(t *testing.T) {
	tmp := t.TempDir()
	si := NewSkillInstaller(testAgentDefs())

	_, err := si.InstallSkill("test-skill", InstallOptions{ProjectDir: tmp})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check cursor symlink.
	cursorLink := filepath.Join(tmp, ".cursor", "skills", "test-skill")
	info, err := os.Lstat(cursorLink)
	if err != nil {
		t.Fatalf("cursor symlink not found: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("cursor path is not a symlink")
	}

	// Check claude-code symlink.
	claudeLink := filepath.Join(tmp, ".claude", "skills", "test-skill")
	info, err = os.Lstat(claudeLink)
	if err != nil {
		t.Fatalf("claude-code symlink not found: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("claude-code path is not a symlink")
	}
}

func TestInstallSkill_SymlinkPointsToCanonical(t *testing.T) {
	tmp := t.TempDir()
	si := NewSkillInstaller(testAgentDefs())

	_, err := si.InstallSkill("test-skill", InstallOptions{ProjectDir: tmp})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cursorLink := filepath.Join(tmp, ".cursor", "skills", "test-skill")
	target, err := os.Readlink(cursorLink)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}

	// Resolve the symlink to verify it points to the canonical dir.
	resolved := filepath.Join(filepath.Dir(cursorLink), target)
	canonicalDir := filepath.Join(tmp, agentregistry.CanonicalSkillsDir, "test-skill")

	resolvedAbs, _ := filepath.Abs(resolved)
	canonicalAbs, _ := filepath.Abs(canonicalDir)
	if resolvedAbs != canonicalAbs {
		t.Errorf("symlink resolves to %q, expected %q", resolvedAbs, canonicalAbs)
	}
}

func TestInstallSkill_IncludesAllAgentsInResult(t *testing.T) {
	tmp := t.TempDir()
	si := NewSkillInstaller(testAgentDefs())

	result, err := si.InstallSkill("test-skill", InstallOptions{ProjectDir: tmp})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Agents) != 4 {
		t.Errorf("expected 4 agents, got %d: %v", len(result.Agents), result.Agents)
	}
}

func TestInstallSkill_SelectiveTargeting(t *testing.T) {
	tmp := t.TempDir()
	si := NewSkillInstaller(testAgentDefs())
	targets := []agentregistry.AgentDefinition{
		{Name: "cursor", DisplayName: "Cursor", SkillsDir: ".cursor/skills", Universal: false},
	}

	result, err := si.InstallSkill("test-skill", InstallOptions{
		ProjectDir:   tmp,
		TargetAgents: targets,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(result.Agents))
	}

	// Cursor symlink should exist.
	cursorLink := filepath.Join(tmp, ".cursor", "skills", "test-skill")
	if _, err := os.Lstat(cursorLink); os.IsNotExist(err) {
		t.Error("cursor symlink should exist")
	}

	// Claude symlink should NOT exist.
	claudeLink := filepath.Join(tmp, ".claude", "skills", "test-skill")
	if _, err := os.Lstat(claudeLink); !os.IsNotExist(err) {
		t.Error("claude-code symlink should not exist")
	}
}

func TestInstallSkill_EmptySkillID(t *testing.T) {
	si := NewSkillInstaller(testAgentDefs())
	_, err := si.InstallSkill("", InstallOptions{ProjectDir: "/tmp"})
	if err == nil {
		t.Error("expected error for empty skill ID")
	}
}

func TestInstallSkill_EmptyProjectDir(t *testing.T) {
	si := NewSkillInstaller(testAgentDefs())
	_, err := si.InstallSkill("test", InstallOptions{})
	if err == nil {
		t.Error("expected error for empty project dir")
	}
}

func TestInstallSkill_DoesNotOverwriteExistingSkillMd(t *testing.T) {
	tmp := t.TempDir()
	si := NewSkillInstaller(testAgentDefs())

	// Pre-create SKILL.md with custom content.
	canonicalDir := filepath.Join(tmp, agentregistry.CanonicalSkillsDir, "test-skill")
	if err := os.MkdirAll(canonicalDir, 0o755); err != nil {
		t.Fatal(err)
	}
	customContent := "---\nname: custom\n---\nCustom content"
	if err := os.WriteFile(filepath.Join(canonicalDir, "SKILL.md"), []byte(customContent), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := si.InstallSkill("test-skill", InstallOptions{ProjectDir: tmp})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(canonicalDir, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != customContent {
		t.Error("SKILL.md was overwritten")
	}
}

func TestUninstallSkill_RemovesCanonicalAndSymlinks(t *testing.T) {
	tmp := t.TempDir()
	si := NewSkillInstaller(testAgentDefs())

	// Install first.
	_, err := si.InstallSkill("test-skill", InstallOptions{ProjectDir: tmp})
	if err != nil {
		t.Fatalf("install: %v", err)
	}

	// Uninstall.
	if err := si.UninstallSkill("test-skill", tmp); err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	// Canonical dir should be gone.
	canonicalDir := filepath.Join(tmp, agentregistry.CanonicalSkillsDir, "test-skill")
	if _, err := os.Stat(canonicalDir); !os.IsNotExist(err) {
		t.Error("canonical directory should be removed")
	}

	// Symlinks should be gone.
	cursorLink := filepath.Join(tmp, ".cursor", "skills", "test-skill")
	if _, err := os.Lstat(cursorLink); !os.IsNotExist(err) {
		t.Error("cursor symlink should be removed")
	}
}

func TestUninstallSkill_EmptySkillID(t *testing.T) {
	si := NewSkillInstaller(testAgentDefs())
	if err := si.UninstallSkill("", "/tmp"); err == nil {
		t.Error("expected error for empty skill ID")
	}
}

func TestUninstallSkill_EmptyProjectDir(t *testing.T) {
	si := NewSkillInstaller(testAgentDefs())
	if err := si.UninstallSkill("test", ""); err == nil {
		t.Error("expected error for empty project dir")
	}
}

func TestUninstallSkill_NonexistentSkill(t *testing.T) {
	tmp := t.TempDir()
	si := NewSkillInstaller(testAgentDefs())

	// Should not error on non-existent skill.
	if err := si.UninstallSkill("nonexistent", tmp); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPlanWriteTargets_ReturnsCanonicalAndAgentPaths(t *testing.T) {
	si := NewSkillInstaller(testAgentDefs())
	targets := si.PlanWriteTargets("test-skill", "/project")

	if len(targets) != 3 {
		t.Fatalf("expected 3 targets (canonical + 2 non-universal), got %d: %v", len(targets), targets)
	}
	if targets[0] != filepath.Join("/project", agentregistry.CanonicalSkillsDir, "test-skill") {
		t.Errorf("unexpected canonical target: %q", targets[0])
	}
}

func TestCollectInstalledSkills_FindsCanonicalSkills(t *testing.T) {
	tmp := t.TempDir()
	si := NewSkillInstaller(testAgentDefs())

	// Install two skills.
	if _, err := si.InstallSkill("alpha", InstallOptions{ProjectDir: tmp}); err != nil {
		t.Fatal(err)
	}
	if _, err := si.InstallSkill("beta", InstallOptions{ProjectDir: tmp}); err != nil {
		t.Fatal(err)
	}

	skills, err := si.CollectInstalledSkills(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
	if skills[0] != "alpha" || skills[1] != "beta" {
		t.Errorf("unexpected order: %v", skills)
	}
}

func TestCollectInstalledSkills_EmptyProject(t *testing.T) {
	tmp := t.TempDir()
	si := NewSkillInstaller(testAgentDefs())

	skills, err := si.CollectInstalledSkills(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(skills))
	}
}

func TestSanitizeName_LowercasesAndReplaces(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"My Skill", "my-skill"},
		{"hello_world", "hello-world"},
		{"UPPER", "upper"},
		{"a.b.c", "a-b-c"},
		{"---trim---", "trim"},
		{"", "unnamed-skill"},
		{"valid-name", "valid-name"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SanitizeName(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeName_LongName(t *testing.T) {
	// Names longer than 255 chars should be truncated.
	long := ""
	for i := 0; i < 300; i++ {
		long += "a"
	}
	got := SanitizeName(long)
	if len(got) != 255 {
		t.Errorf("expected length 255, got %d", len(got))
	}
}

func TestSanitizeName_TrimsDots(t *testing.T) {
	got := SanitizeName("...test...")
	if got != "test" {
		t.Errorf("expected %q, got %q", "test", got)
	}
}

func TestCopyDirectory_CopiesFilesAndDirs(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dest")

	// Create source structure: file.txt, sub/nested.txt
	if err := os.WriteFile(filepath.Join(src, "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "sub", "nested.txt"), []byte("world"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := copyDirectory(src, dst); err != nil {
		t.Fatalf("copyDirectory: %v", err)
	}

	// Verify top-level file.
	data, err := os.ReadFile(filepath.Join(dst, "file.txt"))
	if err != nil {
		t.Fatalf("reading copied file: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("expected %q, got %q", "hello", string(data))
	}

	// Verify nested file.
	data, err = os.ReadFile(filepath.Join(dst, "sub", "nested.txt"))
	if err != nil {
		t.Fatalf("reading nested copied file: %v", err)
	}
	if string(data) != "world" {
		t.Errorf("expected %q, got %q", "world", string(data))
	}
}

func TestCopyDirectory_SkipsDotFiles(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dest")

	// Create a hidden file and a hidden directory.
	if err := os.WriteFile(filepath.Join(src, ".hidden"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(src, ".hiddendir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, ".hiddendir", "file.txt"), []byte("inside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "visible.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := copyDirectory(src, dst); err != nil {
		t.Fatalf("copyDirectory: %v", err)
	}

	// Hidden file should NOT be copied.
	if _, err := os.Stat(filepath.Join(dst, ".hidden")); !os.IsNotExist(err) {
		t.Error("hidden file should not be copied")
	}

	// Hidden directory should NOT be copied.
	if _, err := os.Stat(filepath.Join(dst, ".hiddendir")); !os.IsNotExist(err) {
		t.Error("hidden directory should not be copied")
	}

	// Visible file should be copied.
	if _, err := os.Stat(filepath.Join(dst, "visible.txt")); os.IsNotExist(err) {
		t.Error("visible file should be copied")
	}
}

func TestCopyFile_CopiesContent(t *testing.T) {
	src := filepath.Join(t.TempDir(), "source.txt")
	dst := filepath.Join(t.TempDir(), "dest.txt")

	if err := os.WriteFile(src, []byte("copy me"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("reading dest: %v", err)
	}
	if string(data) != "copy me" {
		t.Errorf("expected %q, got %q", "copy me", string(data))
	}
}

func TestCopyFile_NonexistentSource(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "dest.txt")
	if err := copyFile("/nonexistent/path/file.txt", dst); err == nil {
		t.Error("expected error for nonexistent source")
	}
}

func TestInstallSkill_ReplacesExistingSymlink(t *testing.T) {
	tmp := t.TempDir()
	si := NewSkillInstaller(testAgentDefs())

	// Install once.
	if _, err := si.InstallSkill("test-skill", InstallOptions{ProjectDir: tmp}); err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Install again — should succeed (replaces existing symlink).
	result, err := si.InstallSkill("test-skill", InstallOptions{ProjectDir: tmp})
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	if len(result.Agents) != 4 {
		t.Errorf("expected 4 agents, got %d", len(result.Agents))
	}
}

func TestSortStrings(t *testing.T) {
	input := []string{"charlie", "alpha", "bravo"}
	sortStrings(input)
	expected := []string{"alpha", "bravo", "charlie"}
	for i, v := range input {
		if v != expected[i] {
			t.Errorf("index %d: got %q, want %q", i, v, expected[i])
		}
	}
}

func TestSortStrings_Empty(t *testing.T) {
	var input []string
	sortStrings(input) // should not panic
}

func TestSortStrings_Single(t *testing.T) {
	input := []string{"only"}
	sortStrings(input)
	if input[0] != "only" {
		t.Errorf("expected %q, got %q", "only", input[0])
	}
}

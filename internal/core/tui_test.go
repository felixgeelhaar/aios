package core

import (
	"bytes"
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	domainprojectinventory "github.com/felixgeelhaar/aios/internal/domain/projectinventory"
	domainskillsync "github.com/felixgeelhaar/aios/internal/domain/skillsync"
	domainworkspace "github.com/felixgeelhaar/aios/internal/domain/workspaceorchestration"
)

func TestTUISkillInitFlow(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := CLI{Out: buf}
	cli.In = strings.NewReader("4\n1\n/tmp/skill\n\nq\n")
	var got string
	cli.InitSkill = func(dir string) error {
		got = dir
		return nil
	}
	cli.ListProjects = func(context.Context) ([]domainprojectinventory.Project, error) { return nil, nil }
	cli.ValidateWorkspace = func(context.Context) (domainworkspace.ValidationResult, error) {
		return domainworkspace.ValidationResult{}, nil
	}
	cli.RepairWorkspace = func(context.Context) (domainworkspace.RepairResult, error) {
		return domainworkspace.RepairResult{}, nil
	}

	if err := cli.RunTUI(context.Background()); err != nil {
		t.Fatalf("tui failed: %v", err)
	}
	if got != "/tmp/skill" {
		t.Fatalf("expected init skill dir /tmp/skill, got %q", got)
	}
}

func TestTUISkillSyncFlow(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := CLI{Out: buf}
	cli.In = strings.NewReader("4\n2\n/tmp/skill\n\nq\n")
	cli.SyncSkill = func(context.Context, domainskillsync.SyncSkillCommand) (string, error) {
		return "roadmap-reader", nil
	}
	cli.ListProjects = func(context.Context) ([]domainprojectinventory.Project, error) { return nil, nil }
	cli.ValidateWorkspace = func(context.Context) (domainworkspace.ValidationResult, error) {
		return domainworkspace.ValidationResult{}, nil
	}
	cli.RepairWorkspace = func(context.Context) (domainworkspace.RepairResult, error) {
		return domainworkspace.RepairResult{}, nil
	}

	if err := cli.RunTUI(context.Background()); err != nil {
		t.Fatalf("tui failed: %v", err)
	}
	if !strings.Contains(buf.String(), "sync completed for skill roadmap-reader") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestTUIKeyToIndex(t *testing.T) {
	tests := []struct {
		key      string
		expected int
	}{
		{"1", 0},
		{"2", 1},
		{"9", 8},
		{"0", -1},
		{"a", -1},
		{"", -1},
		{"10", -1},
	}

	for _, tt := range tests {
		result := keyToIndex(tt.key)
		if result != tt.expected {
			t.Errorf("keyToIndex(%q) = %d, want %d", tt.key, result, tt.expected)
		}
	}
}

func TestTUIIsNumericKey(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"1", true},
		{"5", true},
		{"9", true},
		{"0", false},
		{"a", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isNumericKey(tt.key)
		if result != tt.expected {
			t.Errorf("isNumericKey(%q) = %v, want %v", tt.key, result, tt.expected)
		}
	}
}

func TestTUIMainMenuItems(t *testing.T) {
	model := tuiModel{}
	items := model.mainMenuItems()
	if len(items) != 5 {
		t.Errorf("mainMenuItems returned %d items, want 5", len(items))
	}
	if items[0] != "Projects" {
		t.Errorf("first item = %q, want 'Projects'", items[0])
	}
}

func TestTUISkillsMenuItems(t *testing.T) {
	model := tuiModel{}
	items := model.skillsMenuItems()
	if len(items) != 3 {
		t.Errorf("skillsMenuItems returned %d items, want 3", len(items))
	}
	if items[0] != "Init skill" {
		t.Errorf("first item = %q, want 'Init skill'", items[0])
	}
}

func TestTUIUpdateMenuQuit(t *testing.T) {
	model := tuiModel{cursor: 0}
	_, _ = model.updateMenu("q", []string{"a", "b"}, func(int) (tea.Model, tea.Cmd) {
		return model, func() tea.Msg { return nil }
	})
}

func TestTUIUpdateMenuUp(t *testing.T) {
	model := tuiModel{cursor: 1}
	newModel, _ := model.updateMenu("up", []string{"a", "b"}, func(int) (tea.Model, tea.Cmd) {
		return model, nil
	})
	if newModel.(tuiModel).cursor != 0 {
		t.Errorf("cursor should be 0 after up, got %d", newModel.(tuiModel).cursor)
	}
}

func TestTUIUpdateMenuDown(t *testing.T) {
	model := tuiModel{cursor: 0}
	newModel, _ := model.updateMenu("down", []string{"a", "b"}, func(int) (tea.Model, tea.Cmd) {
		return model, nil
	})
	if newModel.(tuiModel).cursor != 1 {
		t.Errorf("cursor should be 1 after down, got %d", newModel.(tuiModel).cursor)
	}
}

func TestTUIUpdateMenuInvalidKey(t *testing.T) {
	model := tuiModel{cursor: 0, status: ""}
	newModel, _ := model.updateMenu("x", []string{"a", "b"}, func(int) (tea.Model, tea.Cmd) {
		return model, nil
	})
	_ = newModel
}

func TestTUIUpdateMenuNumericKeyError(t *testing.T) {
	model := tuiModel{cursor: 0, status: ""}
	newModel, _ := model.updateMenu("5", []string{"a", "b"}, func(int) (tea.Model, tea.Cmd) {
		return model, nil
	})
	if newModel.(tuiModel).status != "error" {
		t.Errorf("expected error status, got %q", newModel.(tuiModel).status)
	}
}

func TestTUIUpdateMenuNumericKey(t *testing.T) {
	model := tuiModel{cursor: 0}
	_, _ = model.updateMenu("2", []string{"a", "b"}, func(idx int) (tea.Model, tea.Cmd) {
		if idx != 1 {
			t.Errorf("expected idx 1, got %d", idx)
		}
		return model, nil
	})
}

func TestTUIViewMainScreen(t *testing.T) {
	model := tuiModel{screen: screenMain, cursor: 0}
	view := model.View()
	if !strings.Contains(view, "AIOS Operations Console") {
		t.Errorf("expected header in view, got %q", view)
	}
}

func TestTUIViewSkillsScreen(t *testing.T) {
	model := tuiModel{screen: screenSkills, cursor: 0}
	view := model.View()
	if !strings.Contains(view, "Skills") {
		t.Errorf("expected Skills in view, got %q", view)
	}
}

func TestTUIViewSkillInitScreen(t *testing.T) {
	model := tuiModel{screen: screenSkillInit, input: "test-skill"}
	view := model.View()
	if !strings.Contains(view, "test-skill") {
		t.Errorf("expected input in view, got %q", view)
	}
}

func TestTUIViewSuccessMessage(t *testing.T) {
	model := tuiModel{screen: screenMain, status: "success", message: "operation succeeded"}
	view := model.View()
	if !strings.Contains(view, "operation succeeded") {
		t.Errorf("expected message in view, got %q", view)
	}
}

func TestTUIViewErrorMessage(t *testing.T) {
	model := tuiModel{screen: screenMain, status: "error", message: "something failed"}
	view := model.View()
	if !strings.Contains(view, "something failed") {
		t.Errorf("expected error in view, got %q", view)
	}
}

func TestTUIInit(t *testing.T) {
	model := tuiModel{}
	cmd := model.Init()
	if cmd != nil {
		t.Errorf("Init should return nil cmd, got %v", cmd)
	}
}

func TestTUIUpdateWithOtherMsg(t *testing.T) {
	model := tuiModel{screen: screenMain}
	_, _ = model.Update("some other message")
}

func TestIsTerminalWriter(t *testing.T) {
	result := isTerminalWriter(&strings.Builder{})
	if result != false {
		t.Errorf("isTerminalWriter should return false for non-terminal")
	}
}

func TestCurrentBuildInfo(t *testing.T) {
	info := CurrentBuildInfo()
	if info.Version == "" {
		t.Errorf("Version should not be empty")
	}
	if info.Commit == "" {
		t.Errorf("Commit should not be empty")
	}
	if info.BuildDate == "" {
		t.Errorf("BuildDate should not be empty")
	}
}

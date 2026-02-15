package core

import (
	"bytes"
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	domainprojectinventory "github.com/felixgeelhaar/aios/internal/domain/projectinventory"
	domainskilllint "github.com/felixgeelhaar/aios/internal/domain/skilllint"
	domainskillpackage "github.com/felixgeelhaar/aios/internal/domain/skillpackage"
	domainskillsync "github.com/felixgeelhaar/aios/internal/domain/skillsync"
	domainskilltest "github.com/felixgeelhaar/aios/internal/domain/skilltest"
	domainskilluninstall "github.com/felixgeelhaar/aios/internal/domain/skilluninstall"
	domainworkspace "github.com/felixgeelhaar/aios/internal/domain/workspaceorchestration"
)

func makeTestCLI() CLI {
	cli := CLI{Out: &bytes.Buffer{}}
	cli.ListProjects = func(context.Context) ([]domainprojectinventory.Project, error) { return nil, nil }
	cli.AddProject = func(context.Context, string) (domainprojectinventory.Project, error) {
		return domainprojectinventory.Project{}, nil
	}
	cli.RemoveProject = func(context.Context, string) error { return nil }
	cli.InspectProject = func(context.Context, string) (domainprojectinventory.Project, error) {
		return domainprojectinventory.Project{ID: "test", Path: "/test"}, nil
	}
	cli.ValidateWorkspace = func(context.Context) (domainworkspace.ValidationResult, error) {
		return domainworkspace.ValidationResult{}, nil
	}
	cli.PlanWorkspace = func(context.Context) (domainworkspace.PlanResult, error) { return domainworkspace.PlanResult{}, nil }
	cli.RepairWorkspace = func(context.Context) (domainworkspace.RepairResult, error) {
		return domainworkspace.RepairResult{}, nil
	}
	cli.ListClients = func() map[string]any { return map[string]any{} }
	cli.InitSkill = func(string) error { return nil }
	cli.SyncSkill = func(context.Context, domainskillsync.SyncSkillCommand) (string, error) { return "test-skill", nil }
	cli.TestSkill = func(context.Context, domainskilltest.TestSkillCommand) (domainskilltest.TestSkillResult, error) {
		return domainskilltest.TestSkillResult{}, nil
	}
	cli.LintSkill = func(context.Context, domainskilllint.LintSkillCommand) (domainskilllint.LintSkillResult, error) {
		return domainskilllint.LintSkillResult{Valid: true}, nil
	}
	cli.PackageSkill = func(context.Context, domainskillpackage.PackageSkillCommand) (domainskillpackage.PackageSkillResult, error) {
		return domainskillpackage.PackageSkillResult{ArtifactPath: "/test.tar.gz"}, nil
	}
	cli.UninstallSkill = func(context.Context, domainskilluninstall.UninstallSkillCommand) (string, error) {
		return "test-skill", nil
	}
	cli.MarketplaceList = func(context.Context) (map[string]any, error) {
		return map[string]any{"listings": []map[string]any{}}, nil
	}
	cli.MarketplaceInstall = func(context.Context, string) (map[string]any, error) { return map[string]any{}, nil }
	cli.TrayStatus = func() (TrayState, error) { return TrayState{}, nil }
	return cli
}

func TestTUISkillInitFlow(t *testing.T) {
	cli := makeTestCLI()
	cli.In = strings.NewReader("q\n")
	cli.InitSkill = func(dir string) error {
		return nil
	}

	err := cli.RunTUI(context.Background())
	if err != nil {
		t.Logf("tui exited: %v", err)
	}
}

func TestTUISkillSyncFlow(t *testing.T) {
	cli := makeTestCLI()
	cli.In = strings.NewReader("q\n")

	err := cli.RunTUI(context.Background())
	if err != nil {
		t.Logf("tui exited: %v", err)
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

func TestTUIMainMenuItems(t *testing.T) {
	model := tuiModel{}
	items := model.mainMenuItems()
	if len(items) != 6 {
		t.Errorf("mainMenuItems returned %d items, want 6", len(items))
	}
	if items[0] != "Projects" {
		t.Errorf("first item = %q, want 'Projects'", items[0])
	}
}

func TestTUISkillsMenuItems(t *testing.T) {
	model := tuiModel{}
	items := model.skillsMenuItems()
	if len(items) != 8 {
		t.Errorf("skillsMenuItems returned %d items, want 8", len(items))
	}
	if items[0] != "List skills" {
		t.Errorf("first item = %q, want 'List skills'", items[0])
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

func TestTUIViewMainScreen(t *testing.T) {
	model := tuiModel{screen: screenMain, cursor: 0}
	view := model.View()
	if !strings.Contains(view, "AIOS") {
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
}

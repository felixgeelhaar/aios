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

func TestTUIProjectsMenuItems(t *testing.T) {
	model := tuiModel{}
	items := model.projectsMenuItems()
	if len(items) != 3 {
		t.Errorf("projectsMenuItems returned %d items, want 3", len(items))
	}
	if items[0] != "List projects" {
		t.Errorf("first item = %q, want 'List projects'", items[0])
	}
}

func TestTUIMarketplaceMenuItems(t *testing.T) {
	model := tuiModel{}
	items := model.marketplaceMenuItems()
	if len(items) != 2 {
		t.Errorf("marketplaceMenuItems returned %d items, want 2", len(items))
	}
	if items[0] != "List skills" {
		t.Errorf("first item = %q, want 'List skills'", items[0])
	}
}

func TestTUIWorkspaceMenuItems(t *testing.T) {
	model := tuiModel{}
	items := model.workspaceMenuItems()
	if len(items) != 3 {
		t.Errorf("workspaceMenuItems returned %d items, want 3", len(items))
	}
	if items[0] != "Validate" {
		t.Errorf("first item = %q, want 'Validate'", items[0])
	}
}

func TestTUIViewProjectsScreen(t *testing.T) {
	model := tuiModel{screen: screenProjects, cursor: 0}
	view := model.View()
	if !strings.Contains(view, "Projects") {
		t.Errorf("expected Projects in view, got %q", view)
	}
}

func TestTUIViewProjectAddScreen(t *testing.T) {
	model := tuiModel{screen: screenProjectAdd, input: "/path/to/project"}
	view := model.View()
	if !strings.Contains(view, "Add Project") {
		t.Errorf("expected Add Project in view, got %q", view)
	}
}

func TestTUIViewProjectRemoveScreen(t *testing.T) {
	model := tuiModel{screen: screenProjectRemove, input: "my-project"}
	view := model.View()
	if !strings.Contains(view, "Remove Project") {
		t.Errorf("expected Remove Project in view, got %q", view)
	}
}

func TestTUIViewSkillSyncScreen(t *testing.T) {
	model := tuiModel{screen: screenSkillSync, input: "my-skill"}
	view := model.View()
	if !strings.Contains(view, "Sync Skill") {
		t.Errorf("expected Sync Skill in view, got %q", view)
	}
}

func TestTUIViewSkillTestScreen(t *testing.T) {
	model := tuiModel{screen: screenSkillTest, input: "my-skill"}
	view := model.View()
	if !strings.Contains(view, "Test Skill") {
		t.Errorf("expected Test Skill in view, got %q", view)
	}
}

func TestTUIViewSkillLintScreen(t *testing.T) {
	model := tuiModel{screen: screenSkillLint, input: "my-skill"}
	view := model.View()
	if !strings.Contains(view, "Lint Skill") {
		t.Errorf("expected Lint Skill in view, got %q", view)
	}
}

func TestTUIViewSkillPackageScreen(t *testing.T) {
	model := tuiModel{screen: screenSkillPackage, input: "my-skill"}
	view := model.View()
	if !strings.Contains(view, "Package Skill") {
		t.Errorf("expected Package Skill in view, got %q", view)
	}
}

func TestTUIViewSkillUninstallScreen(t *testing.T) {
	model := tuiModel{screen: screenSkillUninstall, input: "my-skill"}
	view := model.View()
	if !strings.Contains(view, "Uninstall Skill") {
		t.Errorf("expected Uninstall Skill in view, got %q", view)
	}
}

func TestTUIViewSkillInspectScreen(t *testing.T) {
	model := tuiModel{screen: screenSkillInspect, input: "my-skill"}
	view := model.View()
	if !strings.Contains(view, "Inspect Skill") {
		t.Errorf("expected Inspect Skill in view, got %q", view)
	}
}

func TestTUIViewMarketplaceScreen(t *testing.T) {
	model := tuiModel{screen: screenMarketplace, cursor: 0}
	view := model.View()
	if !strings.Contains(view, "Marketplace") {
		t.Errorf("expected Marketplace in view, got %q", view)
	}
}

func TestTUIViewMarketplaceInstallScreen(t *testing.T) {
	model := tuiModel{screen: screenMarketplaceInstall, input: "skill-id"}
	view := model.View()
	if !strings.Contains(view, "Install Skill") {
		t.Errorf("expected Install Skill in view, got %q", view)
	}
}

func TestTUIViewConnectorsScreen(t *testing.T) {
	cli := makeTestCLI()
	cli.TrayStatus = func() (TrayState, error) {
		return TrayState{Connections: map[string]bool{"google_drive": true}}, nil
	}
	model := tuiModel{screen: screenConnectors, cursor: 0, cli: cli}
	view := model.View()
	if !strings.Contains(view, "Connectors") {
		t.Errorf("expected Connectors in view, got %q", view)
	}
}

func TestTUIViewWorkspaceScreen(t *testing.T) {
	model := tuiModel{screen: screenWorkspace, cursor: 0}
	view := model.View()
	if !strings.Contains(view, "Workspace") {
		t.Errorf("expected Workspace in view, got %q", view)
	}
}

func TestTUIViewSettingsScreen(t *testing.T) {
	model := tuiModel{screen: screenSettings, cursor: 0}
	view := model.View()
	if !strings.Contains(view, "Settings") {
		t.Errorf("expected Settings in view, got %q", view)
	}
}

func TestTUIHandleMainMenu(t *testing.T) {
	model := tuiModel{}
	result, _ := model.handleMainMenu(0)
	resultModel := result.(tuiModel)
	if resultModel.screen != screenProjects {
		t.Errorf("expected screenProjects, got %v", resultModel.screen)
	}

	model = tuiModel{}
	result, _ = model.handleMainMenu(1)
	resultModel = result.(tuiModel)
	if resultModel.screen != screenSkills {
		t.Errorf("expected screenSkills, got %v", resultModel.screen)
	}

	model = tuiModel{}
	result, _ = model.handleMainMenu(2)
	resultModel = result.(tuiModel)
	if resultModel.screen != screenMarketplace {
		t.Errorf("expected screenMarketplace, got %v", resultModel.screen)
	}

	model = tuiModel{}
	result, _ = model.handleMainMenu(3)
	resultModel = result.(tuiModel)
	if resultModel.screen != screenConnectors {
		t.Errorf("expected screenConnectors, got %v", resultModel.screen)
	}

	model = tuiModel{}
	result, _ = model.handleMainMenu(4)
	resultModel = result.(tuiModel)
	if resultModel.screen != screenWorkspace {
		t.Errorf("expected screenWorkspace, got %v", resultModel.screen)
	}

	model = tuiModel{}
	result, _ = model.handleMainMenu(5)
	resultModel = result.(tuiModel)
	if resultModel.screen != screenSettings {
		t.Errorf("expected screenSettings, got %v", resultModel.screen)
	}
}

func TestTUIHandleProjectsMenu(t *testing.T) {
	cli := makeTestCLI()
	ctx := context.Background()
	model := tuiModel{cli: cli, ctx: ctx}

	result, _ := model.handleProjectsMenu(3)
	resultModel := result.(tuiModel)
	if resultModel.screen != screenMain {
		t.Errorf("expected screenMain, got %v", resultModel.screen)
	}
}

func TestTUIHandleSkillsMenu(t *testing.T) {
	model := tuiModel{}
	cli := makeTestCLI()
	model.cli = cli

	_, _ = model.handleSkillsMenu(8)
	if model.screen != screenMain {
		t.Errorf("expected screenMain, got %v", model.screen)
	}
}

func TestTUIHandleMarketplaceMenu(t *testing.T) {
	cli := makeTestCLI()
	cli.MarketplaceList = func(context.Context) (map[string]any, error) {
		return map[string]any{
			"listings": []map[string]any{
				{"skill_id": "test-skill", "versions": "1.0.0"},
			},
		}, nil
	}
	model := tuiModel{cli: cli, ctx: context.Background()}

	result, _ := model.handleMarketplaceMenu(0)
	resultModel := result.(tuiModel)
	if resultModel.status != "info" {
		t.Errorf("expected info status, got %q", resultModel.status)
	}

	result, _ = model.handleMarketplaceMenu(2)
	resultModel = result.(tuiModel)
	if resultModel.screen != screenMain {
		t.Errorf("expected screenMain, got %v", resultModel.screen)
	}
}

func TestTUIHandleWorkspaceMenu(t *testing.T) {
	cli := makeTestCLI()
	cli.ValidateWorkspace = func(context.Context) (domainworkspace.ValidationResult, error) {
		return domainworkspace.ValidationResult{Healthy: false}, nil
	}
	cli.PlanWorkspace = func(context.Context) (domainworkspace.PlanResult, error) {
		return domainworkspace.PlanResult{Actions: []domainworkspace.PlanAction{{Kind: domainworkspace.ActionCreate}}}, nil
	}
	cli.RepairWorkspace = func(context.Context) (domainworkspace.RepairResult, error) {
		return domainworkspace.RepairResult{Applied: []domainworkspace.PlanAction{{Kind: domainworkspace.ActionCreate}}, Skipped: []domainworkspace.PlanAction{}}, nil
	}
	model := tuiModel{cli: cli, ctx: context.Background()}

	result, _ := model.handleWorkspaceMenu(0)
	resultModel := result.(tuiModel)
	if resultModel.status != "info" {
		t.Errorf("expected info status, got %q", resultModel.status)
	}

	result, _ = model.handleWorkspaceMenu(1)
	resultModel = result.(tuiModel)
	if resultModel.status != "info" {
		t.Errorf("expected info status for plan, got %q", resultModel.status)
	}

	result, _ = model.handleWorkspaceMenu(2)
	resultModel = result.(tuiModel)
	if resultModel.status != "success" {
		t.Errorf("expected success status for repair, got %q", resultModel.status)
	}

	result, _ = model.handleWorkspaceMenu(3)
	resultModel = result.(tuiModel)
	if resultModel.screen != screenMain {
		t.Errorf("expected screenMain, got %v", resultModel.screen)
	}
}

func TestTUIHandleSkillOperation(t *testing.T) {
	cli := makeTestCLI()
	ctx := context.Background()

	model := tuiModel{screen: screenSkillInit, input: "test-skill", cli: cli, ctx: ctx}
	result, _ := model.handleSkillOperation("enter")
	resultModel := result.(tuiModel)
	if resultModel.status != "success" {
		t.Errorf("expected success status, got %q", resultModel.status)
	}

	model = tuiModel{screen: screenSkillInit, input: "", cli: cli, ctx: ctx}
	result, _ = model.handleSkillOperation("enter")
	resultModel = result.(tuiModel)
	if resultModel.status != "error" {
		t.Errorf("expected error status for empty input, got %q", resultModel.status)
	}

	model = tuiModel{screen: screenSkillInit, cli: cli, ctx: ctx}
	result, _ = model.handleSkillOperation("a")
	resultModel = result.(tuiModel)
	if resultModel.input != "a" {
		t.Errorf("expected input 'a', got %q", resultModel.input)
	}

	model = tuiModel{screen: screenSkillInit, input: "test", cli: cli, ctx: ctx}
	result, _ = model.handleSkillOperation("backspace")
	resultModel = result.(tuiModel)
	if resultModel.input != "tes" {
		t.Errorf("expected input 'tes', got %q", resultModel.input)
	}

	model = tuiModel{screen: screenSkillInit, cli: cli, ctx: ctx}
	result, _ = model.handleSkillOperation("esc")
	resultModel = result.(tuiModel)
	if resultModel.screen != screenSkills {
		t.Errorf("expected screenSkills, got %v", resultModel.screen)
	}

	model = tuiModel{screen: screenSkillSync, input: "test-skill", cli: cli, ctx: ctx}
	result, _ = model.handleSkillOperation("enter")
	resultModel = result.(tuiModel)
	if resultModel.status != "success" {
		t.Errorf("expected success status for sync, got %q", resultModel.status)
	}

	model = tuiModel{screen: screenSkillTest, input: "test-skill", cli: cli, ctx: ctx}
	result, _ = model.handleSkillOperation("enter")
	resultModel = result.(tuiModel)
	if resultModel.status != "success" {
		t.Errorf("expected success status for test, got %q", resultModel.status)
	}

	model = tuiModel{screen: screenSkillLint, input: "test-skill", cli: cli, ctx: ctx}
	result, _ = model.handleSkillOperation("enter")
	resultModel = result.(tuiModel)
	if resultModel.status != "success" {
		t.Errorf("expected success status for lint, got %q", resultModel.status)
	}

	model = tuiModel{screen: screenSkillPackage, input: "test-skill", cli: cli, ctx: ctx}
	result, _ = model.handleSkillOperation("enter")
	resultModel = result.(tuiModel)
	if resultModel.status != "success" {
		t.Errorf("expected success status for package, got %q", resultModel.status)
	}

	model = tuiModel{screen: screenSkillUninstall, input: "test-skill", cli: cli, ctx: ctx}
	result, _ = model.handleSkillOperation("enter")
	resultModel = result.(tuiModel)
	if resultModel.status != "success" {
		t.Errorf("expected success status for uninstall, got %q", resultModel.status)
	}
}

func TestTUIHandleInputScreen(t *testing.T) {
	model := tuiModel{}

	_, _ = model.handleInputScreen("esc", "test operation", func(string) error { return nil }, func() {})

	model = tuiModel{input: "test"}
	result, _ := model.handleInputScreen("enter", "test", func(string) error { return nil }, func() {})
	resultModel := result.(tuiModel)
	if resultModel.status != "success" {
		t.Errorf("expected success status, got %q", resultModel.status)
	}

	model = tuiModel{input: "test"}
	result, _ = model.handleInputScreen("enter", "test", func(string) error { return nil }, func() {})
	resultModel = result.(tuiModel)
	if resultModel.input != "" {
		t.Errorf("expected cleared input, got %q", resultModel.input)
	}

	model = tuiModel{}
	result, _ = model.handleInputScreen("a", "test", func(string) error { return nil }, func() {})
	resultModel = result.(tuiModel)
	if resultModel.input != "a" {
		t.Errorf("expected input 'a', got %q", resultModel.input)
	}
}

func TestTUIHandleInputScreenError(t *testing.T) {
	model := tuiModel{input: "test"}
	result, _ := model.handleInputScreen("enter", "test operation", func(string) error { return assertError("test") }, func() {})
	resultModel := result.(tuiModel)
	if resultModel.status != "error" {
		t.Errorf("expected error status, got %q", resultModel.status)
	}
}

func TestTUIHandleInputScreenEmpty(t *testing.T) {
	model := tuiModel{}
	result, _ := model.handleInputScreen("enter", "test operation", func(string) error { return nil }, func() {})
	resultModel := result.(tuiModel)
	if resultModel.status != "error" {
		t.Errorf("expected error status for empty input, got %q", resultModel.status)
	}
}

func TestTUIHandleGenericBack(t *testing.T) {
	result, _ := tuiModel{}.handleGenericBack("b", screenSkills)
	resultModel := result.(tuiModel)
	if resultModel.screen != screenSkills {
		t.Errorf("expected screenSkills, got %v", resultModel.screen)
	}

	result, _ = tuiModel{}.handleGenericBack("back", screenSkills)
	resultModel = result.(tuiModel)
	if resultModel.screen != screenSkills {
		t.Errorf("expected screenSkills for back, got %v", resultModel.screen)
	}

	result, _ = tuiModel{}.handleGenericBack("esc", screenSkills)
	resultModel = result.(tuiModel)
	if resultModel.screen != screenSkills {
		t.Errorf("expected screenSkills for esc, got %v", resultModel.screen)
	}

	result, _ = tuiModel{}.handleGenericBack("unknown", screenSkills)
	resultModel = result.(tuiModel)
	if resultModel.screen != 0 {
		t.Errorf("expected unchanged screen for unknown key, got %v", resultModel.screen)
	}
}

func TestTUIUpdateMenuBack(t *testing.T) {
	model := tuiModel{screen: screenSkills}
	newModel, _ := model.updateMenu("b", []string{"a", "b"}, func(int) (tea.Model, tea.Cmd) {
		return model, nil
	})
	if newModel.(tuiModel).screen != screenMain {
		t.Errorf("expected screenMain after back, got %v", newModel.(tuiModel).screen)
	}
}

func TestTUIViewSkillListScreen(t *testing.T) {
	model := tuiModel{screen: screenSkillList, message: "skill1\nskill2"}
	view := model.View()
	if !strings.Contains(view, "Installed Skills") {
		t.Errorf("expected Installed Skills in view, got %q", view)
	}
}

func TestTUIUpdateProjectsScreen(t *testing.T) {
	cli := makeTestCLI()
	model := tuiModel{screen: screenProjects, cli: cli}
	result, _ := model.Update("enter")
	_ = result.(tuiModel)
}

func TestTUIUpdateSkillsScreen(t *testing.T) {
	cli := makeTestCLI()
	model := tuiModel{screen: screenSkills, cli: cli}
	result, _ := model.Update("enter")
	_ = result.(tuiModel)
}

func TestTUIUpdateMarketplaceScreen(t *testing.T) {
	cli := makeTestCLI()
	model := tuiModel{screen: screenMarketplace, cli: cli}
	result, _ := model.Update("enter")
	_ = result.(tuiModel)
}

func TestTUIUpdateWorkspaceScreen(t *testing.T) {
	cli := makeTestCLI()
	model := tuiModel{screen: screenWorkspace, cli: cli}
	result, _ := model.Update("enter")
	_ = result.(tuiModel)
}

func TestTUIUpdateSettingsScreen(t *testing.T) {
	model := tuiModel{screen: screenSettings}
	result, _ := model.Update("b")
	_ = result.(tuiModel)
}

func TestTUIUpdateConnectorsScreen(t *testing.T) {
	cli := makeTestCLI()
	cli.TrayStatus = func() (TrayState, error) {
		return TrayState{Connections: map[string]bool{"gdrive": true}}, nil
	}
	model := tuiModel{screen: screenConnectors, cli: cli}
	result, _ := model.Update("b")
	_ = result.(tuiModel)
}

func TestTUIUpdateSkillListScreen(t *testing.T) {
	cli := makeTestCLI()
	model := tuiModel{screen: screenSkillList, cli: cli}
	result, _ := model.Update("b")
	resultModel := result.(tuiModel)
	if resultModel.screen != screenSkillList {
		t.Errorf("expected unchanged screen (string not handled), got %v", resultModel.screen)
	}
}

func TestTUIUpdateWorkspacePlanScreen(t *testing.T) {
	cli := makeTestCLI()
	model := tuiModel{screen: screenWorkspacePlan, cli: cli}
	result, _ := model.Update("b")
	resultModel := result.(tuiModel)
	if resultModel.screen != screenWorkspacePlan {
		t.Errorf("expected unchanged screen (string not handled), got %v", resultModel.screen)
	}
}

func TestTUIUpdateWorkspaceRepairScreen(t *testing.T) {
	cli := makeTestCLI()
	model := tuiModel{screen: screenWorkspaceRepair, cli: cli}
	result, _ := model.Update("b")
	resultModel := result.(tuiModel)
	if resultModel.screen != screenWorkspaceRepair {
		t.Errorf("expected unchanged screen (string not handled), got %v", resultModel.screen)
	}
}

func TestTUIUpdateScreenWithOtherKey(t *testing.T) {
	cli := makeTestCLI()
	model := tuiModel{screen: screenMain, cli: cli}
	result, _ := model.Update("x")
	_ = result
}

func TestIsTerminalReader(t *testing.T) {
	result := isTerminalReader(strings.NewReader("test"))
	if result != false {
		t.Errorf("isTerminalReader should return false for non-terminal")
	}
}

func assertError(msg string) error {
	return &testError{msg: msg}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

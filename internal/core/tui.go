package core

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	domainskilllint "github.com/felixgeelhaar/aios/internal/domain/skilllint"
	domainskillpackage "github.com/felixgeelhaar/aios/internal/domain/skillpackage"
	domainskillsync "github.com/felixgeelhaar/aios/internal/domain/skillsync"
	domainskilltest "github.com/felixgeelhaar/aios/internal/domain/skilltest"
	domainskilluninstall "github.com/felixgeelhaar/aios/internal/domain/skilluninstall"
	"golang.org/x/term"
)

var errTUIQuit = errors.New("tui quit")

type tuiScreen int

const (
	screenMain tuiScreen = iota
	screenProjects
	screenProjectAdd
	screenProjectRemove
	screenSkills
	screenSkillList
	screenSkillInspect
	screenSkillInit
	screenSkillSync
	screenSkillTest
	screenSkillLint
	screenSkillPackage
	screenSkillUninstall
	screenMarketplace
	screenMarketplaceInstall
	screenConnectors
	screenWorkspace
	screenWorkspacePlan
	screenWorkspaceRepair
	screenSettings
)

type tuiModel struct {
	ctx     context.Context
	cli     CLI
	screen  tuiScreen
	cursor  int
	input   string
	message string
	status  string
	data    interface{}
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch m.screen {
		case screenMain:
			return m.updateMenu(key, m.mainMenuItems(), m.handleMainMenu)
		case screenProjects:
			return m.updateMenu(key, m.projectsMenuItems(), m.handleProjectsMenu)
		case screenProjectAdd:
			return m.handleInputScreen(key, "add project", func(input string) error {
				_, err := m.cli.AddProject(m.ctx, input)
				return err
			}, func() { m.screen = screenProjects })
		case screenProjectRemove:
			return m.handleInputScreen(key, "remove project", func(input string) error {
				return m.cli.RemoveProject(m.ctx, input)
			}, func() { m.screen = screenProjects })
		case screenSkills:
			return m.updateMenu(key, m.skillsMenuItems(), m.handleSkillsMenu)
		case screenSkillList:
			return m.handleGenericBack(key, screenSkills)
		case screenSkillInspect:
			return m.handleInputScreen(key, "inspect skill", func(input string) error {
				project, err := m.cli.InspectProject(m.ctx, input)
				if err != nil {
					return err
				}
				m.data = fmt.Sprintf("ID: %s\nPath: %s\nAdded: %s", project.ID, project.Path, project.AddedAt)
				return nil
			}, func() { m.screen = screenSkills })
		case screenSkillInit, screenSkillSync, screenSkillTest, screenSkillLint, screenSkillPackage, screenSkillUninstall:
			return m.handleSkillOperation(key)
		case screenMarketplace:
			return m.updateMenu(key, m.marketplaceMenuItems(), m.handleMarketplaceMenu)
		case screenMarketplaceInstall:
			return m.handleInputScreen(key, "install skill", func(input string) error {
				_, err := m.cli.MarketplaceInstall(m.ctx, input)
				return err
			}, func() { m.screen = screenMarketplace })
		case screenConnectors:
			return m.handleGenericBack(key, screenMain)
		case screenWorkspace:
			return m.updateMenu(key, m.workspaceMenuItems(), m.handleWorkspaceMenu)
		case screenWorkspacePlan:
			return m.handleGenericBack(key, screenWorkspace)
		case screenWorkspaceRepair:
			return m.handleGenericBack(key, screenWorkspace)
		case screenSettings:
			return m.handleGenericBack(key, screenMain)
		}
	}
	return m, nil
}

func (m tuiModel) handleMainMenu(idx int) (tea.Model, tea.Cmd) {
	switch idx {
	case 0:
		m.screen = screenProjects
	case 1:
		m.screen = screenSkills
	case 2:
		m.screen = screenMarketplace
	case 3:
		m.screen = screenConnectors
	case 4:
		m.screen = screenWorkspace
	case 5:
		m.screen = screenSettings
	case 6:
		return m, tea.Quit
	}
	m.cursor = 0
	return m, nil
}

func (m tuiModel) handleProjectsMenu(idx int) (tea.Model, tea.Cmd) {
	switch idx {
	case 0:
		projects, err := m.cli.ListProjects(m.ctx)
		if err != nil {
			m.status = "error"
			m.message = err.Error()
			return m, nil
		}
		if len(projects) == 0 {
			m.status = "info"
			m.message = "no tracked projects"
		} else {
			lines := make([]string, len(projects))
			for i, p := range projects {
				lines[i] = fmt.Sprintf("%s: %s", p.ID, p.Path)
			}
			m.status = "info"
			m.message = strings.Join(lines, "\n")
		}
	case 1:
		m.screen = screenProjectAdd
		m.input = ""
		m.message = ""
	case 2:
		m.screen = screenProjectRemove
		m.input = ""
		m.message = ""
	case 3:
		m.screen = screenMain
		m.cursor = 0
	}
	return m, nil
}

func (m tuiModel) handleSkillsMenu(idx int) (tea.Model, tea.Cmd) {
	switch idx {
	case 0:
		clients := m.cli.ListClients()
		lines := []string{}
		for name, data := range clients {
			if m, ok := data.(map[string]interface{}); ok {
				installed, _ := m["installed"].(bool)
				if installed {
					if skills, ok := m["skills"].([]string); ok {
						lines = append(lines, fmt.Sprintf("%s: %v", name, skills))
					}
				}
			}
		}
		if len(lines) == 0 {
			m.status = "info"
			m.message = "no skills installed"
		} else {
			m.status = "info"
			m.message = strings.Join(lines, "\n")
		}
	case 1:
		m.screen = screenSkillInspect
		m.input = ""
		m.message = ""
	case 2:
		m.screen = screenSkillInit
		m.input = ""
		m.message = ""
	case 3:
		m.screen = screenSkillSync
		m.input = ""
		m.message = ""
	case 4:
		m.screen = screenSkillTest
		m.input = ""
		m.message = ""
	case 5:
		m.screen = screenSkillLint
		m.input = ""
		m.message = ""
	case 6:
		m.screen = screenSkillPackage
		m.input = ""
		m.message = ""
	case 7:
		m.screen = screenSkillUninstall
		m.input = ""
		m.message = ""
	case 8:
		m.screen = screenMain
		m.cursor = 0
	}
	return m, nil
}

func (m tuiModel) handleSkillOperation(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+c", "esc":
		m.screen = screenSkills
		m.input = ""
		m.message = ""
		return m, nil
	case "enter":
		input := strings.TrimSpace(m.input)
		if input == "" {
			m.status = "error"
			m.message = "skill directory is required"
			return m, nil
		}
		var err error
		switch m.screen {
		case screenSkillInit:
			err = m.cli.InitSkill(input)
			if err == nil {
				m.status = "success"
				m.message = fmt.Sprintf("skill created: %s", filepath.Base(input))
			}
		case screenSkillSync:
			skillID, err := m.cli.SyncSkill(m.ctx, domainskillsync.SyncSkillCommand{SkillDir: input})
			if err == nil {
				m.status = "success"
				m.message = fmt.Sprintf("synced: %s", skillID)
			}
		case screenSkillTest:
			result, err := m.cli.TestSkill(m.ctx, domainskilltest.TestSkillCommand{SkillDir: input})
			if err == nil {
				if result.Failed > 0 {
					m.status = "error"
					m.message = fmt.Sprintf("%d tests failed", result.Failed)
				} else {
					m.status = "success"
					m.message = fmt.Sprintf("all %d tests passed", len(result.Results))
				}
			}
		case screenSkillLint:
			result, err := m.cli.LintSkill(m.ctx, domainskilllint.LintSkillCommand{SkillDir: input})
			if err == nil {
				if result.Valid {
					m.status = "success"
					m.message = "lint: ok"
				} else {
					m.status = "error"
					m.message = fmt.Sprintf("lint: %d issues", len(result.Issues))
				}
			}
		case screenSkillPackage:
			result, err := m.cli.PackageSkill(m.ctx, domainskillpackage.PackageSkillCommand{SkillDir: input})
			if err == nil {
				m.status = "success"
				m.message = fmt.Sprintf("packaged: %s", filepath.Base(result.ArtifactPath))
			}
		case screenSkillUninstall:
			skillID, err := m.cli.UninstallSkill(m.ctx, domainskilluninstall.UninstallSkillCommand{SkillDir: input})
			if err == nil {
				m.status = "success"
				m.message = fmt.Sprintf("uninstalled: %s", skillID)
			}
		}
		if err != nil {
			m.status = "error"
			m.message = err.Error()
		}
		m.input = ""
		return m, nil
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil
	default:
		if len(key) == 1 {
			m.input += key
			return m, nil
		}
	}
	return m, nil
}

func (m tuiModel) handleMarketplaceMenu(idx int) (tea.Model, tea.Cmd) {
	switch idx {
	case 0:
		result, err := m.cli.MarketplaceList(m.ctx)
		if err != nil {
			m.status = "error"
			m.message = err.Error()
			return m, nil
		}
		if listings, ok := result["listings"].([]map[string]interface{}); ok && len(listings) > 0 {
			lines := []string{}
			for _, l := range listings {
				lines = append(lines, fmt.Sprintf("%s (v%s)", l["skill_id"], l["versions"]))
			}
			m.status = "info"
			m.message = strings.Join(lines, "\n")
		} else {
			m.status = "info"
			m.message = "no marketplace skills"
		}
	case 1:
		m.screen = screenMarketplaceInstall
		m.input = ""
		m.message = ""
	case 2:
		m.screen = screenMain
		m.cursor = 0
	}
	return m, nil
}

func (m tuiModel) handleWorkspaceMenu(idx int) (tea.Model, tea.Cmd) {
	switch idx {
	case 0:
		result, err := m.cli.ValidateWorkspace(m.ctx)
		if err != nil {
			m.status = "error"
			m.message = err.Error()
			return m, nil
		}
		state := "healthy"
		if !result.Healthy {
			state = "issues found"
		}
		m.status = "info"
		m.message = fmt.Sprintf("workspace: %s\n%d links", state, len(result.Links))
	case 1:
		result, err := m.cli.PlanWorkspace(m.ctx)
		if err != nil {
			m.status = "error"
			m.message = err.Error()
			return m, nil
		}
		if len(result.Actions) == 0 {
			m.status = "info"
			m.message = "no actions needed"
		} else {
			lines := []string{}
			for _, a := range result.Actions {
				lines = append(lines, fmt.Sprintf("%s: %s -> %s", a.Kind, a.LinkPath, a.TargetPath))
			}
			m.status = "info"
			m.message = strings.Join(lines, "\n")
		}
	case 2:
		result, err := m.cli.RepairWorkspace(m.ctx)
		if err != nil {
			m.status = "error"
			m.message = err.Error()
			return m, nil
		}
		m.status = "success"
		m.message = fmt.Sprintf("applied: %d, skipped: %d", len(result.Applied), len(result.Skipped))
	case 3:
		m.screen = screenMain
		m.cursor = 0
	}
	return m, nil
}

func (m tuiModel) handleInputScreen(key string, operation string, handler func(string) error, onSuccess func()) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+c", "esc":
		onSuccess()
		return m, nil
	case "enter":
		input := strings.TrimSpace(m.input)
		if input == "" {
			m.status = "error"
			m.message = fmt.Sprintf("%s: input required", operation)
			return m, nil
		}
		if err := handler(input); err != nil {
			m.status = "error"
			m.message = err.Error()
		} else {
			m.status = "success"
			m.message = fmt.Sprintf("%s: done", operation)
		}
		m.input = ""
		return m, nil
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil
	default:
		if len(key) == 1 {
			m.input += key
			return m, nil
		}
	}
	return m, nil
}

func (m tuiModel) handleGenericBack(key string, backScreen tuiScreen) (tea.Model, tea.Cmd) {
	switch key {
	case "b", "back", "esc":
		m.screen = backScreen
		m.cursor = 0
		return m, nil
	}
	return m, nil
}

func (m tuiModel) View() string {
	var b strings.Builder
	styleHeader := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	styleSelected := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	styleMuted := lipgloss.NewStyle().Faint(true)
	styleSuccess := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82"))
	styleError := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("204"))
	styleInput := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	styleSubtle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	b.WriteString(styleHeader.Render("AIOS"))
	b.WriteString("\n\n")

	switch m.screen {
	case screenMain:
		m.renderMenu(&b, m.mainMenuItems(), styleSelected)
	case screenProjects:
		b.WriteString(styleHeader.Render("Projects"))
		b.WriteString("\n")
		m.renderMenu(&b, m.projectsMenuItems(), styleSelected)
	case screenProjectAdd, screenProjectRemove:
		title := "Add Project"
		if m.screen == screenProjectRemove {
			title = "Remove Project"
		}
		b.WriteString(styleHeader.Render(title))
		b.WriteString("\n")
		b.WriteString(styleSubtle.Render("path or id: "))
		b.WriteString(styleInput.Render(m.input))
		b.WriteString("\n")
		b.WriteString(styleMuted.Render("enter to confirm, esc to cancel"))
	case screenSkills:
		b.WriteString(styleHeader.Render("Skills"))
		b.WriteString("\n")
		m.renderMenu(&b, m.skillsMenuItems(), styleSelected)
	case screenSkillList:
		b.WriteString(styleHeader.Render("Installed Skills"))
		b.WriteString("\n")
		b.WriteString(m.message)
		b.WriteString("\n")
		b.WriteString(styleMuted.Render("b) back"))
	case screenSkillInspect:
		b.WriteString(styleHeader.Render("Inspect Skill"))
		b.WriteString("\n")
		b.WriteString(styleSubtle.Render("skill id or path: "))
		b.WriteString(styleInput.Render(m.input))
		b.WriteString("\n")
		b.WriteString(styleMuted.Render("enter to confirm, esc to cancel"))
	case screenSkillInit:
		b.WriteString(styleHeader.Render("Create Skill"))
		b.WriteString("\n")
		b.WriteString(styleSubtle.Render("skill directory: "))
		b.WriteString(styleInput.Render(m.input))
		b.WriteString("\n")
		b.WriteString(styleMuted.Render("enter to confirm, esc to cancel"))
	case screenSkillSync:
		b.WriteString(styleHeader.Render("Sync Skill"))
		b.WriteString("\n")
		b.WriteString(styleSubtle.Render("skill directory: "))
		b.WriteString(styleInput.Render(m.input))
		b.WriteString("\n")
		b.WriteString(styleMuted.Render("enter to confirm, esc to cancel"))
	case screenSkillTest:
		b.WriteString(styleHeader.Render("Test Skill"))
		b.WriteString("\n")
		b.WriteString(styleSubtle.Render("skill directory: "))
		b.WriteString(styleInput.Render(m.input))
		b.WriteString("\n")
		b.WriteString(styleMuted.Render("enter to confirm, esc to cancel"))
	case screenSkillLint:
		b.WriteString(styleHeader.Render("Lint Skill"))
		b.WriteString("\n")
		b.WriteString(styleSubtle.Render("skill directory: "))
		b.WriteString(styleInput.Render(m.input))
		b.WriteString("\n")
		b.WriteString(styleMuted.Render("enter to confirm, esc to cancel"))
	case screenSkillPackage:
		b.WriteString(styleHeader.Render("Package Skill"))
		b.WriteString("\n")
		b.WriteString(styleSubtle.Render("skill directory: "))
		b.WriteString(styleInput.Render(m.input))
		b.WriteString("\n")
		b.WriteString(styleMuted.Render("enter to confirm, esc to cancel"))
	case screenSkillUninstall:
		b.WriteString(styleHeader.Render("Uninstall Skill"))
		b.WriteString("\n")
		b.WriteString(styleSubtle.Render("skill directory: "))
		b.WriteString(styleInput.Render(m.input))
		b.WriteString("\n")
		b.WriteString(styleMuted.Render("enter to confirm, esc to cancel"))
	case screenMarketplace:
		b.WriteString(styleHeader.Render("Marketplace"))
		b.WriteString("\n")
		m.renderMenu(&b, m.marketplaceMenuItems(), styleSelected)
	case screenMarketplaceInstall:
		b.WriteString(styleHeader.Render("Install Skill"))
		b.WriteString("\n")
		b.WriteString(styleSubtle.Render("skill id: "))
		b.WriteString(styleInput.Render(m.input))
		b.WriteString("\n")
		b.WriteString(styleMuted.Render("enter to confirm, esc to cancel"))
	case screenConnectors:
		b.WriteString(styleHeader.Render("Connectors"))
		b.WriteString("\n")
		state, _ := m.cli.TrayStatus()
		lines := []string{}
		for name, connected := range state.Connections {
			status := "disconnected"
			if connected {
				status = "connected"
			}
			lines = append(lines, fmt.Sprintf("%s: %s", name, status))
		}
		if len(lines) == 0 {
			lines = []string{"no connectors configured"}
		}
		m.status = "info"
		m.message = strings.Join(lines, "\n")
		b.WriteString(m.message)
		b.WriteString("\n\n")
		b.WriteString(styleMuted.Render("b) back"))
	case screenWorkspace:
		b.WriteString(styleHeader.Render("Workspace"))
		b.WriteString("\n")
		m.renderMenu(&b, m.workspaceMenuItems(), styleSelected)
	case screenSettings:
		b.WriteString(styleHeader.Render("Settings"))
		b.WriteString("\n")
		b.WriteString("AIOS v0.1.0\n")
		b.WriteString(styleMuted.Render("b) back"))
	}

	if m.message != "" {
		b.WriteString("\n\n")
		switch m.status {
		case "success":
			b.WriteString(styleSuccess.Render(m.message))
		case "error":
			b.WriteString(styleError.Render(m.message))
		default:
			b.WriteString(m.message)
		}
	}
	return b.String()
}

func (m tuiModel) renderMenu(b *strings.Builder, items []string, selected lipgloss.Style) {
	for i, item := range items {
		cursor := " "
		line := fmt.Sprintf("%d) %s", i+1, item)
		if i == m.cursor {
			cursor = ">"
			line = selected.Render(line)
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, line))
	}
	b.WriteString("\nb) Back\nq) Quit\n")
}

func (m tuiModel) updateMenu(key string, items []string, onSelect func(int) (tea.Model, tea.Cmd)) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+c", "q", "quit", "exit":
		return m, tea.Quit
	case "b", "back":
		m.screen = screenMain
		m.cursor = 0
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case "down", "j":
		if m.cursor < len(items)-1 {
			m.cursor++
		}
		return m, nil
	case "enter":
		return onSelect(m.cursor)
	default:
		if idx := keyToIndex(key); idx >= 0 && idx < len(items) {
			m.cursor = idx
			return onSelect(m.cursor)
		}
	}
	return m, nil
}

func (m tuiModel) mainMenuItems() []string {
	return []string{
		"Projects",
		"Skills",
		"Marketplace",
		"Connectors",
		"Workspace",
		"Settings",
	}
}

func (m tuiModel) projectsMenuItems() []string {
	return []string{
		"List projects",
		"Add project",
		"Remove project",
	}
}

func (m tuiModel) skillsMenuItems() []string {
	return []string{
		"List skills",
		"Inspect skill",
		"Create skill",
		"Sync skill",
		"Test skill",
		"Lint skill",
		"Package skill",
		"Uninstall skill",
	}
}

func (m tuiModel) marketplaceMenuItems() []string {
	return []string{
		"List skills",
		"Install skill",
	}
}

func (m tuiModel) workspaceMenuItems() []string {
	return []string{
		"Validate",
		"Plan repairs",
		"Repair",
	}
}

func keyToIndex(key string) int {
	if len(key) != 1 {
		return -1
	}
	if key[0] < '1' || key[0] > '9' {
		return -1
	}
	return int(key[0] - '1')
}

func (c CLI) RunTUI(ctx context.Context) error {
	if !isTerminalReader(c.In) || !isTerminalWriter(c.Out) {
		return runTUIScript(ctx, c)
	}
	model := tuiModel{
		ctx:    ctx,
		cli:    c,
		screen: screenMain,
	}
	options := []tea.ProgramOption{tea.WithInput(c.In), tea.WithOutput(c.Out)}
	if !isTerminalWriter(c.Out) {
		options = append(options, tea.WithoutRenderer())
	}
	program := tea.NewProgram(model, options...)
	_, err := program.Run()
	return err
}

func runTUIScript(ctx context.Context, c CLI) error {
	reader := bufio.NewReader(c.In)
	for {
		_, _ = fmt.Fprintln(c.Out, "AIOS Operations Console")
		_, _ = fmt.Fprintln(c.Out, "1) Projects")
		_, _ = fmt.Fprintln(c.Out, "2) Skills")
		_, _ = fmt.Fprintln(c.Out, "3) Marketplace")
		_, _ = fmt.Fprintln(c.Out, "4) Connectors")
		_, _ = fmt.Fprintln(c.Out, "5) Workspace")
		_, _ = fmt.Fprintln(c.Out, "6) Settings")
		_, _ = fmt.Fprintln(c.Out, "q) Quit")
		_, _ = fmt.Fprint(c.Out, "> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		choice := strings.TrimSpace(strings.ToLower(line))
		switch choice {
		case "1":
			if err := runProjectsScript(ctx, c, reader); err != nil {
				return err
			}
		case "2":
			if err := runSkillsScript(ctx, c, reader); err != nil {
				return err
			}
		case "3":
			if err := runMarketplaceScript(ctx, c, reader); err != nil {
				return err
			}
		case "4":
			state, _ := c.TrayStatus()
			for name, connected := range state.Connections {
				_, _ = fmt.Fprintf(c.Out, "%s: %t\n", name, connected)
			}
		case "5":
			if err := runWorkspaceScript(ctx, c, reader); err != nil {
				return err
			}
		case "6":
			b := c.BuildInfo()
			_, _ = fmt.Fprintf(c.Out, "AIOS v%s\n", b.Version)
		case "q", "quit", "exit":
			return nil
		default:
			_, _ = fmt.Fprintln(c.Out, "unknown choice")
		}
	}
}

func runProjectsScript(ctx context.Context, c CLI, reader *bufio.Reader) error {
	for {
		_, _ = fmt.Fprintln(c.Out, "\nProjects")
		_, _ = fmt.Fprintln(c.Out, "1) List projects")
		_, _ = fmt.Fprintln(c.Out, "2) Add project")
		_, _ = fmt.Fprintln(c.Out, "3) Remove project")
		_, _ = fmt.Fprintln(c.Out, "b) Back")
		_, _ = fmt.Fprint(c.Out, "> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		choice := strings.TrimSpace(strings.ToLower(line))
		switch choice {
		case "1":
			projects, err := c.ListProjects(ctx)
			if err != nil {
				_, _ = fmt.Fprintln(c.Out, "error:", err)
				continue
			}
			if len(projects) == 0 {
				_, _ = fmt.Fprintln(c.Out, "no tracked projects")
			}
			for _, p := range projects {
				_, _ = fmt.Fprintf(c.Out, "- %s: %s\n", p.ID, p.Path)
			}
		case "2":
			_, _ = fmt.Print("project path: ")
			pathLine, _ := reader.ReadString('\n')
			path := strings.TrimSpace(pathLine)
			project, err := c.AddProject(ctx, path)
			if err != nil {
				_, _ = fmt.Fprintln(c.Out, "error:", err)
			} else {
				_, _ = fmt.Fprintf(c.Out, "added: %s\n", project.ID)
			}
		case "3":
			_, _ = fmt.Print("project id or path: ")
			idLine, _ := reader.ReadString('\n')
			id := strings.TrimSpace(idLine)
			if err := c.RemoveProject(ctx, id); err != nil {
				_, _ = fmt.Fprintln(c.Out, "error:", err)
			} else {
				_, _ = fmt.Fprintln(c.Out, "removed")
			}
		case "b", "back":
			return nil
		}
	}
}

func runMarketplaceScript(ctx context.Context, c CLI, reader *bufio.Reader) error {
	for {
		_, _ = fmt.Fprintln(c.Out, "\nMarketplace")
		_, _ = fmt.Fprintln(c.Out, "1) List skills")
		_, _ = fmt.Fprintln(c.Out, "2) Install skill")
		_, _ = fmt.Fprintln(c.Out, "b) Back")
		_, _ = fmt.Fprint(c.Out, "> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		choice := strings.TrimSpace(strings.ToLower(line))
		switch choice {
		case "1":
			result, err := c.MarketplaceList(ctx)
			if err != nil {
				_, _ = fmt.Fprintln(c.Out, "error:", err)
				continue
			}
			if listings, ok := result["listings"].([]map[string]interface{}); ok {
				for _, l := range listings {
					_, _ = fmt.Fprintf(c.Out, "- %s\n", l["skill_id"])
				}
			}
		case "2":
			_, _ = fmt.Print("skill id: ")
			idLine, _ := reader.ReadString('\n')
			id := strings.TrimSpace(idLine)
			if _, err := c.MarketplaceInstall(ctx, id); err != nil {
				_, _ = fmt.Fprintln(c.Out, "error:", err)
			} else {
				_, _ = fmt.Fprintln(c.Out, "installed")
			}
		case "b", "back":
			return nil
		}
	}
}

func runWorkspaceScript(ctx context.Context, c CLI, reader *bufio.Reader) error {
	for {
		_, _ = fmt.Fprintln(c.Out, "\nWorkspace")
		_, _ = fmt.Fprintln(c.Out, "1) Validate")
		_, _ = fmt.Fprintln(c.Out, "2) Plan repairs")
		_, _ = fmt.Fprintln(c.Out, "3) Repair")
		_, _ = fmt.Fprintln(c.Out, "b) Back")
		_, _ = fmt.Fprint(c.Out, "> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		choice := strings.TrimSpace(strings.ToLower(line))
		switch choice {
		case "1":
			result, err := c.ValidateWorkspace(ctx)
			if err != nil {
				_, _ = fmt.Fprintln(c.Out, "error:", err)
				continue
			}
			state := "healthy"
			if !result.Healthy {
				state = "issues found"
			}
			_, _ = fmt.Fprintf(c.Out, "workspace: %s\n", state)
		case "2":
			result, err := c.PlanWorkspace(ctx)
			if err != nil {
				_, _ = fmt.Fprintln(c.Out, "error:", err)
				continue
			}
			for _, a := range result.Actions {
				_, _ = fmt.Fprintf(c.Out, "- %s %s -> %s\n", a.Kind, a.LinkPath, a.TargetPath)
			}
		case "3":
			result, err := c.RepairWorkspace(ctx)
			if err != nil {
				_, _ = fmt.Fprintln(c.Out, "error:", err)
				continue
			}
			_, _ = fmt.Fprintf(c.Out, "applied: %d, skipped: %d\n", len(result.Applied), len(result.Skipped))
		case "b", "back":
			return nil
		}
	}
}

func runSkillsScript(ctx context.Context, c CLI, reader *bufio.Reader) error {
	for {
		_, _ = fmt.Fprintln(c.Out, "\nSkills")
		_, _ = fmt.Fprintln(c.Out, "1) List skills")
		_, _ = fmt.Fprintln(c.Out, "2) Create skill")
		_, _ = fmt.Fprintln(c.Out, "3) Sync skill")
		_, _ = fmt.Fprintln(c.Out, "4) Test skill")
		_, _ = fmt.Fprintln(c.Out, "5) Lint skill")
		_, _ = fmt.Fprintln(c.Out, "6) Package skill")
		_, _ = fmt.Fprintln(c.Out, "7) Uninstall skill")
		_, _ = fmt.Fprintln(c.Out, "b) Back")
		_, _ = fmt.Fprint(c.Out, "> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		choice := strings.TrimSpace(strings.ToLower(line))
		switch choice {
		case "1":
			clients := c.ListClients()
			for name, data := range clients {
				if m, ok := data.(map[string]interface{}); ok {
					if installed, _ := m["installed"].(bool); installed {
						_, _ = fmt.Fprintf(c.Out, "- %s\n", name)
					}
				}
			}
		case "2":
			_, _ = fmt.Print("skill directory: ")
			dirLine, _ := reader.ReadString('\n')
			dir := strings.TrimSpace(dirLine)
			if err := c.InitSkill(dir); err != nil {
				_, _ = fmt.Fprintln(c.Out, "error:", err)
			} else {
				_, _ = fmt.Fprintln(c.Out, "skill created")
			}
		case "3":
			_, _ = fmt.Print("skill directory: ")
			dirLine, _ := reader.ReadString('\n')
			dir := strings.TrimSpace(dirLine)
			skillID, err := c.SyncSkill(ctx, domainskillsync.SyncSkillCommand{SkillDir: dir})
			if err != nil {
				_, _ = fmt.Fprintln(c.Out, "error:", err)
			} else {
				_, _ = fmt.Fprintf(c.Out, "synced: %s\n", skillID)
			}
		case "4":
			_, _ = fmt.Print("skill directory: ")
			dirLine, _ := reader.ReadString('\n')
			dir := strings.TrimSpace(dirLine)
			result, err := c.TestSkill(ctx, domainskilltest.TestSkillCommand{SkillDir: dir})
			if err != nil {
				_, _ = fmt.Fprintln(c.Out, "error:", err)
			} else {
				_, _ = fmt.Fprintf(c.Out, "tests: %d passed, %d failed\n", len(result.Results)-result.Failed, result.Failed)
			}
		case "5":
			_, _ = fmt.Print("skill directory: ")
			dirLine, _ := reader.ReadString('\n')
			dir := strings.TrimSpace(dirLine)
			result, err := c.LintSkill(ctx, domainskilllint.LintSkillCommand{SkillDir: dir})
			if err != nil {
				_, _ = fmt.Fprintln(c.Out, "error:", err)
			} else if result.Valid {
				_, _ = fmt.Fprintln(c.Out, "lint: ok")
			} else {
				_, _ = fmt.Fprintf(c.Out, "lint: %d issues\n", len(result.Issues))
			}
		case "6":
			_, _ = fmt.Print("skill directory: ")
			dirLine, _ := reader.ReadString('\n')
			dir := strings.TrimSpace(dirLine)
			result, err := c.PackageSkill(ctx, domainskillpackage.PackageSkillCommand{SkillDir: dir})
			if err != nil {
				_, _ = fmt.Fprintln(c.Out, "error:", err)
			} else {
				_, _ = fmt.Fprintf(c.Out, "packaged: %s\n", result.ArtifactPath)
			}
		case "7":
			_, _ = fmt.Print("skill directory: ")
			dirLine, _ := reader.ReadString('\n')
			dir := strings.TrimSpace(dirLine)
			skillID, err := c.UninstallSkill(ctx, domainskilluninstall.UninstallSkillCommand{SkillDir: dir})
			if err != nil {
				_, _ = fmt.Fprintln(c.Out, "error:", err)
			} else {
				_, _ = fmt.Fprintf(c.Out, "uninstalled: %s\n", skillID)
			}
		case "b", "back":
			return nil
		}
	}
}

func isTerminalReader(r io.Reader) bool {
	file, ok := r.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}

func isTerminalWriter(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}

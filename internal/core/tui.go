package core

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	domainskillsync "github.com/felixgeelhaar/aios/internal/domain/skillsync"
	"golang.org/x/term"
)

var errTUIQuit = errors.New("tui quit")

type tuiScreen int

const (
	screenMain tuiScreen = iota
	screenSkills
	screenSkillInit
	screenSkillSync
)

type tuiModel struct {
	ctx     context.Context
	cli     CLI
	screen  tuiScreen
	cursor  int
	input   string
	message string
	status  string
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
			return m.updateMenu(key, m.mainMenuItems(), func(idx int) (tea.Model, tea.Cmd) {
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
						return m, nil
					}
					lines := make([]string, 0, len(projects))
					for _, p := range projects {
						lines = append(lines, fmt.Sprintf("- %s %s", p.ID, p.Path))
					}
					m.status = "info"
					m.message = strings.Join(lines, "\n")
					return m, nil
				case 1:
					result, err := m.cli.ValidateWorkspace(m.ctx)
					if err != nil {
						m.status = "error"
						m.message = err.Error()
						return m, nil
					}
					state := "healthy"
					if !result.Healthy {
						state = "issues_found"
					}
					m.status = "info"
					m.message = fmt.Sprintf("workspace links: %s", state)
					return m, nil
				case 2:
					result, err := m.cli.RepairWorkspace(m.ctx)
					if err != nil {
						m.status = "error"
						m.message = err.Error()
						return m, nil
					}
					m.status = "info"
					m.message = fmt.Sprintf("applied: %d skipped: %d", len(result.Applied), len(result.Skipped))
					return m, nil
				case 3:
					m.screen = screenSkills
					m.cursor = 0
					return m, nil
				case 4:
					return m, tea.Quit
				}
				return m, nil
			})
		case screenSkills:
			return m.updateMenu(key, m.skillsMenuItems(), func(idx int) (tea.Model, tea.Cmd) {
				switch idx {
				case 0:
					m.screen = screenSkillInit
					m.input = ""
					m.status = "info"
					m.message = ""
					return m, nil
				case 1:
					m.screen = screenSkillSync
					m.input = ""
					m.status = "info"
					m.message = ""
					return m, nil
				case 2:
					m.screen = screenMain
					m.cursor = 0
					return m, nil
				}
				return m, nil
			})
		case screenSkillInit, screenSkillSync:
			switch key {
			case "ctrl+c", "esc":
				m.screen = screenSkills
				m.input = ""
				return m, nil
			case "enter":
				input := strings.TrimSpace(m.input)
				if input == "" {
					m.status = "error"
					m.message = "skill directory is required"
					return m, nil
				}
				if m.screen == screenSkillInit {
					if err := m.cli.InitSkill(input); err != nil {
						m.status = "error"
						m.message = err.Error()
						return m, nil
					}
					m.status = "success"
					m.message = "skill scaffold created"
				} else {
					skillID, err := m.cli.SyncSkill(m.ctx, domainskillsync.SyncSkillCommand{SkillDir: input})
					if err != nil {
						m.status = "error"
						m.message = err.Error()
						return m, nil
					}
					m.status = "success"
					m.message = fmt.Sprintf("sync completed for skill %s", skillID)
				}
				m.screen = screenSkills
				m.input = ""
				return m, nil
			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
				return m, nil
			default:
				if msg.Type == tea.KeyRunes {
					m.input += string(msg.Runes)
					return m, nil
				}
			}
		}
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

	b.WriteString(styleHeader.Render("AIOS Operations Console"))
	b.WriteString("\n\n")

	switch m.screen {
	case screenMain:
		m.renderMenu(&b, m.mainMenuItems(), styleSelected)
	case screenSkills:
		b.WriteString(styleHeader.Render("Skills"))
		b.WriteString("\n")
		m.renderMenu(&b, m.skillsMenuItems(), styleSelected)
	case screenSkillInit:
		b.WriteString(styleHeader.Render("Init skill"))
		b.WriteString("\n")
		b.WriteString("skill directory: ")
		b.WriteString(styleInput.Render(m.input))
		b.WriteString("\n")
		b.WriteString(styleMuted.Render("enter to confirm, esc to cancel"))
	case screenSkillSync:
		b.WriteString(styleHeader.Render("Sync skill"))
		b.WriteString("\n")
		b.WriteString("skill directory: ")
		b.WriteString(styleInput.Render(m.input))
		b.WriteString("\n")
		b.WriteString(styleMuted.Render("enter to confirm, esc to cancel"))
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
	b.WriteString("\nq) Quit\n")
}

func (m tuiModel) updateMenu(key string, items []string, onSelect func(int) (tea.Model, tea.Cmd)) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+c", "q", "quit", "exit":
		return m, tea.Quit
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
		if isNumericKey(key) {
			m.status = "error"
			m.message = "unknown choice"
		}
	}
	return m, nil
}

func (m tuiModel) mainMenuItems() []string {
	return []string{"Projects", "Workspace Validate", "Workspace Repair", "Skills", "Quit"}
}

func (m tuiModel) skillsMenuItems() []string {
	return []string{"Init skill", "Sync skill", "Back"}
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

func isNumericKey(key string) bool {
	return keyToIndex(key) >= 0
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
		_, _ = fmt.Fprintln(c.Out, "2) Workspace Validate")
		_, _ = fmt.Fprintln(c.Out, "3) Workspace Repair")
		_, _ = fmt.Fprintln(c.Out, "4) Skills")
		_, _ = fmt.Fprintln(c.Out, "q) Quit")
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
				return err
			}
			if len(projects) == 0 {
				_, _ = fmt.Fprintln(c.Out, "no tracked projects")
				continue
			}
			for _, p := range projects {
				_, _ = fmt.Fprintf(c.Out, "- %s %s\n", p.ID, p.Path)
			}
		case "2":
			result, err := c.ValidateWorkspace(ctx)
			if err != nil {
				return err
			}
			state := "healthy"
			if !result.Healthy {
				state = "issues_found"
			}
			_, _ = fmt.Fprintf(c.Out, "workspace links: %s\n", state)
		case "3":
			result, err := c.RepairWorkspace(ctx)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(c.Out, "applied: %d skipped: %d\n", len(result.Applied), len(result.Skipped))
		case "4":
			if err := runSkillsScript(ctx, c, reader); err != nil {
				if errors.Is(err, errTUIQuit) {
					return nil
				}
				return err
			}
		case "q", "quit", "exit":
			return nil
		default:
			_, _ = fmt.Fprintln(c.Out, "unknown choice")
		}
	}
}

func runSkillsScript(ctx context.Context, c CLI, reader *bufio.Reader) error {
	for {
		_, _ = fmt.Fprintln(c.Out, "Skills")
		_, _ = fmt.Fprintln(c.Out, "1) Init skill")
		_, _ = fmt.Fprintln(c.Out, "2) Sync skill")
		_, _ = fmt.Fprintln(c.Out, "b) Back")
		_, _ = fmt.Fprint(c.Out, "> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		choice := strings.TrimSpace(strings.ToLower(line))
		switch choice {
		case "1":
			_, _ = fmt.Fprint(c.Out, "skill directory: ")
			skillDirLine, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			skillDir := strings.TrimSpace(skillDirLine)
			if err := c.InitSkill(skillDir); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(c.Out, "skill scaffold created")
		case "2":
			_, _ = fmt.Fprint(c.Out, "skill directory: ")
			skillDirLine, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			skillDir := strings.TrimSpace(skillDirLine)
			skillID, err := c.SyncSkill(ctx, domainskillsync.SyncSkillCommand{SkillDir: skillDir})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(c.Out, "sync completed for skill %s\n", skillID)
		case "b", "back":
			return nil
		case "q", "quit", "exit":
			return errTUIQuit
		default:
			_, _ = fmt.Fprintln(c.Out, "unknown choice")
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

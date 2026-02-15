package core

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	domainprojectinventory "github.com/felixgeelhaar/aios/internal/domain/projectinventory"
	domainskillsync "github.com/felixgeelhaar/aios/internal/domain/skillsync"
	domainworkspace "github.com/felixgeelhaar/aios/internal/domain/workspaceorchestration"
)

func TestTUIUpdateMainProjectsEmpty(t *testing.T) {
	model := tuiModel{
		ctx:    context.Background(),
		cli:    CLI{ListProjects: func(context.Context) ([]domainprojectinventory.Project, error) { return nil, nil }},
		screen: screenMain,
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	if got := updated.(tuiModel).message; got != "no tracked projects" {
		t.Fatalf("expected no tracked projects message, got %q", got)
	}
}

func TestTUIUpdateMainProjectsError(t *testing.T) {
	model := tuiModel{
		ctx: context.Background(),
		cli: CLI{ListProjects: func(context.Context) ([]domainprojectinventory.Project, error) {
			return nil, errors.New("boom")
		}},
		screen: screenMain,
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	if updated.(tuiModel).status != "error" {
		t.Fatalf("expected error status")
	}
}

func TestTUIUpdateMainProjectsList(t *testing.T) {
	model := tuiModel{
		ctx: context.Background(),
		cli: CLI{ListProjects: func(context.Context) ([]domainprojectinventory.Project, error) {
			return []domainprojectinventory.Project{{ID: "p1", Path: "/tmp"}}, nil
		}},
		screen: screenMain,
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	if updated.(tuiModel).status != "info" {
		t.Fatalf("expected info status")
	}
}

func TestTUIUpdateMainValidateWorkspace(t *testing.T) {
	model := tuiModel{
		ctx: context.Background(),
		cli: CLI{ValidateWorkspace: func(context.Context) (domainworkspace.ValidationResult, error) {
			return domainworkspace.ValidationResult{Healthy: false}, nil
		}},
		screen: screenMain,
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	if updated.(tuiModel).message == "" {
		t.Fatalf("expected message for workspace validation")
	}
}

func TestTUIUpdateMainRepairWorkspace(t *testing.T) {
	model := tuiModel{
		ctx: context.Background(),
		cli: CLI{RepairWorkspace: func(context.Context) (domainworkspace.RepairResult, error) {
			return domainworkspace.RepairResult{Applied: []domainworkspace.PlanAction{{}}, Skipped: []domainworkspace.PlanAction{{}}}, nil
		}},
		screen: screenMain,
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	if updated.(tuiModel).message == "" {
		t.Fatalf("expected repair message")
	}
}

func TestTUIUpdateMainToSkills(t *testing.T) {
	model := tuiModel{screen: screenMain}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	if updated.(tuiModel).screen != screenSkills {
		t.Fatalf("expected skills screen")
	}
}

func TestTUIUpdateSkillsNavigation(t *testing.T) {
	model := tuiModel{screen: screenSkills}

	initScreen, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	if initScreen.(tuiModel).screen != screenSkillInit {
		t.Fatalf("expected init screen")
	}

	syncScreen, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	if syncScreen.(tuiModel).screen != screenSkillSync {
		t.Fatalf("expected sync screen")
	}

	backScreen, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	if backScreen.(tuiModel).screen != screenMain {
		t.Fatalf("expected main screen")
	}
}

func TestTUIUpdateSkillInitFlow(t *testing.T) {
	model := tuiModel{
		ctx:    context.Background(),
		cli:    CLI{InitSkill: func(string) error { return nil }},
		screen: screenSkillInit,
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if updated.(tuiModel).status != "error" {
		t.Fatalf("expected error on empty input")
	}

	model.input = "skill"
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if updated.(tuiModel).status != "success" {
		t.Fatalf("expected success on init")
	}
}

func TestTUIUpdateSkillInitRunesAndBackspace(t *testing.T) {
	model := tuiModel{screen: screenSkillInit}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	updated, _ = updated.(tuiModel).Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if updated.(tuiModel).input != "" {
		t.Fatalf("expected input cleared")
	}
}

func TestTUIUpdateSkillSyncFlow(t *testing.T) {
	model := tuiModel{
		ctx: context.Background(),
		cli: CLI{SyncSkill: func(context.Context, domainskillsync.SyncSkillCommand) (string, error) {
			return "skill-id", nil
		}},
		screen: screenSkillSync,
		input:  "skill",
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if updated.(tuiModel).status != "success" {
		t.Fatalf("expected success on sync")
	}
}

func TestTUIUpdateSkillCancel(t *testing.T) {
	model := tuiModel{screen: screenSkillSync, input: "skill"}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if updated.(tuiModel).screen != screenSkills {
		t.Fatalf("expected skills screen")
	}
}

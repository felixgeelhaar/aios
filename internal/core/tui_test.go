package core

import (
	"bytes"
	"context"
	"strings"
	"testing"

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

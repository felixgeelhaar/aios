package core

import (
	"bufio"
	"context"
	"fmt"
	"strings"
)

func (c CLI) runTUI(ctx context.Context) error {
	reader := bufio.NewReader(c.In)
	for {
		_, _ = fmt.Fprintln(c.Out, "AIOS Operations Console")
		_, _ = fmt.Fprintln(c.Out, "1) Projects")
		_, _ = fmt.Fprintln(c.Out, "2) Workspace Validate")
		_, _ = fmt.Fprintln(c.Out, "3) Workspace Repair")
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
		case "q", "quit", "exit":
			_, _ = fmt.Fprintln(c.Out, "bye")
			return nil
		default:
			_, _ = fmt.Fprintln(c.Out, "unknown choice")
		}
	}
}

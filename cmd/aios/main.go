package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/felixgeelhaar/aios/internal/core"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("aios", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	mode := fs.String("mode", "tray", "run mode: tray or cli")
	command := fs.String("command", "status", "cli command when mode=cli")
	skillDir := fs.String("skill-dir", "", "skill artifact directory used by sync command")
	skillID := fs.String("skill-id", "", "deprecated: use --skill-dir")
	mcpTransport := fs.String("mcp-transport", "stdio", "MCP transport for serve-mcp: stdio|http|ws")
	mcpAddr := fs.String("mcp-addr", ":8080", "MCP listen address for http/ws")
	output := fs.String("output", "text", "output format: text|json")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	writeError := func(err error) {
		if *output == "json" {
			body, _ := json.Marshal(map[string]any{
				"error":   err.Error(),
				"command": *command,
			})
			fmt.Fprintln(stderr, string(body))
		} else {
			fmt.Fprintln(stderr, err)
		}
	}

	cfg := core.DefaultConfig()
	app := core.NewApp(cfg)

	if *mode == "cli" {
		cli := core.DefaultCLI(stdout, cfg)
		argSkillDir := *skillDir
		if argSkillDir == "" && *skillID != "" {
			argSkillDir = *skillID
		}
		if err := cli.Run(context.Background(), *command, argSkillDir, *mcpTransport, *mcpAddr, *output); err != nil {
			writeError(err)
			return 1
		}
		return 0
	}

	if err := app.Run(*mode); err != nil {
		writeError(err)
		return 1
	}
	return 0
}

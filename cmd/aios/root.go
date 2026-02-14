package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/felixgeelhaar/aios/internal/core"
	"github.com/spf13/cobra"
)

const (
	defaultMCPAddr      = ":8080"
	defaultMCPTransport = "stdio"
)

type rootOptions struct {
	output       string
	mode         string
	command      string
	skillDir     string
	skillID      string
	mcpTransport string
	mcpAddr      string
}

func newRootCmd(stdout, stderr io.Writer) *cobra.Command {
	opts := &rootOptions{
		output:  "text",
		command: "status",
	}

	root := &cobra.Command{
		Use:           "aios",
		Short:         "AIOS CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.mode != "" || cmd.Flags().Changed("command") {
				return runLegacy(cmd.Context(), stdout, opts)
			}
			return cmd.Help()
		},
	}
	root.SetOut(stdout)
	root.SetErr(stderr)

	root.PersistentFlags().StringVar(&opts.output, "output", "text", "output format: text|json")
	root.PersistentFlags().StringVar(&opts.mode, "mode", "", "DEPRECATED: use subcommands")
	root.PersistentFlags().StringVar(&opts.command, "command", "status", "DEPRECATED: use subcommands")
	root.PersistentFlags().StringVar(&opts.skillDir, "skill-dir", "", "DEPRECATED: use subcommands")
	root.PersistentFlags().StringVar(&opts.skillID, "skill-id", "", "DEPRECATED: use subcommands")
	root.PersistentFlags().StringVar(&opts.mcpTransport, "mcp-transport", defaultMCPTransport, "DEPRECATED: use mcp serve flags")
	root.PersistentFlags().StringVar(&opts.mcpAddr, "mcp-addr", defaultMCPAddr, "DEPRECATED: use mcp serve flags")
	_ = root.PersistentFlags().MarkDeprecated("mode", "use subcommands instead")
	_ = root.PersistentFlags().MarkDeprecated("command", "use subcommands instead")
	_ = root.PersistentFlags().MarkDeprecated("skill-dir", "use subcommands instead")
	_ = root.PersistentFlags().MarkDeprecated("skill-id", "use subcommands instead")
	_ = root.PersistentFlags().MarkDeprecated("mcp-transport", "use mcp serve flags instead")
	_ = root.PersistentFlags().MarkDeprecated("mcp-addr", "use mcp serve flags instead")
	_ = root.PersistentFlags().MarkHidden("mode")
	_ = root.PersistentFlags().MarkHidden("command")
	_ = root.PersistentFlags().MarkHidden("skill-dir")
	_ = root.PersistentFlags().MarkHidden("skill-id")
	_ = root.PersistentFlags().MarkHidden("mcp-transport")
	_ = root.PersistentFlags().MarkHidden("mcp-addr")

	root.AddCommand(newStatusCmd(opts, stdout))
	root.AddCommand(newVersionCmd(opts, stdout))
	root.AddCommand(newDoctorCmd(opts, stdout))
	root.AddCommand(newTrayStatusCmd(opts, stdout))
	root.AddCommand(newListClientsCmd(opts, stdout))
	root.AddCommand(newModelPolicyPacksCmd(opts, stdout))
	root.AddCommand(newAnalyticsCmd(opts, stdout))
	root.AddCommand(newMarketplaceCmd(opts, stdout))
	root.AddCommand(newProjectCmd(opts, stdout))
	root.AddCommand(newWorkspaceCmd(opts, stdout))
	root.AddCommand(newSkillsCmd(opts, stdout))
	root.AddCommand(newAuditCmd(opts, stdout))
	root.AddCommand(newRuntimeCmd(opts, stdout))
	root.AddCommand(newMCPServerCmd(opts, stdout))
	root.AddCommand(newBackupCmd(opts, stdout))
	root.AddCommand(newRestoreCmd(opts, stdout))
	root.AddCommand(newExportStatusReportCmd(opts, stdout))
	root.AddCommand(newConnectGoogleDriveCmd(opts, stdout))
	root.AddCommand(newTuiCmd(opts, stdout))
	root.AddCommand(newTrayCmd(opts))

	return root
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	root := newRootCmd(stdout, stderr)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		output := extractOutputFlag(args)
		writeError(stderr, err, output, args)
		if isFlagParseError(err) {
			return 2
		}
		return 1
	}
	return 0
}

func runLegacy(ctx context.Context, stdout io.Writer, opts *rootOptions) error {
	cfg := core.DefaultConfig()
	app := core.NewApp(cfg)
	if opts.mode == "tray" {
		return app.Run("tray")
	}
	if opts.mode != "cli" {
		return fmt.Errorf("unsupported mode %q", opts.mode)
	}
	cli := core.DefaultCLI(stdout, cfg)
	skillDir := opts.skillDir
	if skillDir == "" && opts.skillID != "" {
		skillDir = opts.skillID
	}
	return cli.Run(ctx, opts.command, skillDir, opts.mcpTransport, opts.mcpAddr, opts.output)
}

func writeError(stderr io.Writer, err error, output string, args []string) {
	if output == "json" {
		cmd := ""
		if len(args) > 0 {
			cmd = args[0]
		}
		body, _ := json.Marshal(map[string]any{
			"error":   err.Error(),
			"command": cmd,
		})
		fmt.Fprintln(stderr, string(body))
		return
	}
	fmt.Fprintln(stderr, err)
}

func extractOutputFlag(args []string) string {
	for i, arg := range args {
		if arg == "--output" && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(arg, "--output=") {
			return strings.TrimPrefix(arg, "--output=")
		}
	}
	return "text"
}

func isFlagParseError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "unknown flag") || strings.Contains(msg, "requires an argument")
}

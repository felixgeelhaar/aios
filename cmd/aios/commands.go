package main

import (
	"context"
	"fmt"
	"io"

	"github.com/felixgeelhaar/aios/internal/core"
	"github.com/spf13/cobra"
)

func runCLI(ctx context.Context, stdout io.Writer, opts *rootOptions, command string, arg string, mcpTransport string, mcpAddr string) error {
	cli := core.DefaultCLI(stdout, core.DefaultConfig())
	return cli.Run(ctx, command, arg, mcpTransport, mcpAddr, opts.output)
}

func argOrFlag(cmd *cobra.Command, args []string, flagName string) string {
	if len(args) > 0 {
		return args[0]
	}
	if flagName == "" {
		return ""
	}
	val, _ := cmd.Flags().GetString(flagName)
	return val
}

func requireArgOrFlag(cmd *cobra.Command, args []string, flagName string, name string) (string, error) {
	val := argOrFlag(cmd, args, flagName)
	if val == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return val, nil
}

func newStatusCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show runtime status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "status", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newVersionCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "version", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newDoctorCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run health checks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "doctor", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newTrayStatusCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "tray-status",
		Short: "Show tray state",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "tray-status", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newListClientsCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "list-clients",
		Short: "List detected agent skill directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "list-clients", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newModelPolicyPacksCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "model-policy-packs",
		Short: "List available model policy packs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "model-policy-packs", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newAnalyticsCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "analytics", Short: "Analytics commands"}

	summary := &cobra.Command{
		Use:   "summary",
		Short: "Show analytics summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "analytics-summary", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
	record := &cobra.Command{
		Use:   "record",
		Short: "Record analytics snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "analytics-record", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
	trend := &cobra.Command{
		Use:   "trend",
		Short: "Show analytics trends",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "analytics-trend", "", defaultMCPTransport, defaultMCPAddr)
		},
	}

	cmd.AddCommand(summary, record, trend)
	return cmd
}

func newMarketplaceCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "marketplace", Short: "Marketplace commands"}

	publish := &cobra.Command{
		Use:   "publish <skill-dir>",
		Short: "Publish a skill",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillDir, err := requireArgOrFlag(cmd, args, "skill-dir", "skill-dir")
			if err != nil {
				return err
			}
			return runCLI(cmd.Context(), stdout, opts, "marketplace-publish", skillDir, defaultMCPTransport, defaultMCPAddr)
		},
	}
	addSkillDirFlag(publish)

	list := &cobra.Command{
		Use:   "list",
		Short: "List marketplace skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "marketplace-list", "", defaultMCPTransport, defaultMCPAddr)
		},
	}

	install := &cobra.Command{
		Use:   "install <skill-id>",
		Short: "Install a marketplace skill",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillID, err := requireArgOrFlag(cmd, args, "skill-id", "skill-id")
			if err != nil {
				return err
			}
			return runCLI(cmd.Context(), stdout, opts, "marketplace-install", skillID, defaultMCPTransport, defaultMCPAddr)
		},
	}
	addSkillIDFlag(install)

	matrix := &cobra.Command{
		Use:   "matrix",
		Short: "Show client compatibility matrix",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "marketplace-matrix", "", defaultMCPTransport, defaultMCPAddr)
		},
	}

	cmd.AddCommand(publish, list, install, matrix)
	return cmd
}

func newProjectCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "project", Short: "Project inventory commands"}

	list := &cobra.Command{
		Use:   "list",
		Short: "List tracked projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "project-list", "", defaultMCPTransport, defaultMCPAddr)
		},
	}

	add := &cobra.Command{
		Use:   "add <path>",
		Short: "Track a project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := requireArgOrFlag(cmd, args, "path", "path")
			if err != nil {
				return err
			}
			return runCLI(cmd.Context(), stdout, opts, "project-add", path, defaultMCPTransport, defaultMCPAddr)
		},
	}
	add.Flags().String("path", "", "project path")

	remove := &cobra.Command{
		Use:   "remove <path-or-id>",
		Short: "Untrack a project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			selector, err := requireArgOrFlag(cmd, args, "selector", "selector")
			if err != nil {
				return err
			}
			return runCLI(cmd.Context(), stdout, opts, "project-remove", selector, defaultMCPTransport, defaultMCPAddr)
		},
	}
	remove.Flags().String("selector", "", "project path or ID")

	inspect := &cobra.Command{
		Use:   "inspect <path-or-id>",
		Short: "Inspect a tracked project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			selector, err := requireArgOrFlag(cmd, args, "selector", "selector")
			if err != nil {
				return err
			}
			return runCLI(cmd.Context(), stdout, opts, "project-inspect", selector, defaultMCPTransport, defaultMCPAddr)
		},
	}
	inspect.Flags().String("selector", "", "project path or ID")

	cmd.AddCommand(list, add, remove, inspect)
	return cmd
}

func newWorkspaceCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "workspace", Short: "Workspace orchestration commands"}

	validate := &cobra.Command{
		Use:   "validate",
		Short: "Validate workspace links",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "workspace-validate", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
	plan := &cobra.Command{
		Use:   "plan",
		Short: "Plan workspace repairs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "workspace-plan", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
	repair := &cobra.Command{
		Use:   "repair",
		Short: "Repair workspace links",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "workspace-repair", "", defaultMCPTransport, defaultMCPAddr)
		},
	}

	cmd.AddCommand(validate, plan, repair)
	return cmd
}

func newSkillsCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "skills", Short: "Skill commands"}

	init := &cobra.Command{
		Use:   "init <skill-dir>",
		Short: "Create a skill scaffold",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillDir, err := requireArgOrFlag(cmd, args, "skill-dir", "skill-dir")
			if err != nil {
				return err
			}
			return runCLI(cmd.Context(), stdout, opts, "init-skill", skillDir, defaultMCPTransport, defaultMCPAddr)
		},
	}
	addSkillDirFlag(init)

	sync := &cobra.Command{
		Use:   "sync <skill-dir>",
		Short: "Sync a skill to agents",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillDir, err := requireArgOrFlag(cmd, args, "skill-dir", "skill-dir")
			if err != nil {
				return err
			}
			return runCLI(cmd.Context(), stdout, opts, "sync", skillDir, defaultMCPTransport, defaultMCPAddr)
		},
	}
	addSkillDirFlag(sync)

	plan := &cobra.Command{
		Use:   "plan <skill-dir>",
		Short: "Plan skill writes",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillDir, err := requireArgOrFlag(cmd, args, "skill-dir", "skill-dir")
			if err != nil {
				return err
			}
			return runCLI(cmd.Context(), stdout, opts, "sync-plan", skillDir, defaultMCPTransport, defaultMCPAddr)
		},
	}
	addSkillDirFlag(plan)

	testCmd := &cobra.Command{
		Use:   "test <skill-dir>",
		Short: "Run skill fixture suite",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillDir, err := requireArgOrFlag(cmd, args, "skill-dir", "skill-dir")
			if err != nil {
				return err
			}
			return runCLI(cmd.Context(), stdout, opts, "test-skill", skillDir, defaultMCPTransport, defaultMCPAddr)
		},
	}
	addSkillDirFlag(testCmd)

	lint := &cobra.Command{
		Use:   "lint <skill-dir>",
		Short: "Lint a skill",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillDir, err := requireArgOrFlag(cmd, args, "skill-dir", "skill-dir")
			if err != nil {
				return err
			}
			return runCLI(cmd.Context(), stdout, opts, "lint-skill", skillDir, defaultMCPTransport, defaultMCPAddr)
		},
	}
	addSkillDirFlag(lint)

	packageCmd := &cobra.Command{
		Use:   "package <skill-dir>",
		Short: "Package a skill",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillDir, err := requireArgOrFlag(cmd, args, "skill-dir", "skill-dir")
			if err != nil {
				return err
			}
			return runCLI(cmd.Context(), stdout, opts, "package-skill", skillDir, defaultMCPTransport, defaultMCPAddr)
		},
	}
	addSkillDirFlag(packageCmd)

	uninstall := &cobra.Command{
		Use:   "uninstall <skill-dir>",
		Short: "Uninstall a skill",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillDir, err := requireArgOrFlag(cmd, args, "skill-dir", "skill-dir")
			if err != nil {
				return err
			}
			return runCLI(cmd.Context(), stdout, opts, "uninstall-skill", skillDir, defaultMCPTransport, defaultMCPAddr)
		},
	}
	addSkillDirFlag(uninstall)

	cmd.AddCommand(init, sync, plan, testCmd, lint, packageCmd, uninstall)
	return cmd
}

func newAuditCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "audit", Short: "Governance audit commands"}

	exportCmd := &cobra.Command{
		Use:   "export [path]",
		Short: "Export audit bundle",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := argOrFlag(cmd, args, "path")
			return runCLI(cmd.Context(), stdout, opts, "audit-export", path, defaultMCPTransport, defaultMCPAddr)
		},
	}
	exportCmd.Flags().String("path", "", "output file path")

	verify := &cobra.Command{
		Use:   "verify <path>",
		Short: "Verify audit bundle",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := requireArgOrFlag(cmd, args, "path", "path")
			if err != nil {
				return err
			}
			return runCLI(cmd.Context(), stdout, opts, "audit-verify", path, defaultMCPTransport, defaultMCPAddr)
		},
	}
	verify.Flags().String("path", "", "input file path")

	cmd.AddCommand(exportCmd, verify)
	return cmd
}

func newRuntimeCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "runtime", Short: "Runtime commands"}

	export := &cobra.Command{
		Use:   "execution-report [path]",
		Short: "Export runtime execution report",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := argOrFlag(cmd, args, "path")
			return runCLI(cmd.Context(), stdout, opts, "runtime-execution-report", path, defaultMCPTransport, defaultMCPAddr)
		},
	}
	export.Flags().String("path", "", "output file path")

	cmd.AddCommand(export)
	return cmd
}

func newMCPServerCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "mcp", Short: "MCP server commands"}

	serve := &cobra.Command{
		Use:   "serve",
		Short: "Serve MCP endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			transport, _ := cmd.Flags().GetString("transport")
			addr, _ := cmd.Flags().GetString("addr")
			return runCLI(cmd.Context(), stdout, opts, "serve-mcp", "", transport, addr)
		},
	}
	serve.Flags().String("transport", defaultMCPTransport, "transport: stdio|http|ws")
	serve.Flags().String("addr", defaultMCPAddr, "listen address for http/ws")

	cmd.AddCommand(serve)
	return cmd
}

func newBackupCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "backup-configs",
		Short: "Backup client configs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "backup-configs", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newRestoreCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore-configs [backup-dir]",
		Short: "Restore client configs",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backupDir := argOrFlag(cmd, args, "backup-dir")
			return runCLI(cmd.Context(), stdout, opts, "restore-configs", backupDir, defaultMCPTransport, defaultMCPAddr)
		},
	}
	cmd.Flags().String("backup-dir", "", "backup directory")
	return cmd
}

func newExportStatusReportCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-status-report [path]",
		Short: "Export status report",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := argOrFlag(cmd, args, "path")
			return runCLI(cmd.Context(), stdout, opts, "export-status-report", path, defaultMCPTransport, defaultMCPAddr)
		},
	}
	cmd.Flags().String("path", "", "output file path")
	return cmd
}

func newConnectGoogleDriveCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "connect-google-drive",
		Short: "Connect Google Drive",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "connect-google-drive", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newTuiCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive console",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "tui", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newTrayCmd(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "tray",
		Short: "Run tray mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := core.NewApp(core.DefaultConfig())
			return app.Run("tray")
		},
	}
}

func addSkillDirFlag(cmd *cobra.Command) {
	cmd.Flags().String("skill-dir", "", "skill directory")
	_ = cmd.Flags().MarkDeprecated("skill-dir", "use positional argument instead")
}

func addSkillIDFlag(cmd *cobra.Command) {
	cmd.Flags().String("skill-id", "", "skill identifier")
	_ = cmd.Flags().MarkDeprecated("skill-id", "use positional argument instead")
}

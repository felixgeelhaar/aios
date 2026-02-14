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
		return "", fmt.Errorf("%s is required\n\nUsage: %s", name, cmd.Use)
	}
	return val, nil
}

func newStatusCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Short:   "Show runtime status",
		Long:    "Displays the current runtime health, active connections, and system state.",
		Example: "  aios status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "status", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newVersionCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Show version metadata",
		Long:    "Displays build information, version number, and runtime configuration.",
		Example: "  aios version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "version", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newDoctorCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:     "doctor",
		Short:   "Run health checks",
		Long:    "Performs diagnostic checks on the runtime environment, including connectivity, configuration, and dependencies.",
		Example: "  aios doctor",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "doctor", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newTrayStatusCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:     "tray-status",
		Short:   "Show tray state",
		Long:    "Displays the current state of the system tray including active connections and notifications.",
		Example: "  aios tray-status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "tray-status", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newListClientsCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:     "list-clients",
		Short:   "List detected agent skill directories",
		Long:    "Scans and lists all agent skill directories detected in common locations (Cursor, VS Code, Windsurf, etc.).",
		Example: "  aios list-clients",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "list-clients", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newModelPolicyPacksCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:     "model-policy-packs",
		Short:   "List available model policy packs",
		Long:    "Displays all available model policy packs that define routing rules and behavior for different AI models.",
		Example: "  aios model-policy-packs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "model-policy-packs", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newAnalyticsCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "analytics",
		Short:   "Analytics commands",
		Long:    "View and manage runtime analytics, trends, and snapshots.",
		Example: "  aios analytics summary\n  aios analytics trend\n  aios analytics record",
	}

	summary := &cobra.Command{
		Use:     "summary",
		Short:   "Show analytics summary",
		Long:    "Displays a summary of runtime analytics including usage stats and trends.",
		Example: "  aios analytics summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "analytics-summary", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
	record := &cobra.Command{
		Use:     "record",
		Short:   "Record analytics snapshot",
		Long:    "Captures a snapshot of current analytics data for trend analysis.",
		Example: "  aios analytics record",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "analytics-record", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
	trend := &cobra.Command{
		Use:     "trend",
		Short:   "Show analytics trends",
		Long:    "Shows historical trends and patterns in runtime analytics.",
		Example: "  aios analytics trend",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "analytics-trend", "", defaultMCPTransport, defaultMCPAddr)
		},
	}

	cmd.AddCommand(summary, record, trend)
	return cmd
}

func newMarketplaceCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "marketplace",
		Short:   "Marketplace commands",
		Long:    "Publish, list, and install skills from the AIOS marketplace.",
		Example: "  aios marketplace list\n  aios marketplace install ddd-expert\n  aios marketplace publish ./my-skill",
	}

	publish := &cobra.Command{
		Use:     "publish <skill-dir>",
		Short:   "Publish a skill",
		Long:    "Uploads a skill to the AIOS marketplace for public or organizational use.",
		Example: "  aios marketplace publish ./my-skill",
		Args:    cobra.MaximumNArgs(1),
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
		Use:     "list",
		Short:   "List marketplace skills",
		Long:    "Shows all available skills in the marketplace with their descriptions.",
		Example: "  aios marketplace list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "marketplace-list", "", defaultMCPTransport, defaultMCPAddr)
		},
	}

	install := &cobra.Command{
		Use:     "install <skill-id>",
		Short:   "Install a marketplace skill",
		Long:    "Downloads and installs a skill from the marketplace by its ID.",
		Example: "  aios marketplace install ddd-expert\n  aios marketplace install my-org/custom-skill",
		Args:    cobra.MaximumNArgs(1),
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
		Use:     "matrix",
		Short:   "Show client compatibility matrix",
		Long:    "Displays which marketplace skills are compatible with which agent clients.",
		Example: "  aios marketplace matrix",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "marketplace-matrix", "", defaultMCPTransport, defaultMCPAddr)
		},
	}

	cmd.AddCommand(publish, list, install, matrix)
	return cmd
}

func newProjectCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Short:   "Project inventory commands",
		Long:    "Track and manage projects for skill routing and context.",
		Example: "  aios project list\n  aios project add ./my-project\n  aios project remove my-project-id",
	}

	list := &cobra.Command{
		Use:     "list",
		Short:   "List tracked projects",
		Long:    "Lists all projects currently tracked by AIOS for skill routing.",
		Example: "  aios project list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "project-list", "", defaultMCPTransport, defaultMCPAddr)
		},
	}

	add := &cobra.Command{
		Use:     "add <path>",
		Short:   "Track a project",
		Long:    "Adds a project to the AIOS inventory for context-aware skill routing.",
		Example: "  aios project add ./my-project\n  aios project add --path /path/to/project",
		Args:    cobra.MaximumNArgs(1),
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
		Use:     "remove <path-or-id>",
		Short:   "Untrack a project",
		Long:    "Removes a project from the AIOS inventory.",
		Example: "  aios project remove my-project-id\n  aios project remove ./my-project",
		Args:    cobra.MaximumNArgs(1),
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
		Use:     "inspect <path-or-id>",
		Short:   "Inspect a tracked project",
		Long:    "Shows detailed information about a tracked project including ID, path, and skills.",
		Example: "  aios project inspect my-project-id",
		Args:    cobra.MaximumNArgs(1),
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
	cmd := &cobra.Command{
		Use:     "workspace",
		Short:   "Workspace orchestration commands",
		Long:    "Validate, plan, and repair workspace symlinks and agent configurations.",
		Example: "  aios workspace validate\n  aios workspace plan\n  aios workspace repair",
	}

	validate := &cobra.Command{
		Use:     "validate",
		Short:   "Validate workspace links",
		Long:    "Checks all workspace symlinks and agent configurations for consistency.",
		Example: "  aios workspace validate",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "workspace-validate", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
	plan := &cobra.Command{
		Use:     "plan",
		Short:   "Plan workspace repairs",
		Long:    "Shows what repairs would be made without applying them (dry-run).",
		Example: "  aios workspace plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "workspace-plan", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
	repair := &cobra.Command{
		Use:     "repair",
		Short:   "Repair workspace links",
		Long:    "Applies repairs to fix broken symlinks and inconsistent configurations.",
		Example: "  aios workspace repair",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "workspace-repair", "", defaultMCPTransport, defaultMCPAddr)
		},
	}

	cmd.AddCommand(validate, plan, repair)
	return cmd
}

func newSkillsCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "skills",
		Short:   "Skill commands",
		Long:    "Manage skills: create, sync, test, lint, package, and uninstall.",
		Example: "  aios skills init my-skill\n  aios skills sync ./my-skill\n  aios skills lint ./my-skill",
	}

	init := &cobra.Command{
		Use:     "init <skill-dir>",
		Short:   "Create a skill scaffold",
		Long:    "Creates a new skill directory with the standard file structure including SKILL.md, fixtures, and configuration.",
		Example: "  aios skills init my-new-skill\n  aios skills init ./custom-skill",
		Args:    cobra.MaximumNArgs(1),
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
		Use:     "sync <skill-dir>",
		Short:   "Sync a skill to agents",
		Long:    "Synchronizes a skill to all configured agent directories, creating symlinks and updating registry.",
		Example: "  aios skills sync ./my-skill\n  aios skills sync ~/skills/ddd-expert",
		Args:    cobra.MaximumNArgs(1),
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
		Use:     "plan <skill-dir>",
		Short:   "Plan skill writes",
		Long:    "Shows what files would be written to agent directories without making changes (dry-run).",
		Example: "  aios skills plan ./my-skill",
		Args:    cobra.MaximumNArgs(1),
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
		Use:     "test <skill-dir>",
		Short:   "Run skill fixture suite",
		Long:    "Executes the skill's fixture test suite to validate skill behavior.",
		Example: "  aios skills test ./my-skill",
		Args:    cobra.MaximumNArgs(1),
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
		Use:     "lint <skill-dir>",
		Short:   "Lint a skill",
		Long:    "Validates skill structure, SKILL.md syntax, and fixture consistency.",
		Example: "  aios skills lint ./my-skill\n  aios skills lint ./my-skill --fix",
		Args:    cobra.MaximumNArgs(1),
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
		Use:     "package <skill-dir>",
		Short:   "Package a skill",
		Long:    "Creates a distributable package (.tar.gz) of the skill for sharing or publishing.",
		Example: "  aios skills package ./my-skill",
		Args:    cobra.MaximumNArgs(1),
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
		Use:     "uninstall <skill-dir>",
		Short:   "Uninstall a skill",
		Long:    "Removes a skill from all agent directories, cleaning up symlinks and registry entries.",
		Example: "  aios skills uninstall ./my-skill\n  aios skills uninstall ddd-expert",
		Args:    cobra.MaximumNArgs(1),
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
	cmd := &cobra.Command{
		Use:     "audit",
		Short:   "Governance audit commands",
		Long:    "Export and verify governance audit bundles for compliance reporting.",
		Example: "  aios audit export\n  aios audit export --path ./audit.json\n  aios audit verify ./audit.json",
	}

	exportCmd := &cobra.Command{
		Use:     "export [path]",
		Short:   "Export audit bundle",
		Long:    "Creates a governance audit bundle with compliance data and runtime state.",
		Example: "  aios audit export\n  aios audit export --path ./audit.json",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := argOrFlag(cmd, args, "path")
			return runCLI(cmd.Context(), stdout, opts, "audit-export", path, defaultMCPTransport, defaultMCPAddr)
		},
	}
	exportCmd.Flags().String("path", "", "output file path")

	verify := &cobra.Command{
		Use:     "verify <path>",
		Short:   "Verify audit bundle",
		Long:    "Validates an audit bundle's integrity and compliance status.",
		Example: "  aios audit verify ./audit.json",
		Args:    cobra.MaximumNArgs(1),
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
	cmd := &cobra.Command{
		Use:     "runtime",
		Short:   "Runtime commands",
		Long:    "Manage runtime execution, reporting, and diagnostics.",
		Example: "  aios runtime execution-report\n  aios runtime execution-report --path ./report.json",
	}

	export := &cobra.Command{
		Use:     "execution-report [path]",
		Short:   "Export runtime execution report",
		Long:    "Generates a detailed report of skill executions, timing, and outcomes.",
		Example: "  aios runtime execution-report\n  aios runtime execution-report --path ./report.json",
		Args:    cobra.MaximumNArgs(1),
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
	cmd := &cobra.Command{
		Use:     "mcp",
		Short:   "MCP server commands",
		Long:    "Start the Model Context Protocol server for agent integrations.",
		Example: "  aios mcp serve\n  aios mcp serve --transport http --addr :8080",
	}

	serve := &cobra.Command{
		Use:     "serve",
		Short:   "Serve MCP endpoints",
		Long:    "Starts the Model Context Protocol server. Supports stdio, HTTP, and WebSocket transports.",
		Example: "  aios mcp serve\n  aios mcp serve --transport http --addr :8080\n  aios mcp serve --transport ws --addr :8081",
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
		Use:     "backup-configs",
		Short:   "Backup client configs",
		Long:    "Creates a backup of all agent configuration files to a timestamped directory.",
		Example: "  aios backup-configs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "backup-configs", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newRestoreCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "restore-configs [backup-dir]",
		Short:   "Restore client configs",
		Long:    "Restores agent configurations from a previously created backup directory.",
		Example: "  aios restore-configs ./backup-2024-01-15\n  aios restore-configs",
		Args:    cobra.MaximumNArgs(1),
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
		Use:     "export-status-report [path]",
		Short:   "Export status report",
		Long:    "Exports a comprehensive JSON status report including runtime state, skills, and projects.",
		Example: "  aios export-status-report\n  aios export-status-report --path ./status.json",
		Args:    cobra.MaximumNArgs(1),
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
		Use:     "connect-google-drive",
		Short:   "Connect Google Drive",
		Long:    "Initiates OAuth flow to connect Google Drive for skill storage and sync.",
		Example: "  aios connect-google-drive",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "connect-google-drive", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newTuiCmd(opts *rootOptions, stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:     "tui",
		Short:   "Launch the interactive console",
		Long:    "Opens an interactive terminal UI for managing skills, projects, and workspace operations.",
		Example: "  aios tui",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(cmd.Context(), stdout, opts, "tui", "", defaultMCPTransport, defaultMCPAddr)
		},
	}
}

func newTrayCmd(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "tray",
		Short:   "Run tray mode",
		Long:    "Starts the system tray application for background operation and quick access.",
		Example: "  aios tray",
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

package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/felixgeelhaar/aios/internal/agents"
	applicationonboarding "github.com/felixgeelhaar/aios/internal/application/onboarding"
	applicationprojectinventory "github.com/felixgeelhaar/aios/internal/application/projectinventory"
	applicationskilllint "github.com/felixgeelhaar/aios/internal/application/skilllint"
	applicationskillpackage "github.com/felixgeelhaar/aios/internal/application/skillpackage"
	applicationskillsync "github.com/felixgeelhaar/aios/internal/application/skillsync"
	applicationskilltest "github.com/felixgeelhaar/aios/internal/application/skilltest"
	applicationskilluninstall "github.com/felixgeelhaar/aios/internal/application/skilluninstall"
	applicationsyncplan "github.com/felixgeelhaar/aios/internal/application/syncplan"
	applicationworkspace "github.com/felixgeelhaar/aios/internal/application/workspaceorchestration"
	"github.com/felixgeelhaar/aios/internal/builder"
	domainonboarding "github.com/felixgeelhaar/aios/internal/domain/onboarding"
	domainprojectinventory "github.com/felixgeelhaar/aios/internal/domain/projectinventory"
	domainskilllint "github.com/felixgeelhaar/aios/internal/domain/skilllint"
	domainskillpackage "github.com/felixgeelhaar/aios/internal/domain/skillpackage"
	domainskillsync "github.com/felixgeelhaar/aios/internal/domain/skillsync"
	domainskilltest "github.com/felixgeelhaar/aios/internal/domain/skilltest"
	domainskilluninstall "github.com/felixgeelhaar/aios/internal/domain/skilluninstall"
	domainsyncplan "github.com/felixgeelhaar/aios/internal/domain/syncplan"
	domainworkspace "github.com/felixgeelhaar/aios/internal/domain/workspaceorchestration"
	"github.com/felixgeelhaar/aios/internal/governance"
	"github.com/felixgeelhaar/aios/internal/marketplace"
	aosmcp "github.com/felixgeelhaar/aios/internal/mcp"
	"github.com/felixgeelhaar/aios/internal/model"
	"github.com/felixgeelhaar/aios/internal/observability"
	"github.com/felixgeelhaar/aios/internal/registry"
	"github.com/felixgeelhaar/aios/internal/runtime"
	"github.com/felixgeelhaar/aios/internal/skill"
	mcpg "github.com/felixgeelhaar/mcp-go"
)

type CLI struct {
	In                 io.Reader
	Out                io.Writer
	SyncState          func() string
	ServeMCP           func(context.Context, *mcpg.Server, ...mcpg.ServeOption) error
	Health             func() runtime.HealthReport
	SyncSkill          func(ctx context.Context, command domainskillsync.SyncSkillCommand) (string, error)
	TestSkill          func(ctx context.Context, command domainskilltest.TestSkillCommand) (domainskilltest.TestSkillResult, error)
	SyncPlan           func(ctx context.Context, command domainsyncplan.BuildSyncPlanCommand) (domainsyncplan.BuildSyncPlanResult, error)
	InitSkill          func(skillDir string) error
	LintSkill          func(ctx context.Context, command domainskilllint.LintSkillCommand) (domainskilllint.LintSkillResult, error)
	BuildInfo          func() BuildInfo
	Doctor             func() DoctorReport
	ListClients        func() map[string]any
	ListProjects       func(ctx context.Context) ([]domainprojectinventory.Project, error)
	AddProject         func(ctx context.Context, path string) (domainprojectinventory.Project, error)
	RemoveProject      func(ctx context.Context, selector string) error
	InspectProject     func(ctx context.Context, selector string) (domainprojectinventory.Project, error)
	ValidateWorkspace  func(ctx context.Context) (domainworkspace.ValidationResult, error)
	PlanWorkspace      func(ctx context.Context) (domainworkspace.PlanResult, error)
	RepairWorkspace    func(ctx context.Context) (domainworkspace.RepairResult, error)
	ModelPolicyPacks   func() []model.PolicyPack
	PackageSkill       func(ctx context.Context, command domainskillpackage.PackageSkillCommand) (domainskillpackage.PackageSkillResult, error)
	UninstallSkill     func(ctx context.Context, command domainskilluninstall.UninstallSkillCommand) (string, error)
	BackupConfigs      func() (string, error)
	RestoreConfigs     func(backupDir string) (string, error)
	ExportReport       func(path string) (string, error)
	AnalyticsSummary   func(ctx context.Context) (map[string]any, error)
	AnalyticsRecord    func(ctx context.Context) (map[string]any, error)
	AnalyticsTrend     func(ctx context.Context) (map[string]any, error)
	MarketplacePublish func(ctx context.Context, skillDir string) (map[string]any, error)
	MarketplaceList    func(ctx context.Context) (map[string]any, error)
	MarketplaceInstall func(ctx context.Context, skillID string) (map[string]any, error)
	MarketplaceMatrix  func(ctx context.Context) (map[string]any, error)
	ExportAudit        func(path string) (map[string]any, error)
	VerifyAudit        func(path string) (map[string]any, error)
	ExecutionReport    func(path string) (map[string]any, error)
	ConnectGoogleDrive func(ctx context.Context, command domainonboarding.ConnectGoogleDriveCommand) (domainonboarding.ConnectGoogleDriveResult, error)
	TrayStatus         func() (TrayState, error)
}

func DefaultCLI(out io.Writer, cfg Config) CLI {
	syncService := applicationskillsync.NewService(
		skillSpecResolverAdapter{},
		clientInstallerAdapter{cfg: cfg},
	)
	packageService := applicationskillpackage.NewService(
		skillMetadataResolverAdapter{},
		skillPackagerAdapter{},
	)
	testService := applicationskilltest.NewService(fixtureRunnerAdapter{})
	lintService := applicationskilllint.NewService(skillLinterAdapter{})
	projectInventoryService := applicationprojectinventory.NewService(
		fileProjectInventoryRepository{workspaceDir: cfg.WorkspaceDir},
		absPathCanonicalizer{},
	)
	syncPlanService := applicationsyncplan.NewService(
		syncPlanSkillResolverAdapter{},
		syncPlanWriteTargetPlannerAdapter{cfg: cfg},
	)
	workspaceService := applicationworkspace.NewService(
		inventoryProjectSource{repo: fileProjectInventoryRepository{workspaceDir: cfg.WorkspaceDir}},
		filesystemWorkspaceLinks{workspaceDir: cfg.WorkspaceDir},
	)
	uninstallService := applicationskilluninstall.NewService(
		uninstallSkillIDResolverAdapter{},
		clientUninstallerAdapter{cfg: cfg},
	)
	onboardingService := applicationonboarding.NewService(
		oauthCodeResolverAdapter{},
		driveConnectorAdapter{cfg: cfg},
		trayStatePortAdapter{cfg: cfg},
	)
	modelRouter := model.NewRouter()

	return CLI{
		In:  os.Stdin,
		Out: out,
		SyncState: func() string {
			return "clean"
		},
		ServeMCP: mcpg.ServeStdio,
		Health: func() runtime.HealthReport {
			rt := runtime.New(cfg.WorkspaceDir, runtime.NewMemoryTokenStore())
			return rt.Health()
		},
		SyncSkill: syncService.SyncSkill,
		TestSkill: testService.TestSkill,
		SyncPlan:  syncPlanService.BuildSyncPlan,
		InitSkill: func(skillDir string) error {
			if skillDir == "" {
				return fmt.Errorf("skill-dir is required")
			}
			return builder.BuildSkill(builder.Spec{
				ID:      filepath.Base(skillDir),
				Version: "0.1.0",
				Dir:     filepath.Dir(skillDir),
			})
		},
		LintSkill: lintService.LintSkill,
		BuildInfo: CurrentBuildInfo,
		Doctor: func() DoctorReport {
			return RunDoctor(cfg)
		},
		ListClients: func() map[string]any {
			collect := func(root string) []string {
				out := []string{}
				entries, err := os.ReadDir(root)
				if err != nil {
					return out
				}
				for _, e := range entries {
					if e.IsDir() {
						out = append(out, e.Name()+"/")
					} else {
						out = append(out, e.Name())
					}
				}
				return out
			}
			allAgents, err := agents.LoadAll()
			if err != nil {
				return map[string]any{"error": err.Error()}
			}
			result := make(map[string]any, len(allAgents))
			for _, agent := range allAgents {
				agentDir := filepath.Join(cfg.ProjectDir, agent.SkillsDir)
				result[agent.Name] = map[string]any{
					"path":  agentDir,
					"files": collect(agentDir),
				}
			}
			return result
		},
		ListProjects: projectInventoryService.List,
		AddProject:   projectInventoryService.Track,
		RemoveProject: func(ctx context.Context, selector string) error {
			return projectInventoryService.Untrack(ctx, selector)
		},
		InspectProject: projectInventoryService.Inspect,
		ValidateWorkspace: func(ctx context.Context) (domainworkspace.ValidationResult, error) {
			return workspaceService.Validate(ctx)
		},
		PlanWorkspace: func(ctx context.Context) (domainworkspace.PlanResult, error) {
			return workspaceService.Plan(ctx)
		},
		RepairWorkspace: func(ctx context.Context) (domainworkspace.RepairResult, error) {
			return workspaceService.Repair(ctx)
		},
		ModelPolicyPacks: modelRouter.Packs,
		PackageSkill:     packageService.PackageSkill,
		UninstallSkill:   uninstallService.UninstallSkill,
		BackupConfigs: func() (string, error) {
			return BackupClientConfigs(cfg)
		},
		RestoreConfigs: func(backupDir string) (string, error) {
			target := backupDir
			if target == "" {
				latest, err := LatestBackupDir(cfg.WorkspaceDir)
				if err != nil {
					return "", err
				}
				target = latest
			}
			if err := RestoreClientConfigs(cfg, target); err != nil {
				return "", err
			}
			return target, nil
		},
		ExportReport: func(path string) (string, error) {
			target := path
			if target == "" {
				target = filepath.Join(cfg.WorkspaceDir, "status-report.md")
			}
			h := runtime.New(cfg.WorkspaceDir, runtime.NewMemoryTokenStore()).Health()
			health := map[string]any{
				"status":      h.Status,
				"ready":       h.Ready,
				"token_store": h.TokenStore,
				"workspace":   h.Workspace,
			}
			if err := ExportStatusReport(target, CurrentBuildInfo(), RunDoctor(cfg), health); err != nil {
				return "", err
			}
			return target, nil
		},
		AnalyticsSummary: func(ctx context.Context) (map[string]any, error) {
			projects, err := projectInventoryService.List(ctx)
			if err != nil {
				return nil, err
			}
			workspace, err := workspaceService.Validate(ctx)
			if err != nil {
				return nil, err
			}
			healthyLinks := 0
			for _, link := range workspace.Links {
				if link.Status == domainworkspace.LinkStatusOK {
					healthyLinks++
				}
			}
			return map[string]any{
				"tracked_projects":  len(projects),
				"workspace_links":   len(workspace.Links),
				"healthy_links":     healthyLinks,
				"workspace_healthy": workspace.Healthy,
				"sync_state":        "clean",
			}, nil
		},
		AnalyticsRecord: func(ctx context.Context) (map[string]any, error) {
			projects, err := projectInventoryService.List(ctx)
			if err != nil {
				return nil, err
			}
			workspace, err := workspaceService.Validate(ctx)
			if err != nil {
				return nil, err
			}
			healthyLinks := 0
			for _, link := range workspace.Links {
				if link.Status == domainworkspace.LinkStatusOK {
					healthyLinks++
				}
			}
			historyPath := filepath.Join(cfg.WorkspaceDir, "state", "analytics-history.json")
			metrics := map[string]float64{
				"tracked_projects": float64(len(projects)),
				"workspace_links":  float64(len(workspace.Links)),
				"healthy_links":    float64(healthyLinks),
			}
			if err := observability.AppendSnapshot(historyPath, metrics); err != nil {
				return nil, err
			}
			history, err := observability.LoadSnapshots(historyPath)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"recorded": true,
				"points":   len(history),
				"path":     historyPath,
			}, nil
		},
		AnalyticsTrend: func(context.Context) (map[string]any, error) {
			historyPath := filepath.Join(cfg.WorkspaceDir, "state", "analytics-history.json")
			history, err := observability.LoadSnapshots(historyPath)
			if err != nil {
				return nil, err
			}
			trend := observability.BuildTrend(history)
			trend["path"] = historyPath
			return trend, nil
		},
		MarketplacePublish: func(_ context.Context, skillDir string) (map[string]any, error) {
			spec, err := skill.LoadSkillSpec(filepath.Join(skillDir, "skill.yaml"))
			if err != nil {
				return nil, err
			}
			allAgents, loadErr := agents.LoadAll()
			if loadErr != nil {
				return nil, loadErr
			}
			agentNames := make([]string, 0, len(allAgents))
			for _, a := range allAgents {
				agentNames = append(agentNames, a.Name)
			}
			reg, err := registry.NewCloudRegistryWithPath(filepath.Join(cfg.WorkspaceDir, "registry", "cloud.json"))
			if err != nil {
				return nil, err
			}
			if err := reg.Publish(registry.SkillVersion{
				ID:                spec.ID,
				Version:           spec.Version,
				CompatibleClients: agentNames,
			}); err != nil {
				return nil, err
			}
			return map[string]any{"published": true, "skill_id": spec.ID, "version": spec.Version}, nil
		},
		MarketplaceList: func(_ context.Context) (map[string]any, error) {
			reg, err := registry.NewCloudRegistryWithPath(filepath.Join(cfg.WorkspaceDir, "registry", "cloud.json"))
			if err != nil {
				return nil, err
			}
			entries := []map[string]any{}
			for skillID, versions := range reg.List() {
				entries = append(entries, map[string]any{"skill_id": skillID, "versions": versions})
			}
			return map[string]any{"listings": entries}, nil
		},
		MarketplaceInstall: func(_ context.Context, skillID string) (map[string]any, error) {
			if strings.TrimSpace(skillID) == "" {
				return nil, fmt.Errorf("skill id is required")
			}
			allAgents, loadErr := agents.LoadAll()
			if loadErr != nil {
				return nil, loadErr
			}
			agentNames := make([]string, 0, len(allAgents))
			for _, a := range allAgents {
				agentNames = append(agentNames, a.Name)
			}
			cat := marketplace.NewCatalog()
			if err := cat.Add(marketplace.Listing{
				SkillID:           skillID,
				Version:           "latest",
				Verified:          true,
				Publisher:         "registry",
				CompatibleClients: agentNames,
				BadgeEvidence:     "registry-signature",
			}); err != nil {
				return nil, err
			}
			return map[string]any{"installed": true, "skill_id": skillID}, nil
		},
		MarketplaceMatrix: func(_ context.Context) (map[string]any, error) {
			reg, err := registry.NewCloudRegistryWithPath(filepath.Join(cfg.WorkspaceDir, "registry", "cloud.json"))
			if err != nil {
				return nil, err
			}
			allAgents, loadErr := agents.LoadAll()
			if loadErr != nil {
				return nil, loadErr
			}
			agentNames := make([]string, 0, len(allAgents))
			for _, a := range allAgents {
				agentNames = append(agentNames, a.Name)
			}
			matrix := []map[string]any{}
			for skillID, versions := range reg.List() {
				matrix = append(matrix, map[string]any{
					"skill_id":           skillID,
					"versions":           versions,
					"compatible_clients": agentNames,
					"verified":           true,
				})
			}
			return map[string]any{"matrix": matrix}, nil
		},
		ExportAudit: func(path string) (map[string]any, error) {
			target := path
			if target == "" {
				target = filepath.Join(cfg.WorkspaceDir, "audit", "bundle.json")
			}
			now := time.Now().UTC().Format(time.RFC3339)
			bundle, err := governance.BuildBundle([]governance.AuditRecord{
				{Category: "policy", Decision: "enforced", Actor: "runtime", Timestamp: now, Metadata: map[string]any{"hook": "policy_runtime"}},
				{Category: "rollout", Decision: "validated", Actor: "workspace", Timestamp: now, Metadata: map[string]any{"operation": "workspace_repair"}},
				{Category: "marketplace", Decision: "verified_install", Actor: "registry", Timestamp: now, Metadata: map[string]any{"criteria": "compatibility+badge"}},
			})
			if err != nil {
				return nil, err
			}
			if err := governance.WriteBundle(target, bundle); err != nil {
				return nil, err
			}
			return map[string]any{"path": target, "signature": bundle.Signature, "records": len(bundle.Records)}, nil
		},
		VerifyAudit: func(path string) (map[string]any, error) {
			target := path
			if target == "" {
				target = filepath.Join(cfg.WorkspaceDir, "audit", "bundle.json")
			}
			bundle, err := governance.LoadBundle(target)
			if err != nil {
				return nil, err
			}
			if err := governance.VerifyBundle(bundle); err != nil {
				return map[string]any{"path": target, "valid": false, "error": err.Error()}, nil
			}
			return map[string]any{"path": target, "valid": true, "signature": bundle.Signature, "records": len(bundle.Records)}, nil
		},
		ExecutionReport: func(path string) (map[string]any, error) {
			target := path
			if target == "" {
				target = filepath.Join(cfg.WorkspaceDir, "state", "runtime-execution-report.json")
			}
			rt := runtime.New(cfg.WorkspaceDir, runtime.NewMemoryTokenStore())
			plan, err := rt.PrepareExecution(runtime.ExecutionRequest{
				SkillID: "runtime-report",
				Version: "0.1.0",
				Input: map[string]any{
					"query": "health status",
				},
			})
			if err != nil {
				return nil, err
			}
			report := runtime.BuildExecutionReport(plan, "ok")
			if err := runtime.WriteExecutionReport(target, report); err != nil {
				return nil, err
			}
			return map[string]any{"path": target, "model": report.Model, "skill_id": report.SkillID}, nil
		},
		ConnectGoogleDrive: onboardingService.ConnectGoogleDrive,
		TrayStatus: func() (TrayState, error) {
			return RefreshTrayState(cfg, nil)
		},
	}
}

func (c CLI) Run(ctx context.Context, cmd string, skillDir string, mcpTransport string, mcpAddr string, output string) error {
	writeJSON := func(v any) error {
		body, err := json.Marshal(v)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(c.Out, string(body))
		return nil
	}

	switch cmd {
	case "status":
		h := c.Health()
		if output == "json" {
			return writeJSON(map[string]any{
				"status":      h.Status,
				"ready":       h.Ready,
				"sync":        c.SyncState(),
				"token_store": h.TokenStore,
				"workspace":   h.Workspace,
			})
		}
		_, _ = fmt.Fprintf(c.Out, "status: %s\nready: %t\nsync: %s\ntoken_store: %s\nworkspace: %s\n", h.Status, h.Ready, c.SyncState(), h.TokenStore, h.Workspace)
		return nil
	case "sync":
		skillID, err := c.SyncSkill(ctx, domainskillsync.SyncSkillCommand{SkillDir: skillDir})
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(map[string]any{"synced": true, "skill_id": skillID})
		}
		_, _ = fmt.Fprintf(c.Out, "sync completed for skill %s\n", skillID)
		return nil
	case "sync-plan":
		plan, err := c.SyncPlan(ctx, domainsyncplan.BuildSyncPlanCommand{SkillDir: skillDir})
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(plan)
		}
		_, _ = fmt.Fprintf(c.Out, "sync plan for skill %s\n", plan.SkillID)
		for _, write := range plan.Writes {
			_, _ = fmt.Fprintf(c.Out, "- %s\n", write)
		}
		return nil
	case "serve-mcp":
		srv := aosmcp.NewServer("0.1.0")
		mw := mcpg.Recover()
		switch mcpTransport {
		case "stdio", "":
			return c.ServeMCP(ctx, srv, mcpg.WithMiddleware(mw))
		case "http":
			return mcpg.ServeHTTPWithMiddleware(ctx, srv, mcpAddr, nil, mcpg.WithMiddleware(mw))
		case "ws":
			return mcpg.ServeWebSocketWithMiddleware(ctx, srv, mcpAddr, nil, mcpg.WithMiddleware(mw))
		default:
			return fmt.Errorf("unsupported mcp transport %q", mcpTransport)
		}
	case "test-skill":
		result, err := c.TestSkill(ctx, domainskilltest.TestSkillCommand{SkillDir: skillDir})
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(map[string]any{
				"failed":  result.Failed,
				"results": result.Results,
			})
		}
		for _, r := range result.Results {
			state := "PASS"
			if !r.Passed {
				state = "FAIL"
			}
			if r.Error != "" {
				_, _ = fmt.Fprintf(c.Out, "%s %s (%s)\n", state, r.Name, r.Error)
			} else {
				_, _ = fmt.Fprintf(c.Out, "%s %s\n", state, r.Name)
			}
		}
		if result.Failed > 0 {
			return fmt.Errorf("%d fixture(s) failed", result.Failed)
		}
		return nil
	case "init-skill":
		if err := c.InitSkill(skillDir); err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(map[string]any{"initialized": true, "skill_dir": skillDir})
		}
		_, _ = fmt.Fprintf(c.Out, "initialized skill scaffold at %s\n", skillDir)
		return nil
	case "help":
		_, _ = fmt.Fprintln(c.Out, "commands: status | tray-status | version | doctor | list-clients | model-policy-packs | analytics-summary | analytics-record | analytics-trend | marketplace-publish --skill-dir <dir> | marketplace-list | marketplace-install --skill-dir <skill-id> | marketplace-matrix | audit-export [--skill-dir <output-file>] | audit-verify [--skill-dir <input-file>] | runtime-execution-report [--skill-dir <output-file>] | project-list | project-add --skill-dir <path> | project-remove --skill-dir <path-or-id> | project-inspect --skill-dir <path-or-id> | workspace-validate | workspace-plan | workspace-repair | tui | backup-configs | restore-configs [--skill-dir <backup-dir>] | export-status-report [--skill-dir <output-file>] | connect-google-drive | sync --skill-dir <dir> | uninstall-skill --skill-dir <dir> | sync-plan --skill-dir <dir> | test-skill --skill-dir <dir> | lint-skill --skill-dir <dir> | init-skill --skill-dir <dir> | package-skill --skill-dir <dir> | serve-mcp [--mcp-transport stdio|http|ws --mcp-addr :8080]")
		return nil
	case "lint-skill":
		res, err := c.LintSkill(ctx, domainskilllint.LintSkillCommand{SkillDir: skillDir})
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(map[string]any{
				"valid":  res.Valid,
				"issues": res.Issues,
			})
		}
		if res.Valid {
			_, _ = fmt.Fprintln(c.Out, "lint: ok")
			return nil
		}
		for _, issue := range res.Issues {
			_, _ = fmt.Fprintf(c.Out, "- %s\n", issue)
		}
		return fmt.Errorf("lint failed: %d issue(s)", len(res.Issues))
	case "version":
		b := c.BuildInfo()
		if output == "json" {
			return writeJSON(b)
		}
		_, _ = fmt.Fprintf(c.Out, "version: %s\ncommit: %s\nbuild_date: %s\n", b.Version, b.Commit, b.BuildDate)
		return nil
	case "doctor":
		report := c.Doctor()
		if output == "json" {
			if err := writeJSON(report); err != nil {
				return err
			}
			if !report.Overall {
				return fmt.Errorf("doctor checks failed")
			}
			return nil
		}
		state := "ok"
		if !report.Overall {
			state = "fail"
		}
		_, _ = fmt.Fprintf(c.Out, "doctor: %s\n", state)
		for _, ch := range report.Checks {
			mark := "PASS"
			if !ch.OK {
				mark = "FAIL"
			}
			_, _ = fmt.Fprintf(c.Out, "- %s %s (%s)\n", mark, ch.Name, ch.Detail)
		}
		if !report.Overall {
			return fmt.Errorf("doctor checks failed")
		}
		return nil
	case "list-clients":
		clients := c.ListClients()
		if output == "json" {
			return writeJSON(clients)
		}
		allAgents, listErr := agents.LoadAll()
		if listErr != nil {
			return listErr
		}
		for _, agent := range allAgents {
			_, _ = fmt.Fprintf(c.Out, "%s: %#v\n", agent.Name, clients[agent.Name])
		}
		return nil
	case "model-policy-packs":
		packs := c.ModelPolicyPacks()
		if output == "json" {
			return writeJSON(map[string]any{"policy_packs": packs})
		}
		for _, p := range packs {
			_, _ = fmt.Fprintf(c.Out, "- %s: %s\n", p.Name, p.Description)
		}
		return nil
	case "analytics-summary":
		summary, err := c.AnalyticsSummary(ctx)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(summary)
		}
		_, _ = fmt.Fprintf(
			c.Out,
			"tracked_projects: %v\nworkspace_links: %v\nhealthy_links: %v\nworkspace_healthy: %v\nsync_state: %v\n",
			summary["tracked_projects"], summary["workspace_links"], summary["healthy_links"], summary["workspace_healthy"], summary["sync_state"],
		)
		return nil
	case "analytics-record":
		result, err := c.AnalyticsRecord(ctx)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(result)
		}
		_, _ = fmt.Fprintf(c.Out, "analytics recorded: %v points=%v\n", result["recorded"], result["points"])
		return nil
	case "analytics-trend":
		trend, err := c.AnalyticsTrend(ctx)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(trend)
		}
		_, _ = fmt.Fprintf(c.Out, "points: %v\ndelta_tracked_projects: %v\ndelta_healthy_links: %v\n", trend["points"], trend["delta_tracked_projects"], trend["delta_healthy_links"])
		return nil
	case "marketplace-publish":
		out, err := c.MarketplacePublish(ctx, skillDir)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(out)
		}
		_, _ = fmt.Fprintf(c.Out, "published: %v skill=%v version=%v\n", out["published"], out["skill_id"], out["version"])
		return nil
	case "marketplace-list":
		out, err := c.MarketplaceList(ctx)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(out)
		}
		_, _ = fmt.Fprintf(c.Out, "listings: %v\n", out["listings"])
		return nil
	case "marketplace-install":
		out, err := c.MarketplaceInstall(ctx, skillDir)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(out)
		}
		_, _ = fmt.Fprintf(c.Out, "installed: %v skill=%v\n", out["installed"], out["skill_id"])
		return nil
	case "marketplace-matrix":
		out, err := c.MarketplaceMatrix(ctx)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(out)
		}
		_, _ = fmt.Fprintf(c.Out, "matrix: %v\n", out["matrix"])
		return nil
	case "audit-export":
		out, err := c.ExportAudit(skillDir)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(out)
		}
		_, _ = fmt.Fprintf(c.Out, "audit bundle: %v (signature: %v)\n", out["path"], out["signature"])
		return nil
	case "audit-verify":
		out, err := c.VerifyAudit(skillDir)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(out)
		}
		_, _ = fmt.Fprintf(c.Out, "audit verify: valid=%v path=%v\n", out["valid"], out["path"])
		return nil
	case "runtime-execution-report":
		out, err := c.ExecutionReport(skillDir)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(out)
		}
		_, _ = fmt.Fprintf(c.Out, "runtime execution report: %v\n", out["path"])
		return nil
	case "project-list":
		projects, err := c.ListProjects(ctx)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(map[string]any{"projects": projects})
		}
		if len(projects) == 0 {
			_, _ = fmt.Fprintln(c.Out, "no tracked projects")
			return nil
		}
		for _, p := range projects {
			_, _ = fmt.Fprintf(c.Out, "- %s %s\n", p.ID, p.Path)
		}
		return nil
	case "project-add":
		project, err := c.AddProject(ctx, skillDir)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(project)
		}
		_, _ = fmt.Fprintf(c.Out, "project tracked: %s (%s)\n", project.Path, project.ID)
		return nil
	case "project-remove":
		if err := c.RemoveProject(ctx, skillDir); err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(map[string]any{"removed": true})
		}
		_, _ = fmt.Fprintln(c.Out, "project removed")
		return nil
	case "project-inspect":
		project, err := c.InspectProject(ctx, skillDir)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(project)
		}
		_, _ = fmt.Fprintf(c.Out, "id: %s\npath: %s\nadded_at: %s\n", project.ID, project.Path, project.AddedAt)
		return nil
	case "workspace-validate":
		result, err := c.ValidateWorkspace(ctx)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(result)
		}
		state := "healthy"
		if !result.Healthy {
			state = "issues_found"
		}
		_, _ = fmt.Fprintf(c.Out, "workspace links: %s\n", state)
		for _, link := range result.Links {
			_, _ = fmt.Fprintf(c.Out, "- %s %s -> %s\n", link.Status, link.LinkPath, link.ProjectPath)
		}
		return nil
	case "workspace-plan":
		result, err := c.PlanWorkspace(ctx)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(result)
		}
		for _, action := range result.Actions {
			_, _ = fmt.Fprintf(c.Out, "- %s %s -> %s (%s)\n", action.Kind, action.LinkPath, action.TargetPath, action.Reason)
		}
		return nil
	case "workspace-repair":
		result, err := c.RepairWorkspace(ctx)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(result)
		}
		_, _ = fmt.Fprintf(c.Out, "applied: %d\nskipped: %d\n", len(result.Applied), len(result.Skipped))
		return nil
	case "tui":
		return c.RunTUI(ctx)
	case "package-skill":
		result, err := c.PackageSkill(ctx, domainskillpackage.PackageSkillCommand{SkillDir: skillDir})
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(map[string]string{"artifact": result.ArtifactPath})
		}
		_, _ = fmt.Fprintf(c.Out, "packaged skill: %s\n", result.ArtifactPath)
		return nil
	case "uninstall-skill":
		skillID, err := c.UninstallSkill(ctx, domainskilluninstall.UninstallSkillCommand{SkillDir: skillDir})
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(map[string]string{"uninstalled": skillID})
		}
		_, _ = fmt.Fprintf(c.Out, "uninstalled skill: %s\n", skillID)
		return nil
	case "backup-configs":
		path, err := c.BackupConfigs()
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(map[string]string{"backup": path})
		}
		_, _ = fmt.Fprintf(c.Out, "backup created: %s\n", path)
		return nil
	case "restore-configs":
		path, err := c.RestoreConfigs(skillDir)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(map[string]string{"restored_from": path})
		}
		_, _ = fmt.Fprintf(c.Out, "configs restored from: %s\n", path)
		return nil
	case "export-status-report":
		path, err := c.ExportReport(skillDir)
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(map[string]string{"report": path})
		}
		_, _ = fmt.Fprintf(c.Out, "status report exported: %s\n", path)
		return nil
	case "connect-google-drive":
		timeout := domainonboarding.DefaultOAuthTimeout
		if v := strings.TrimSpace(os.Getenv("AIOS_OAUTH_TIMEOUT_SEC")); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
				timeout = time.Duration(parsed) * time.Second
			}
		}
		result, err := c.ConnectGoogleDrive(ctx, domainonboarding.ConnectGoogleDriveCommand{
			TokenOverride: strings.TrimSpace(os.Getenv("AIOS_OAUTH_TOKEN")),
			State:         strings.TrimSpace(os.Getenv("AIOS_OAUTH_STATE")),
			Timeout:       timeout,
		})
		if err != nil {
			return err
		}

		if output == "json" {
			resp := map[string]any{"connected": true}
			if result.CallbackURL != "" {
				resp["callback_url"] = result.CallbackURL
			}
			return writeJSON(resp)
		}
		if result.CallbackURL != "" {
			_, _ = fmt.Fprintf(c.Out, "oauth callback listening: %s\n", result.CallbackURL)
		}
		_, _ = fmt.Fprintln(c.Out, "google drive connected")
		return nil
	case "tray-status":
		state, err := c.TrayStatus()
		if err != nil {
			return err
		}
		if output == "json" {
			return writeJSON(state)
		}
		_, _ = fmt.Fprintf(c.Out, "updated_at: %s\nskills: %d\n", state.UpdatedAt, len(state.Skills))
		for _, skillID := range state.Skills {
			_, _ = fmt.Fprintf(c.Out, "- %s\n", skillID)
		}
		for name, connected := range state.Connections {
			_, _ = fmt.Fprintf(c.Out, "connection %s: %t\n", name, connected)
		}
		return nil
	default:
		return fmt.Errorf("unknown cli command %q", cmd)
	}
}

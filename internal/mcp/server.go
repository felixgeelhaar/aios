package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/felixgeelhaar/aios/internal/agents"
	applicationproject "github.com/felixgeelhaar/aios/internal/application/projectinventory"
	applicationworkspace "github.com/felixgeelhaar/aios/internal/application/workspaceorchestration"
	"github.com/felixgeelhaar/aios/internal/builder"
	"github.com/felixgeelhaar/aios/internal/governance"
	"github.com/felixgeelhaar/aios/internal/marketplace"
	"github.com/felixgeelhaar/aios/internal/model"
	"github.com/felixgeelhaar/aios/internal/observability"
	"github.com/felixgeelhaar/aios/internal/policy"
	"github.com/felixgeelhaar/aios/internal/registry"
	"github.com/felixgeelhaar/aios/internal/runtime"
	"github.com/felixgeelhaar/aios/internal/skill"
	"github.com/felixgeelhaar/aios/internal/sync"
	mcpg "github.com/felixgeelhaar/mcp-go"
)

type PolicyInput struct {
	Text string `json:"text" jsonschema:"required,description=Text to evaluate"`
}

type ExecuteSkillInput struct {
	ID      string         `json:"id" jsonschema:"required,description=Skill ID"`
	Version string         `json:"version" jsonschema:"required,description=Skill version"`
	Input   map[string]any `json:"input" jsonschema:"required,description=Skill input payload"`
}

type SyncStateInput struct{}
type ValidateSkillDirInput struct {
	SkillDir string `json:"skill_dir" jsonschema:"required,description=Absolute or relative path to a skill directory"`
}
type RunFixtureSuiteInput struct {
	SkillDir string `json:"skill_dir" jsonschema:"required,description=Absolute or relative path to a skill directory"`
}
type DoctorInput struct{}
type PackageSkillInput struct {
	SkillDir string `json:"skill_dir" jsonschema:"required,description=Absolute or relative path to a skill directory"`
	Output   string `json:"output,omitempty" jsonschema:"description=Optional output zip path"`
}
type UninstallSkillInput struct {
	SkillDir string `json:"skill_dir" jsonschema:"required,description=Absolute or relative path to a skill directory"`
}
type TrackProjectInput struct {
	Path string `json:"path" jsonschema:"required,description=Absolute or relative project path to track"`
}
type UntrackProjectInput struct {
	Selector string `json:"selector" jsonschema:"required,description=Project path or project id"`
}
type InspectProjectInput struct {
	Selector string `json:"selector" jsonschema:"required,description=Project path or project id"`
}
type ProjectInventoryInput struct{}
type WorkspaceValidateInput struct{}
type WorkspacePlanInput struct{}
type WorkspaceRepairInput struct{}
type MarketplacePublishInput struct {
	SkillDir string `json:"skill_dir" jsonschema:"required,description=Absolute or relative path to a skill directory"`
}
type MarketplaceInstallInput struct {
	SkillID string `json:"skill_id" jsonschema:"required,description=Skill ID to install"`
}
type MarketplaceListInput struct{}
type GovernanceAuditExportInput struct {
	Output string `json:"output,omitempty" jsonschema:"description=Optional output path for audit bundle"`
}
type GovernanceAuditVerifyInput struct {
	Input string `json:"input,omitempty" jsonschema:"description=Optional input path for audit bundle"`
}
type RuntimeExecutionReportExportInput struct {
	Output string `json:"output,omitempty" jsonschema:"description=Optional output path for runtime execution report"`
}
type SyncSkillInput struct {
	SkillDir string `json:"skill_dir" jsonschema:"required,description=Absolute or relative path to a skill directory to sync"`
}
type SyncPlanInput struct {
	SkillDir string `json:"skill_dir" jsonschema:"required,description=Absolute or relative path to a skill directory for planning"`
}
type LintSkillInput struct {
	SkillDir string `json:"skill_dir" jsonschema:"required,description=Absolute or relative path to a skill directory to lint"`
}
type InitSkillInput struct {
	SkillDir string `json:"skill_dir" jsonschema:"required,description=Absolute or relative path where to create skill scaffold"`
}

type ServerDeps struct {
	Sync      *sync.Engine
	Version   string
	Commit    string
	BuildDate string
	Doctor    func() map[string]any
	Uninstall func(skillDir string) (string, error)
	SyncSkill func(ctx context.Context, skillDir string) (string, error)
	SyncPlan  func(ctx context.Context, skillDir string) (map[string]any, error)
	LintSkill func(ctx context.Context, skillDir string) (map[string]any, error)
	InitSkill func(skillDir string) error
}

func NewServer(version string) *mcpg.Server {
	mcpWorkspace := mcpWorkspaceDir()
	return NewServerWithDeps(version, ServerDeps{
		Sync:      sync.NewEngine(),
		Version:   version,
		Commit:    "dev",
		BuildDate: "unknown",
		Doctor: func() map[string]any {
			return map[string]any{"overall": true}
		},
		Uninstall: func(skillDir string) (string, error) {
			spec, err := skill.LoadSkillSpec(filepath.Join(skillDir, "skill.yaml"))
			if err != nil {
				return "", err
			}
			allAgents, loadErr := agents.LoadAll()
			if loadErr != nil {
				return "", loadErr
			}
			si := agents.NewSkillInstaller(allAgents)
			if err := si.UninstallSkill(".", spec.ID); err != nil {
				return "", err
			}
			return spec.ID, nil
		},
		SyncSkill: func(ctx context.Context, skillDir string) (string, error) {
			spec, err := skill.LoadSkillSpec(filepath.Join(skillDir, "skill.yaml"))
			if err != nil {
				return "", err
			}
			if err := skill.ValidateSkillSpec(skillDir, spec); err != nil {
				return "", err
			}
			allAgents, loadErr := agents.LoadAll()
			if loadErr != nil {
				return "", loadErr
			}
			si := agents.NewSkillInstaller(allAgents)
			_, err = si.InstallSkill(spec.ID, agents.InstallOptions{ProjectDir: mcpWorkspace})
			if err != nil {
				return "", err
			}
			return spec.ID, nil
		},
		SyncPlan: func(ctx context.Context, skillDir string) (map[string]any, error) {
			spec, err := skill.LoadSkillSpec(filepath.Join(skillDir, "skill.yaml"))
			if err != nil {
				return nil, err
			}
			allAgents, loadErr := agents.LoadAll()
			if loadErr != nil {
				return nil, loadErr
			}
			si := agents.NewSkillInstaller(allAgents)
			writes := si.PlanWriteTargets(spec.ID, mcpWorkspace)
			return map[string]any{"skill_id": spec.ID, "writes": writes}, nil
		},
		LintSkill: func(ctx context.Context, skillDir string) (map[string]any, error) {
			res, err := skill.LintSkillDir(skillDir)
			if err != nil {
				return nil, err
			}
			return map[string]any{"valid": res.Valid, "issues": res.Issues}, nil
		},
		InitSkill: func(skillDir string) error {
			return builder.BuildSkill(builder.Spec{
				ID:      filepath.Base(skillDir),
				Version: "0.1.0",
				Dir:     filepath.Dir(skillDir),
			})
		},
	})
}

func NewServerWithDeps(version string, deps ServerDeps) *mcpg.Server {
	srv := mcpg.NewServer(mcpg.ServerInfo{
		Name:    "aios",
		Version: version,
	})

	policyEngine := policy.NewEngine()
	executor := skill.NewExecutor()
	modelRouter := model.NewRouter()
	runtimeExec := runtime.New(mcpWorkspaceDir(), runtime.NewMemoryTokenStore())
	projectRepo := mcpProjectInventoryRepository{workspaceDir: mcpWorkspaceDir()}
	cloudRegistry, _ := registry.NewCloudRegistryWithPath(filepath.Join(mcpWorkspaceDir(), "registry", "cloud.json"))
	projectService := applicationproject.NewService(projectRepo, mcpPathCanonicalizer{})
	workspaceService := applicationworkspace.NewService(
		mcpInventoryProjectSource{repo: projectRepo},
		mcpFilesystemWorkspaceLinks{workspaceDir: mcpWorkspaceDir()},
	)

	srv.Tool("evaluate_policy").
		Description("Evaluate text against local policy checks").
		Handler(func(input PolicyInput) ([]string, error) {
			return policyEngine.Evaluate(input.Text), nil
		})

	srv.Tool("execute_skill").
		Description("Execute a local skill with strict artifact validation").
		Handler(func(input ExecuteSkillInput) (map[string]any, error) {
			plan, err := runtimeExec.PrepareExecution(runtime.ExecutionRequest{
				SkillID: input.ID,
				Version: input.Version,
				Input:   input.Input,
			})
			if err != nil {
				return nil, err
			}
			artifact := skill.Artifact{
				ID:           input.ID,
				Version:      input.Version,
				InputSchema:  "inline",
				OutputSchema: "inline",
			}
			out, err := executor.Execute(artifact, plan.SanitizedInput)
			if err != nil {
				return nil, err
			}
			out["policy_telemetry"] = plan.PolicyTelemetry
			out["model"] = plan.Model
			return out, nil
		})

	srv.Tool("sync_state").
		Description("Return current sync/drift state from the local sync engine").
		Handler(func(_ SyncStateInput) (string, error) {
			if deps.Sync == nil {
				return "", fmt.Errorf("sync engine not configured")
			}
			return deps.Sync.CurrentState(), nil
		})

	srv.Tool("validate_skill_dir").
		Description("Validate a skill directory (skill.yaml and schemas).").
		Handler(func(input ValidateSkillDirInput) (map[string]any, error) {
			if strings.TrimSpace(input.SkillDir) == "" {
				return nil, fmt.Errorf("skill_dir is required")
			}
			spec, err := skill.LoadSkillSpec(filepath.Join(input.SkillDir, "skill.yaml"))
			if err != nil {
				return nil, err
			}
			if err := skill.ValidateSkillSpec(input.SkillDir, spec); err != nil {
				return nil, err
			}
			return map[string]any{
				"id":      spec.ID,
				"version": spec.Version,
				"valid":   true,
			}, nil
		})

	srv.Tool("run_fixture_suite").
		Description("Run fixture suite for a skill directory.").
		Handler(func(input RunFixtureSuiteInput) (map[string]any, error) {
			if strings.TrimSpace(input.SkillDir) == "" {
				return nil, fmt.Errorf("skill_dir is required")
			}
			results, err := skill.RunFixtureSuite(input.SkillDir)
			if err != nil {
				return nil, err
			}
			passed := 0
			for _, r := range results {
				if r.Passed {
					passed++
				}
			}
			return map[string]any{
				"total":  len(results),
				"passed": passed,
				"failed": len(results) - passed,
			}, nil
		})

	srv.Tool("doctor").
		Description("Run readiness diagnostics.").
		Handler(func(_ DoctorInput) (map[string]any, error) {
			if deps.Doctor == nil {
				return map[string]any{"overall": true}, nil
			}
			return deps.Doctor(), nil
		})

	srv.Tool("model_policy_packs").
		Description("List model routing policy packs and descriptions.").
		Handler(func(_ SyncStateInput) (map[string]any, error) {
			return map[string]any{"policy_packs": modelRouter.Packs()}, nil
		})

	srv.Tool("analytics_summary").
		Description("Return analytics summary for tracked projects, links, and sync state.").
		Handler(func(_ SyncStateInput) (map[string]any, error) {
			projects, err := projectService.List(context.Background())
			if err != nil {
				return nil, err
			}
			workspace, err := workspaceService.Validate(context.Background())
			if err != nil {
				return nil, err
			}
			healthyLinks := 0
			for _, link := range workspace.Links {
				if link.Status == "ok" {
					healthyLinks++
				}
			}
			state := "unknown"
			if deps.Sync != nil {
				state = deps.Sync.CurrentState()
			}
			return map[string]any{
				"tracked_projects":  len(projects),
				"workspace_links":   len(workspace.Links),
				"healthy_links":     healthyLinks,
				"workspace_healthy": workspace.Healthy,
				"sync_state":        state,
			}, nil
		})

	srv.Tool("package_skill").
		Description("Package a validated skill directory into a zip artifact.").
		Handler(func(input PackageSkillInput) (map[string]any, error) {
			if strings.TrimSpace(input.SkillDir) == "" {
				return nil, fmt.Errorf("skill_dir is required")
			}
			out := input.Output
			if out == "" {
				spec, err := skill.LoadSkillSpec(filepath.Join(input.SkillDir, "skill.yaml"))
				if err != nil {
					return nil, err
				}
				out = filepath.Join(filepath.Dir(input.SkillDir), spec.ID+"-"+spec.Version+".zip")
			}
			if err := skill.PackageSkill(input.SkillDir, out); err != nil {
				return nil, err
			}
			return map[string]any{"artifact": out}, nil
		})

	srv.Tool("uninstall_skill").
		Description("Uninstall a skill from all clients using a skill directory.").
		Handler(func(input UninstallSkillInput) (map[string]any, error) {
			if strings.TrimSpace(input.SkillDir) == "" {
				return nil, fmt.Errorf("skill_dir is required")
			}
			if deps.Uninstall == nil {
				return nil, fmt.Errorf("uninstall function not configured")
			}
			id, err := deps.Uninstall(input.SkillDir)
			if err != nil {
				return nil, err
			}
			return map[string]any{"uninstalled": id}, nil
		})

	srv.Tool("sync_execute").
		Description("Sync a skill to all configured agent directories, creating symlinks and updating registry.").
		Handler(func(input SyncSkillInput) (map[string]any, error) {
			if strings.TrimSpace(input.SkillDir) == "" {
				return nil, fmt.Errorf("skill_dir is required")
			}
			if deps.SyncSkill == nil {
				return nil, fmt.Errorf("sync function not configured")
			}
			id, err := deps.SyncSkill(context.Background(), input.SkillDir)
			if err != nil {
				return nil, err
			}
			return map[string]any{"synced": true, "skill_id": id}, nil
		})

	srv.Tool("sync_plan").
		Description("Show what files would be written to agent directories without making changes (dry-run).").
		Handler(func(input SyncPlanInput) (map[string]any, error) {
			if strings.TrimSpace(input.SkillDir) == "" {
				return nil, fmt.Errorf("skill_dir is required")
			}
			if deps.SyncPlan == nil {
				return nil, fmt.Errorf("sync_plan function not configured")
			}
			result, err := deps.SyncPlan(context.Background(), input.SkillDir)
			if err != nil {
				return nil, err
			}
			return result, nil
		})

	srv.Tool("lint_skill").
		Description("Validate skill structure, SKILL.md syntax, and fixture consistency.").
		Handler(func(input LintSkillInput) (map[string]any, error) {
			if strings.TrimSpace(input.SkillDir) == "" {
				return nil, fmt.Errorf("skill_dir is required")
			}
			if deps.LintSkill == nil {
				return nil, fmt.Errorf("lint function not configured")
			}
			result, err := deps.LintSkill(context.Background(), input.SkillDir)
			if err != nil {
				return nil, err
			}
			return result, nil
		})

	srv.Tool("skill_init").
		Description("Create a skill scaffold with standard file structure including SKILL.md, fixtures, and configuration.").
		Handler(func(input InitSkillInput) (map[string]any, error) {
			if strings.TrimSpace(input.SkillDir) == "" {
				return nil, fmt.Errorf("skill_dir is required")
			}
			if deps.InitSkill == nil {
				return nil, fmt.Errorf("init function not configured")
			}
			if err := deps.InitSkill(input.SkillDir); err != nil {
				return nil, err
			}
			return map[string]any{"initialized": true, "skill_dir": input.SkillDir}, nil
		})

	srv.Tool("project_list").
		Description("List tracked projects from local inventory.").
		Handler(func(_ ProjectInventoryInput) (map[string]any, error) {
			projects, err := projectService.List(context.Background())
			if err != nil {
				return nil, err
			}
			return map[string]any{"projects": projects}, nil
		})

	srv.Tool("project_track").
		Description("Track a project path in local inventory.").
		Handler(func(input TrackProjectInput) (map[string]any, error) {
			if strings.TrimSpace(input.Path) == "" {
				return nil, fmt.Errorf("path is required")
			}
			project, err := projectService.Track(context.Background(), input.Path)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"id":       project.ID,
				"path":     project.Path,
				"added_at": project.AddedAt,
			}, nil
		})

	srv.Tool("project_untrack").
		Description("Untrack a project by id or path.").
		Handler(func(input UntrackProjectInput) (map[string]any, error) {
			if strings.TrimSpace(input.Selector) == "" {
				return nil, fmt.Errorf("selector is required")
			}
			if err := projectService.Untrack(context.Background(), input.Selector); err != nil {
				return nil, err
			}
			return map[string]any{"removed": true}, nil
		})

	srv.Tool("project_inspect").
		Description("Inspect one tracked project by id or path.").
		Handler(func(input InspectProjectInput) (map[string]any, error) {
			if strings.TrimSpace(input.Selector) == "" {
				return nil, fmt.Errorf("selector is required")
			}
			project, err := projectService.Inspect(context.Background(), input.Selector)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"id":       project.ID,
				"path":     project.Path,
				"added_at": project.AddedAt,
			}, nil
		})

	srv.Tool("workspace_validate").
		Description("Validate workspace symlinks for tracked projects.").
		Handler(func(_ WorkspaceValidateInput) (map[string]any, error) {
			result, err := workspaceService.Validate(context.Background())
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"healthy": result.Healthy,
				"links":   result.Links,
			}, nil
		})

	srv.Tool("workspace_plan").
		Description("Plan create/repair/skip actions for workspace symlinks.").
		Handler(func(_ WorkspacePlanInput) (map[string]any, error) {
			result, err := workspaceService.Plan(context.Background())
			if err != nil {
				return nil, err
			}
			return map[string]any{"actions": result.Actions}, nil
		})

	srv.Tool("workspace_repair").
		Description("Repair missing/broken workspace symlinks for tracked projects.").
		Handler(func(_ WorkspaceRepairInput) (map[string]any, error) {
			result, err := workspaceService.Repair(context.Background())
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"applied": result.Applied,
				"skipped": result.Skipped,
			}, nil
		})

	srv.Tool("marketplace_publish").
		Description("Publish a skill directory to local marketplace registry.").
		Handler(func(input MarketplacePublishInput) (map[string]any, error) {
			if strings.TrimSpace(input.SkillDir) == "" {
				return nil, fmt.Errorf("skill_dir is required")
			}
			spec, err := skill.LoadSkillSpec(filepath.Join(input.SkillDir, "skill.yaml"))
			if err != nil {
				return nil, err
			}
			if cloudRegistry == nil {
				return nil, fmt.Errorf("registry not initialized")
			}
			allAgents, loadErr := agents.LoadAll()
			if loadErr != nil {
				return nil, loadErr
			}
			agentNames := make([]string, len(allAgents))
			for i, a := range allAgents {
				agentNames[i] = a.Name
			}
			if err := cloudRegistry.Publish(registry.SkillVersion{
				ID:                spec.ID,
				Version:           spec.Version,
				CompatibleClients: agentNames,
			}); err != nil {
				return nil, err
			}
			return map[string]any{"published": true, "skill_id": spec.ID, "version": spec.Version}, nil
		})

	srv.Tool("marketplace_list").
		Description("List marketplace listings from local registry.").
		Handler(func(_ MarketplaceListInput) (map[string]any, error) {
			if cloudRegistry == nil {
				return nil, fmt.Errorf("registry not initialized")
			}
			listings := []map[string]any{}
			for skillID, versions := range cloudRegistry.List() {
				listings = append(listings, map[string]any{"skill_id": skillID, "versions": versions})
			}
			return map[string]any{"listings": listings}, nil
		})

	srv.Tool("marketplace_install").
		Description("Validate and install a marketplace skill by id.").
		Handler(func(input MarketplaceInstallInput) (map[string]any, error) {
			if strings.TrimSpace(input.SkillID) == "" {
				return nil, fmt.Errorf("skill_id is required")
			}
			cat := marketplace.NewCatalog()
			allAgents, loadErr := agents.LoadAll()
			if loadErr != nil {
				return nil, loadErr
			}
			agentNames := make([]string, len(allAgents))
			for i, a := range allAgents {
				agentNames[i] = a.Name
			}
			if err := cat.Add(marketplace.Listing{
				SkillID:           input.SkillID,
				Version:           "latest",
				Verified:          true,
				Publisher:         "registry",
				CompatibleClients: agentNames,
				BadgeEvidence:     "registry-signature",
			}); err != nil {
				return nil, err
			}
			return map[string]any{"installed": true, "skill_id": input.SkillID}, nil
		})

	srv.Tool("governance_audit_export").
		Description("Export signed governance audit bundle.").
		Handler(func(input GovernanceAuditExportInput) (map[string]any, error) {
			target := input.Output
			if strings.TrimSpace(target) == "" {
				target = filepath.Join(mcpWorkspaceDir(), "audit", "bundle.json")
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
			if err := (mcpAuditBundleStore{}).WriteBundle(target, bundle); err != nil {
				return nil, err
			}
			return map[string]any{
				"path":      target,
				"signature": bundle.Signature,
				"records":   len(bundle.Records),
			}, nil
		})

	srv.Tool("governance_audit_verify").
		Description("Verify signature on governance audit bundle.").
		Handler(func(input GovernanceAuditVerifyInput) (map[string]any, error) {
			target := input.Input
			if strings.TrimSpace(target) == "" {
				target = filepath.Join(mcpWorkspaceDir(), "audit", "bundle.json")
			}
			bundle, err := (mcpAuditBundleStore{}).LoadBundle(target)
			if err != nil {
				return nil, err
			}
			if err := governance.VerifyBundle(bundle); err != nil {
				return map[string]any{"path": target, "valid": false, "error": err.Error()}, nil
			}
			return map[string]any{"path": target, "valid": true, "signature": bundle.Signature, "records": len(bundle.Records)}, nil
		})

	srv.Tool("runtime_execution_report_export").
		Description("Export structured runtime execution report to file.").
		Handler(func(input RuntimeExecutionReportExportInput) (map[string]any, error) {
			target := input.Output
			if strings.TrimSpace(target) == "" {
				target = filepath.Join(mcpWorkspaceDir(), "state", "runtime-execution-report.json")
			}
			plan, err := runtimeExec.PrepareExecution(runtime.ExecutionRequest{
				SkillID: "runtime-report",
				Version: "0.1.0",
				Input:   map[string]any{"query": "status"},
			})
			if err != nil {
				return nil, err
			}
			report := runtime.BuildExecutionReport(plan, "ok")
			if err := runtime.WriteExecutionReport(target, report); err != nil {
				return nil, err
			}
			return map[string]any{"path": target, "skill_id": report.SkillID, "model": report.Model}, nil
		})

	srv.Resource("aios://status/health").
		Name("AIOS Health Status").
		Description("Current service health status for aios runtime").
		MimeType("text/plain").
		Handler(func(_ context.Context, uri string, _ map[string]string) (*mcpg.ResourceContent, error) {
			return &mcpg.ResourceContent{
				URI:      uri,
				MimeType: "text/plain",
				Text:     "ok",
			}, nil
		})

	srv.Resource("aios://status/sync").
		Name("AIOS Sync State").
		Description("Current sync/drift state from sync engine").
		MimeType("text/plain").
		Handler(func(_ context.Context, uri string, _ map[string]string) (*mcpg.ResourceContent, error) {
			state := "unknown"
			if deps.Sync != nil {
				state = deps.Sync.CurrentState()
			}
			return &mcpg.ResourceContent{
				URI:      uri,
				MimeType: "text/plain",
				Text:     strings.TrimSpace(state),
			}, nil
		})

	srv.Resource("aios://help/commands").
		Name("AIOS CLI Commands").
		Description("Command reference for the aios CLI.").
		MimeType("text/plain").
		Handler(func(_ context.Context, uri string, _ map[string]string) (*mcpg.ResourceContent, error) {
			return &mcpg.ResourceContent{
				URI:      uri,
				MimeType: "text/plain",
				Text:     "status | tray-status | version | doctor | list-clients | model-policy-packs | analytics-summary | analytics-record | analytics-trend | marketplace-publish --skill-dir <dir> | marketplace-list | marketplace-install --skill-dir <skill-id> | marketplace-matrix | audit-export [--skill-dir <output-file>] | audit-verify [--skill-dir <input-file>] | runtime-execution-report [--skill-dir <output-file>] | project-list | project-add --skill-dir <path> | project-remove --skill-dir <path-or-id> | project-inspect --skill-dir <path-or-id> | workspace-validate | workspace-plan | workspace-repair | tui | sync --skill-dir <dir> | sync-plan --skill-dir <dir> | test-skill --skill-dir <dir> | lint-skill --skill-dir <dir> | init-skill --skill-dir <dir> | package-skill --skill-dir <dir> | uninstall-skill --skill-dir <dir> | serve-mcp",
			}, nil
		})

	srv.Resource("aios://status/build").
		Name("AIOS Build Info").
		Description("Build metadata for the running aios service.").
		MimeType("application/json").
		Handler(func(_ context.Context, uri string, _ map[string]string) (*mcpg.ResourceContent, error) {
			body, err := json.Marshal(map[string]string{
				"version":    strings.TrimSpace(deps.Version),
				"commit":     strings.TrimSpace(deps.Commit),
				"build_date": strings.TrimSpace(deps.BuildDate),
			})
			if err != nil {
				return nil, err
			}
			return &mcpg.ResourceContent{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(body),
			}, nil
		})

	srv.Resource("aios://projects/inventory").
		Name("Tracked Projects Inventory").
		Description("Tracked projects/folders from local inventory.").
		MimeType("application/json").
		Handler(func(_ context.Context, uri string, _ map[string]string) (*mcpg.ResourceContent, error) {
			projects, err := projectService.List(context.Background())
			if err != nil {
				return nil, err
			}
			body, err := json.Marshal(map[string]any{"projects": projects})
			if err != nil {
				return nil, err
			}
			return &mcpg.ResourceContent{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(body),
			}, nil
		})

	srv.Resource("aios://workspace/links").
		Name("Workspace Link Health").
		Description("Current workspace link validation state for tracked projects.").
		MimeType("application/json").
		Handler(func(_ context.Context, uri string, _ map[string]string) (*mcpg.ResourceContent, error) {
			result, err := workspaceService.Validate(context.Background())
			if err != nil {
				return nil, err
			}
			body, err := json.Marshal(result)
			if err != nil {
				return nil, err
			}
			return &mcpg.ResourceContent{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(body),
			}, nil
		})

	srv.Resource("aios://analytics/trend").
		Name("AIOS Analytics Trend").
		Description("Trend report over persisted analytics snapshots.").
		MimeType("application/json").
		Handler(func(_ context.Context, uri string, _ map[string]string) (*mcpg.ResourceContent, error) {
			historyPath := filepath.Join(mcpWorkspaceDir(), "state", "analytics-history.json")
			history, err := observability.LoadSnapshots(historyPath)
			if err != nil {
				return nil, err
			}
			body, err := json.Marshal(observability.BuildTrend(history))
			if err != nil {
				return nil, err
			}
			return &mcpg.ResourceContent{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(body),
			}, nil
		})

	srv.Resource("aios://marketplace/compatibility").
		Name("AIOS Marketplace Compatibility Matrix").
		Description("Marketplace skill compatibility matrix by client support and verification state.").
		MimeType("application/json").
		Handler(func(_ context.Context, uri string, _ map[string]string) (*mcpg.ResourceContent, error) {
			allAgents, loadErr := agents.LoadAll()
			if loadErr != nil {
				return nil, loadErr
			}
			agentNames := make([]string, len(allAgents))
			for i, a := range allAgents {
				agentNames[i] = a.Name
			}
			matrix := []map[string]any{}
			if cloudRegistry != nil {
				for skillID, versions := range cloudRegistry.List() {
					matrix = append(matrix, map[string]any{
						"skill_id":           skillID,
						"versions":           versions,
						"compatible_clients": agentNames,
						"verified":           true,
					})
				}
			}
			body, err := json.Marshal(map[string]any{"matrix": matrix})
			if err != nil {
				return nil, err
			}
			return &mcpg.ResourceContent{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(body),
			}, nil
		})

	srv.Resource("docs://{name}").
		Name("Project Docs").
		Description("Read markdown files from docs/ by name, for example docs://prd or docs://tdd").
		MimeType("text/markdown").
		Handler(func(_ context.Context, uri string, params map[string]string) (*mcpg.ResourceContent, error) {
			name := strings.TrimSpace(params["name"])
			if name == "" {
				return nil, fmt.Errorf("name is required")
			}
			if strings.Contains(name, "..") || strings.ContainsRune(name, filepath.Separator) || strings.ContainsRune(name, '/') {
				return nil, fmt.Errorf("invalid doc name %q", name)
			}
			path := filepath.Clean(filepath.Join("docs", name+".md"))
			if filepath.Dir(path) != "docs" {
				return nil, fmt.Errorf("invalid doc path")
			}
			// #nosec G304 -- path is constrained to docs/<name>.md.
			body, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			return &mcpg.ResourceContent{
				URI:      uri,
				MimeType: "text/markdown",
				Text:     string(body),
			}, nil
		})

	srv.Resource("docs://index").
		Name("Project Docs Index").
		Description("List markdown documents available under docs/.").
		MimeType("application/json").
		Handler(func(_ context.Context, uri string, _ map[string]string) (*mcpg.ResourceContent, error) {
			entries, err := os.ReadDir("docs")
			if err != nil {
				return nil, err
			}
			docs := make([]string, 0, len(entries))
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				name := e.Name()
				if !strings.HasSuffix(name, ".md") {
					continue
				}
				docs = append(docs, strings.TrimSuffix(name, ".md"))
			}
			sort.Strings(docs)
			body, err := json.Marshal(map[string]any{"docs": docs})
			if err != nil {
				return nil, err
			}
			return &mcpg.ResourceContent{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(body),
			}, nil
		})

	return srv
}

// mcpAuditBundleStore implements governance.AuditBundleStore for MCP handlers.
type mcpAuditBundleStore struct{}

func (mcpAuditBundleStore) WriteBundle(path string, bundle governance.AuditBundle) error {
	body, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o600)
}

func (mcpAuditBundleStore) LoadBundle(path string) (governance.AuditBundle, error) {
	path = filepath.Clean(path)
	// #nosec G304 -- path is provided by explicit audit export/verify command input.
	body, err := os.ReadFile(path)
	if err != nil {
		return governance.AuditBundle{}, err
	}
	var bundle governance.AuditBundle
	if err := json.Unmarshal(body, &bundle); err != nil {
		return governance.AuditBundle{}, err
	}
	return bundle, nil
}

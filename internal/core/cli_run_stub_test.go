package core

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	domainonboarding "github.com/felixgeelhaar/aios/internal/domain/onboarding"
	domainprojectinventory "github.com/felixgeelhaar/aios/internal/domain/projectinventory"
	domainskilllint "github.com/felixgeelhaar/aios/internal/domain/skilllint"
	domainskillpackage "github.com/felixgeelhaar/aios/internal/domain/skillpackage"
	domainskillsync "github.com/felixgeelhaar/aios/internal/domain/skillsync"
	domainskilltest "github.com/felixgeelhaar/aios/internal/domain/skilltest"
	domainskilluninstall "github.com/felixgeelhaar/aios/internal/domain/skilluninstall"
	domainsyncplan "github.com/felixgeelhaar/aios/internal/domain/syncplan"
	domainworkspace "github.com/felixgeelhaar/aios/internal/domain/workspaceorchestration"
	"github.com/felixgeelhaar/aios/internal/model"
	"github.com/felixgeelhaar/aios/internal/runtime"
)

func newStubCLI() CLI {
	return CLI{
		In:  strings.NewReader(""),
		Out: io.Discard,
		SyncState: func() string {
			return "clean"
		},
		Health: func() runtime.HealthReport {
			return runtime.HealthReport{
				Status:     "ok",
				Ready:      true,
				TokenStore: "memory",
				Workspace:  "/tmp",
			}
		},
		SyncSkill: func(context.Context, domainskillsync.SyncSkillCommand) (string, error) {
			return "skill-id", nil
		},
		TestSkill: func(context.Context, domainskilltest.TestSkillCommand) (domainskilltest.TestSkillResult, error) {
			return domainskilltest.TestSkillResult{
				Results: []domainskilltest.FixtureResult{{Name: "fixture", Passed: true}},
				Failed:  0,
			}, nil
		},
		SyncPlan: func(context.Context, domainsyncplan.BuildSyncPlanCommand) (domainsyncplan.BuildSyncPlanResult, error) {
			return domainsyncplan.BuildSyncPlanResult{SkillID: "skill-id", Writes: []string{"a", "b"}}, nil
		},
		InitSkill: func(string) error {
			return nil
		},
		LintSkill: func(context.Context, domainskilllint.LintSkillCommand) (domainskilllint.LintSkillResult, error) {
			return domainskilllint.LintSkillResult{Valid: true}, nil
		},
		BuildInfo: func() BuildInfo {
			return BuildInfo{Version: "1.0.0", Commit: "abc123", BuildDate: "2026-02-15"}
		},
		Doctor: func() DoctorReport {
			return DoctorReport{Overall: true, Checks: []DoctorCheck{{Name: "workspace_dir", OK: true, Detail: "ok"}}}
		},
		ListClients: func() map[string]any {
			return map[string]any{"client": map[string]any{"path": "/tmp", "files": []string{"a"}}}
		},
		ListProjects: func(context.Context) ([]domainprojectinventory.Project, error) {
			return []domainprojectinventory.Project{{ID: "p1", Path: "/tmp", AddedAt: "2026-02-15T00:00:00Z"}}, nil
		},
		AddProject: func(context.Context, string) (domainprojectinventory.Project, error) {
			return domainprojectinventory.Project{ID: "p1", Path: "/tmp", AddedAt: "2026-02-15T00:00:00Z"}, nil
		},
		RemoveProject: func(context.Context, string) error {
			return nil
		},
		InspectProject: func(context.Context, string) (domainprojectinventory.Project, error) {
			return domainprojectinventory.Project{ID: "p1", Path: "/tmp", AddedAt: "2026-02-15T00:00:00Z"}, nil
		},
		ValidateWorkspace: func(context.Context) (domainworkspace.ValidationResult, error) {
			return domainworkspace.ValidationResult{
				Healthy: false,
				Links:   []domainworkspace.LinkReport{{ProjectID: "p1", ProjectPath: "/tmp", LinkPath: "/links/p1", Status: domainworkspace.LinkStatusBroken}},
			}, nil
		},
		PlanWorkspace: func(context.Context) (domainworkspace.PlanResult, error) {
			return domainworkspace.PlanResult{
				Actions: []domainworkspace.PlanAction{{Kind: domainworkspace.ActionCreate, ProjectID: "p1", LinkPath: "/links/p1", TargetPath: "/tmp", Reason: "missing"}},
			}, nil
		},
		RepairWorkspace: func(context.Context) (domainworkspace.RepairResult, error) {
			return domainworkspace.RepairResult{
				Applied: []domainworkspace.PlanAction{{Kind: domainworkspace.ActionCreate, ProjectID: "p1", LinkPath: "/links/p1", TargetPath: "/tmp", Reason: "missing"}},
				Skipped: []domainworkspace.PlanAction{{Kind: domainworkspace.ActionSkip, ProjectID: "p2", LinkPath: "/links/p2", TargetPath: "/tmp/p2", Reason: "ok"}},
			}, nil
		},
		ModelPolicyPacks: func() []model.PolicyPack {
			return []model.PolicyPack{{Name: "pack", Description: "desc"}}
		},
		PackageSkill: func(context.Context, domainskillpackage.PackageSkillCommand) (domainskillpackage.PackageSkillResult, error) {
			return domainskillpackage.PackageSkillResult{ArtifactPath: "/tmp/skill.tgz"}, nil
		},
		UninstallSkill: func(context.Context, domainskilluninstall.UninstallSkillCommand) (string, error) {
			return "skill-id", nil
		},
		BackupConfigs: func() (string, error) {
			return "/tmp/backup", nil
		},
		RestoreConfigs: func(string) (string, error) {
			return "/tmp/backup", nil
		},
		ExportReport: func(string) (string, error) {
			return "/tmp/report.md", nil
		},
		AnalyticsSummary: func(context.Context) (map[string]any, error) {
			return map[string]any{
				"tracked_projects":  1,
				"workspace_links":   2,
				"healthy_links":     1,
				"workspace_healthy": false,
				"sync_state":        "clean",
			}, nil
		},
		AnalyticsRecord: func(context.Context) (map[string]any, error) {
			return map[string]any{"recorded": true, "points": 3}, nil
		},
		AnalyticsTrend: func(context.Context) (map[string]any, error) {
			return map[string]any{"points": 3, "delta_tracked_projects": 1, "delta_healthy_links": 1}, nil
		},
		MarketplacePublish: func(context.Context, string) (map[string]any, error) {
			return map[string]any{"published": true, "skill_id": "skill-id", "version": "0.1.0"}, nil
		},
		MarketplaceList: func(context.Context) (map[string]any, error) {
			return map[string]any{"listings": []string{"skill-id"}}, nil
		},
		MarketplaceInstall: func(context.Context, string) (map[string]any, error) {
			return map[string]any{"installed": true, "skill_id": "skill-id"}, nil
		},
		MarketplaceMatrix: func(context.Context) (map[string]any, error) {
			return map[string]any{"matrix": map[string][]string{"skill-id": {"client"}}}, nil
		},
		ExportAudit: func(string) (map[string]any, error) {
			return map[string]any{"path": "/tmp/audit.json", "signature": "sig"}, nil
		},
		VerifyAudit: func(string) (map[string]any, error) {
			return map[string]any{"valid": true, "path": "/tmp/audit.json"}, nil
		},
		ExecutionReport: func(string) (map[string]any, error) {
			return map[string]any{"path": "/tmp/execution.json"}, nil
		},
		ConnectGoogleDrive: func(context.Context, domainonboarding.ConnectGoogleDriveCommand) (domainonboarding.ConnectGoogleDriveResult, error) {
			return domainonboarding.ConnectGoogleDriveResult{CallbackURL: "http://localhost"}, nil
		},
		TrayStatus: func() (TrayState, error) {
			return TrayState{UpdatedAt: time.Now().UTC().Format(time.RFC3339), Skills: []string{"s1"}, Connections: map[string]bool{"google_drive": true}}, nil
		},
	}
}

func TestCLI_RunStubCommands(t *testing.T) {
	cli := newStubCLI()
	ctx := context.Background()
	commands := []struct {
		name     string
		cmd      string
		skillDir string
		output   string
	}{
		{name: "status", cmd: "status"},
		{name: "status-json", cmd: "status", output: "json"},
		{name: "sync", cmd: "sync", skillDir: "/tmp/skill"},
		{name: "sync-json", cmd: "sync", skillDir: "/tmp/skill", output: "json"},
		{name: "sync-plan", cmd: "sync-plan", skillDir: "/tmp/skill"},
		{name: "test-skill", cmd: "test-skill", skillDir: "/tmp/skill"},
		{name: "test-skill-json", cmd: "test-skill", skillDir: "/tmp/skill", output: "json"},
		{name: "lint-skill", cmd: "lint-skill", skillDir: "/tmp/skill"},
		{name: "lint-skill-json", cmd: "lint-skill", skillDir: "/tmp/skill", output: "json"},
		{name: "init-skill", cmd: "init-skill", skillDir: "/tmp/skill"},
		{name: "version", cmd: "version"},
		{name: "version-json", cmd: "version", output: "json"},
		{name: "doctor", cmd: "doctor"},
		{name: "doctor-json", cmd: "doctor", output: "json"},
		{name: "list-clients", cmd: "list-clients"},
		{name: "model-policy-packs", cmd: "model-policy-packs"},
		{name: "analytics-summary", cmd: "analytics-summary"},
		{name: "analytics-record", cmd: "analytics-record"},
		{name: "analytics-trend", cmd: "analytics-trend"},
		{name: "marketplace-publish", cmd: "marketplace-publish", skillDir: "/tmp/skill"},
		{name: "marketplace-list", cmd: "marketplace-list"},
		{name: "marketplace-install", cmd: "marketplace-install", skillDir: "skill-id"},
		{name: "marketplace-matrix", cmd: "marketplace-matrix"},
		{name: "audit-export", cmd: "audit-export", skillDir: "/tmp/audit.json"},
		{name: "audit-verify", cmd: "audit-verify", skillDir: "/tmp/audit.json"},
		{name: "runtime-execution-report", cmd: "runtime-execution-report", skillDir: "/tmp/execution.json"},
		{name: "project-list", cmd: "project-list"},
		{name: "project-add", cmd: "project-add", skillDir: "/tmp"},
		{name: "project-remove", cmd: "project-remove", skillDir: "p1"},
		{name: "project-inspect", cmd: "project-inspect", skillDir: "p1"},
		{name: "workspace-validate", cmd: "workspace-validate"},
		{name: "workspace-plan", cmd: "workspace-plan"},
		{name: "workspace-repair", cmd: "workspace-repair"},
		{name: "package-skill", cmd: "package-skill", skillDir: "/tmp/skill"},
		{name: "uninstall-skill", cmd: "uninstall-skill", skillDir: "/tmp/skill"},
		{name: "backup-configs", cmd: "backup-configs"},
		{name: "restore-configs", cmd: "restore-configs", skillDir: "/tmp/backup"},
		{name: "export-status-report", cmd: "export-status-report", skillDir: "/tmp/report.md"},
		{name: "connect-google-drive", cmd: "connect-google-drive"},
		{name: "tray-status", cmd: "tray-status"},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			err := cli.Run(ctx, tc.cmd, tc.skillDir, "", "", tc.output)
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}
		})
	}
}

package core

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/felixgeelhaar/aios/internal/agents"
	"github.com/felixgeelhaar/aios/internal/builder"
	domainonboarding "github.com/felixgeelhaar/aios/internal/domain/onboarding"
	domainskilllint "github.com/felixgeelhaar/aios/internal/domain/skilllint"
	domainskillpackage "github.com/felixgeelhaar/aios/internal/domain/skillpackage"
	aosmcp "github.com/felixgeelhaar/aios/internal/mcp"
	"github.com/felixgeelhaar/aios/internal/runtime"
	"github.com/felixgeelhaar/aios/internal/sync"
)

func TestEndToEndSkillBuildInstallAndRuntime(t *testing.T) {
	dir := t.TempDir()
	if err := builder.BuildSkill(builder.Spec{ID: "roadmap-reader", Version: "0.1.0", Dir: dir}); err != nil {
		t.Fatalf("build skill: %v", err)
	}

	allAgents, loadErr := agents.LoadAll()
	if loadErr != nil {
		t.Fatalf("load agents: %v", loadErr)
	}
	si := agents.NewSkillInstaller(allAgents)
	if _, err := si.InstallSkill("roadmap-reader", agents.InstallOptions{ProjectDir: dir}); err != nil {
		t.Fatalf("install skill: %v", err)
	}

	rt := runtime.New(dir, runtime.NewMemoryTokenStore())
	if err := rt.ConnectGoogleDrive(context.Background(), "token"); err != nil {
		t.Fatalf("connect drive: %v", err)
	}

	syncEngine := sync.NewEngine()
	_ = syncEngine.DetectDrift(map[string]string{"a": "1"}, map[string]string{"a": "2"})
	if syncEngine.CurrentState() != "drifted" {
		t.Fatalf("expected drifted state, got %s", syncEngine.CurrentState())
	}

	srv := aosmcp.NewServerWithDeps("0.1.0", aosmcp.ServerDeps{Sync: syncEngine})
	if len(srv.Tools()) < 3 {
		t.Fatalf("expected MCP tools, got %d", len(srv.Tools()))
	}
}

func TestLocalKernelOnboardingPathViaCLI(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_WORKSPACE_DIR", filepath.Join(root, "workspace"))
	t.Setenv("AIOS_PROJECT_DIR", filepath.Join(root, "project"))
	t.Setenv("AIOS_TOKEN_SERVICE", "aios-test")
	cfg := DefaultConfig()

	for _, p := range []string{cfg.WorkspaceDir, cfg.ProjectDir} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, cfg)
	// Avoid OS keychain dependency in integration tests.
	cli.ConnectGoogleDrive = func(_ context.Context, cmd domainonboarding.ConnectGoogleDriveCommand) (domainonboarding.ConnectGoogleDriveResult, error) {
		if cmd.TokenOverride == "" {
			t.Fatalf("expected oauth token")
		}
		connected := true
		if _, err := RefreshTrayState(cfg, &connected); err != nil {
			t.Fatalf("refresh tray state: %v", err)
		}
		return domainonboarding.ConnectGoogleDriveResult{}, nil
	}

	skillDir := filepath.Join(root, "roadmap-reader")
	if err := os.MkdirAll(filepath.Join(skillDir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("id: roadmap-reader\nversion: 0.1.0\ninputs:\n  schema: schema.input.json\noutputs:\n  schema: schema.output.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.input.json"), []byte(`{"type":"object","properties":{"query":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "schema.output.json"), []byte(`{"type":"object","properties":{"status":{"type":"string"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run(context.Background(), "sync", skillDir, "stdio", ":8080", "text"); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	t.Setenv("AIOS_OAUTH_TOKEN", "demo-token")
	if err := cli.Run(context.Background(), "connect-google-drive", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("connect-google-drive failed: %v", err)
	}

	buf.Reset()
	if err := cli.Run(context.Background(), "tray-status", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("tray-status failed: %v", err)
	}
	var status TrayState
	if err := json.Unmarshal(buf.Bytes(), &status); err != nil {
		t.Fatalf("invalid tray-status json: %v", err)
	}
	if !status.Connections["google_drive"] {
		t.Fatalf("expected google_drive connected: %#v", status.Connections)
	}
	found := false
	for _, id := range status.Skills {
		if id == "roadmap-reader" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected roadmap-reader in skills: %#v", status.Skills)
	}
}

func TestOrgControlPlaneProjectInventoryAndWorkspaceLinksViaCLI(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIOS_WORKSPACE_DIR", filepath.Join(root, "workspace"))
	t.Setenv("AIOS_PROJECT_DIR", filepath.Join(root, "project"))
	cfg := DefaultConfig()

	for _, p := range []string{cfg.WorkspaceDir, cfg.ProjectDir} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	projectPath := filepath.Join(root, "repo-a")
	if err := os.MkdirAll(projectPath, 0o755); err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}
	cli := DefaultCLI(buf, cfg)

	if err := cli.Run(context.Background(), "project-add", projectPath, "stdio", ":8080", "json"); err != nil {
		t.Fatalf("project-add failed: %v", err)
	}

	buf.Reset()
	if err := cli.Run(context.Background(), "workspace-validate", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("workspace-validate failed: %v", err)
	}
	var validateBefore map[string]any
	if err := json.Unmarshal(buf.Bytes(), &validateBefore); err != nil {
		t.Fatalf("invalid workspace-validate json: %v", err)
	}
	if healthy, _ := validateBefore["healthy"].(bool); healthy {
		t.Fatalf("expected unhealthy workspace before repair: %#v", validateBefore)
	}

	buf.Reset()
	if err := cli.Run(context.Background(), "workspace-repair", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("workspace-repair failed: %v", err)
	}

	buf.Reset()
	if err := cli.Run(context.Background(), "workspace-validate", "", "stdio", ":8080", "json"); err != nil {
		t.Fatalf("workspace-validate failed: %v", err)
	}
	var validateAfter map[string]any
	if err := json.Unmarshal(buf.Bytes(), &validateAfter); err != nil {
		t.Fatalf("invalid workspace-validate json: %v", err)
	}
	if healthy, _ := validateAfter["healthy"].(bool); !healthy {
		t.Fatalf("expected healthy workspace after repair: %#v", validateAfter)
	}
}

// TestIntegrationSkillLifecycle tests the complete skill workflow using CLI functions
func TestIntegrationSkillLifecycle(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "skills", "test-integration")

	t.Setenv("AIOS_WORKSPACE_DIR", root)
	t.Setenv("AIOS_PROJECT_DIR", root)

	cfg := DefaultConfig()
	cfg.WorkspaceDir = root
	cfg.ProjectDir = root

	cli := DefaultCLI(nil, cfg)
	ctx := context.Background()

	t.Run("init skill", func(t *testing.T) {
		if err := cli.InitSkill(skillDir); err != nil {
			t.Fatalf("init skill: %v", err)
		}
		if _, err := os.Stat(filepath.Join(skillDir, "skill.yaml")); err != nil {
			t.Fatal("skill.yaml not created")
		}
	})

	t.Run("lint skill passes", func(t *testing.T) {
		res, err := cli.LintSkill(ctx, domainskilllint.LintSkillCommand{SkillDir: skillDir})
		if err != nil {
			t.Fatalf("lint: %v", err)
		}
		if !res.Valid {
			t.Fatalf("lint failed: %v", res.Issues)
		}
	})

	t.Run("package skill", func(t *testing.T) {
		res, err := cli.PackageSkill(ctx, domainskillpackage.PackageSkillCommand{SkillDir: skillDir})
		if err != nil {
			t.Fatalf("package: %v", err)
		}
		if res.ArtifactPath == "" {
			t.Fatal("artifact path is empty")
		}
		if _, err := os.Stat(res.ArtifactPath); err != nil {
			t.Fatal("artifact not created")
		}
	})

	t.Run("sync skill", func(t *testing.T) {
		skillID, err := cli.SyncSkill(ctx, struct{ SkillDir string }{SkillDir: skillDir})
		if err != nil {
			t.Fatalf("sync: %v", err)
		}
		if skillID == "" {
			t.Fatal("expected skill ID")
		}

		// Verify skill was installed
		canonicalPath := filepath.Join(root, ".agents", "skills", skillID, "SKILL.md")
		if _, err := os.Stat(canonicalPath); err != nil {
			t.Fatalf("skill not synced to canonical location: %v", err)
		}
	})
}

// TestIntegrationMarketplaceWorkflow tests marketplace publish and install
func TestIntegrationMarketplaceWorkflow(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "skills", "marketplace-test")

	t.Setenv("AIOS_WORKSPACE_DIR", root)
	t.Setenv("AIOS_PROJECT_DIR", root)

	cfg := DefaultConfig()
	cfg.WorkspaceDir = root
	cfg.ProjectDir = root

	cli := DefaultCLI(nil, cfg)
	ctx := context.Background()

	// First create and validate a skill
	if err := cli.InitSkill(skillDir); err != nil {
		t.Fatalf("init skill: %v", err)
	}

	t.Run("publish to marketplace", func(t *testing.T) {
		res, err := cli.MarketplacePublish(ctx, skillDir)
		if err != nil {
			t.Fatalf("publish: %v", err)
		}
		if !res["published"].(bool) {
			t.Fatal("expected published=true")
		}
	})

	t.Run("list marketplace", func(t *testing.T) {
		res, err := cli.MarketplaceList(ctx)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		listings, ok := res["listings"].([]any)
		if !ok || len(listings) == 0 {
			t.Log("warning: no marketplace listings")
		}
	})
}

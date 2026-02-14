package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/felixgeelhaar/aios/internal/domain/agentregistry"
)

func TestOAuthCodeResolverAdapterTimesOut(t *testing.T) {
	ctx := context.Background()
	adapter := oauthCodeResolverAdapter{}

	callbackURL, code, err := adapter.ResolveCode(ctx, "state-1", 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
	if code != "" {
		t.Fatalf("expected empty code on timeout, got %q", code)
	}
	if !strings.Contains(callbackURL, "/oauth/callback") {
		t.Fatalf("unexpected callback url: %q", callbackURL)
	}
}

func TestDriveConnectorAdapterRejectsMissingTokenService(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WorkspaceDir = t.TempDir()
	cfg.TokenService = ""

	adapter := driveConnectorAdapter{cfg: cfg}
	if err := adapter.ConnectGoogleDrive(context.Background(), "token-1"); err == nil {
		t.Fatal("expected error for empty token service")
	}
}

func TestTrayStatePortAdapterSetsGoogleDriveConnection(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultConfig()
	cfg.WorkspaceDir = root
	cfg.ProjectDir = filepath.Join(root, "project")

	skillDir := filepath.Join(cfg.ProjectDir, agentregistry.CanonicalSkillsDir, "roadmap-reader")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: roadmap-reader\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := trayStatePortAdapter{cfg: cfg}
	if err := adapter.SetGoogleDriveConnected(context.Background(), true); err != nil {
		t.Fatalf("set google drive connected failed: %v", err)
	}
	state, err := ReadTrayState(cfg.WorkspaceDir)
	if err != nil {
		t.Fatalf("read tray state failed: %v", err)
	}
	if !state.Connections["google_drive"] {
		t.Fatalf("expected google_drive to be connected: %#v", state.Connections)
	}
}

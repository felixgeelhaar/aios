package onboarding

import (
	"testing"
	"time"
)

func TestNormalizedDefaultsState(t *testing.T) {
	cmd := ConnectGoogleDriveCommand{}.Normalized()
	if cmd.State != DefaultOAuthState {
		t.Fatalf("expected default state %q, got %q", DefaultOAuthState, cmd.State)
	}
}

func TestNormalizedDefaultsTimeout(t *testing.T) {
	cmd := ConnectGoogleDriveCommand{}.Normalized()
	if cmd.Timeout != DefaultOAuthTimeout {
		t.Fatalf("expected default timeout %v, got %v", DefaultOAuthTimeout, cmd.Timeout)
	}
	if cmd.Timeout != 120*time.Second {
		t.Fatalf("expected 120s default, got %v", cmd.Timeout)
	}
}

func TestNormalizedPreservesExplicitTimeout(t *testing.T) {
	cmd := ConnectGoogleDriveCommand{Timeout: 30 * time.Second}.Normalized()
	if cmd.Timeout != 30*time.Second {
		t.Fatalf("expected 30s, got %v", cmd.Timeout)
	}
}

func TestNormalizedTrimsTokenOverride(t *testing.T) {
	cmd := ConnectGoogleDriveCommand{TokenOverride: "  tok-1  "}.Normalized()
	if cmd.TokenOverride != "tok-1" {
		t.Fatalf("expected trimmed token, got %q", cmd.TokenOverride)
	}
}

func TestNormalizedTrimsState(t *testing.T) {
	cmd := ConnectGoogleDriveCommand{State: "  custom  "}.Normalized()
	if cmd.State != "custom" {
		t.Fatalf("expected trimmed state, got %q", cmd.State)
	}
}

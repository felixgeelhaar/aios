package onboarding

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	domain "github.com/felixgeelhaar/aios/internal/domain/onboarding"
)

type fakeOAuthResolver struct {
	callbackURL  string
	code         string
	err          error
	gotState     string
	gotTimeout   time.Duration
	resolveCalls int
}

func (f *fakeOAuthResolver) ResolveCode(_ context.Context, state string, timeout time.Duration) (string, string, error) {
	f.resolveCalls++
	f.gotState = state
	f.gotTimeout = timeout
	return f.callbackURL, f.code, f.err
}

type fakeConnector struct {
	token string
	err   error
}

func (f *fakeConnector) ConnectGoogleDrive(_ context.Context, token string) error {
	f.token = token
	return f.err
}

type fakeTrayPort struct {
	connected bool
	err       error
}

func (f *fakeTrayPort) SetGoogleDriveConnected(_ context.Context, connected bool) error {
	f.connected = connected
	return f.err
}

func TestServiceConnectGoogleDriveUsesTokenOverride(t *testing.T) {
	resolver := &fakeOAuthResolver{}
	connector := &fakeConnector{}
	tray := &fakeTrayPort{}
	svc := NewService(resolver, connector, tray)

	_, err := svc.ConnectGoogleDrive(context.Background(), domain.ConnectGoogleDriveCommand{TokenOverride: "token-1"})
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	if connector.token != "token-1" {
		t.Fatalf("unexpected token: %q", connector.token)
	}
	if !tray.connected {
		t.Fatal("expected tray connection update")
	}
	if resolver.resolveCalls != 0 {
		t.Fatal("oauth resolver should not be called when token override is present")
	}
}

func TestServiceConnectGoogleDriveResolvesOAuthWhenNoToken(t *testing.T) {
	resolver := &fakeOAuthResolver{callbackURL: "http://127.0.0.1:9999/oauth/callback", code: "oauth-code"}
	connector := &fakeConnector{}
	tray := &fakeTrayPort{}
	svc := NewService(resolver, connector, tray)

	res, err := svc.ConnectGoogleDrive(context.Background(), domain.ConnectGoogleDriveCommand{})
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	if connector.token != "oauth-code" {
		t.Fatalf("unexpected token: %q", connector.token)
	}
	if res.CallbackURL == "" {
		t.Fatalf("expected callback url in result: %#v", res)
	}
}

func TestServiceConnectGoogleDrivePropagatesErrors(t *testing.T) {
	expected := errors.New("store failed")
	connector := &fakeConnector{err: expected}
	tray := &fakeTrayPort{}
	svc := NewService(&fakeOAuthResolver{}, connector, tray)

	_, err := svc.ConnectGoogleDrive(context.Background(), domain.ConnectGoogleDriveCommand{TokenOverride: "token-1"})
	if !errors.Is(err, expected) {
		t.Fatalf("expected connector error, got: %v", err)
	}
}

// AC2: Service passes normalized timeout to OAuthCodeResolver.
func TestServicePassesTimeoutToResolver(t *testing.T) {
	resolver := &fakeOAuthResolver{callbackURL: "http://127.0.0.1:0/oauth/callback", code: "code-1"}
	connector := &fakeConnector{}
	tray := &fakeTrayPort{}
	svc := NewService(resolver, connector, tray)

	_, err := svc.ConnectGoogleDrive(context.Background(), domain.ConnectGoogleDriveCommand{
		Timeout: 45 * time.Second,
	})
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	if resolver.gotTimeout != 45*time.Second {
		t.Fatalf("expected 45s timeout passed to resolver, got %v", resolver.gotTimeout)
	}
}

// AC2: Default timeout is 120s when not specified.
func TestServiceUsesDefaultTimeoutWhenZero(t *testing.T) {
	resolver := &fakeOAuthResolver{callbackURL: "http://127.0.0.1:0/oauth/callback", code: "code-1"}
	connector := &fakeConnector{}
	tray := &fakeTrayPort{}
	svc := NewService(resolver, connector, tray)

	_, err := svc.ConnectGoogleDrive(context.Background(), domain.ConnectGoogleDriveCommand{})
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	if resolver.gotTimeout != domain.DefaultOAuthTimeout {
		t.Fatalf("expected default timeout %v, got %v", domain.DefaultOAuthTimeout, resolver.gotTimeout)
	}
}

// AC1: State parameter is forwarded to OAuthCodeResolver.
func TestServicePassesStateToResolver(t *testing.T) {
	resolver := &fakeOAuthResolver{callbackURL: "http://127.0.0.1:0/oauth/callback", code: "code-1"}
	connector := &fakeConnector{}
	tray := &fakeTrayPort{}
	svc := NewService(resolver, connector, tray)

	_, err := svc.ConnectGoogleDrive(context.Background(), domain.ConnectGoogleDriveCommand{
		State: "custom-state",
	})
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	if resolver.gotState != "custom-state" {
		t.Fatalf("expected state %q, got %q", "custom-state", resolver.gotState)
	}
}

// AC4: Result does not contain the token value — token is never exposed.
func TestServiceResultDoesNotLeakToken(t *testing.T) {
	resolver := &fakeOAuthResolver{callbackURL: "http://127.0.0.1:0/oauth/callback", code: "secret-code-xyz"}
	connector := &fakeConnector{}
	tray := &fakeTrayPort{}
	svc := NewService(resolver, connector, tray)

	result, err := svc.ConnectGoogleDrive(context.Background(), domain.ConnectGoogleDriveCommand{})
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	if strings.Contains(result.CallbackURL, "secret-code-xyz") {
		t.Fatal("callback URL must not contain the token/code value")
	}
}

// AC4: Token override is never returned in the result either.
func TestServiceResultDoesNotLeakTokenOverride(t *testing.T) {
	connector := &fakeConnector{}
	tray := &fakeTrayPort{}
	svc := NewService(&fakeOAuthResolver{}, connector, tray)

	result, err := svc.ConnectGoogleDrive(context.Background(), domain.ConnectGoogleDriveCommand{TokenOverride: "my-secret-token"})
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	if result.CallbackURL != "" {
		t.Fatal("callback URL should be empty when using token override")
	}
}

// AC3: Empty token override falls through to OAuth resolver.
func TestServiceEmptyTokenOverrideFallsToOAuth(t *testing.T) {
	resolver := &fakeOAuthResolver{callbackURL: "http://127.0.0.1:0/oauth/callback", code: "resolved-code"}
	connector := &fakeConnector{}
	tray := &fakeTrayPort{}
	svc := NewService(resolver, connector, tray)

	_, err := svc.ConnectGoogleDrive(context.Background(), domain.ConnectGoogleDriveCommand{TokenOverride: "   "})
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	if resolver.resolveCalls != 1 {
		t.Fatalf("expected oauth resolver to be called, calls=%d", resolver.resolveCalls)
	}
	if connector.token != "resolved-code" {
		t.Fatalf("expected resolved code, got %q", connector.token)
	}
}

// ErrTokenRequired: Resolver returns empty code => domain error.
func TestServiceRejectsEmptyCodeFromResolver(t *testing.T) {
	resolver := &fakeOAuthResolver{callbackURL: "http://127.0.0.1:0/oauth/callback", code: ""}
	connector := &fakeConnector{}
	tray := &fakeTrayPort{}
	svc := NewService(resolver, connector, tray)

	_, err := svc.ConnectGoogleDrive(context.Background(), domain.ConnectGoogleDriveCommand{})
	if !errors.Is(err, domain.ErrTokenRequired) {
		t.Fatalf("expected ErrTokenRequired, got: %v", err)
	}
}

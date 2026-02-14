package runtime

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestOAuthCallbackServerSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url, resultCh, stop, err := StartOAuthCallbackServer(ctx, "", "ok-state")
	if err != nil {
		t.Fatalf("start callback server: %v", err)
	}
	defer func() { _ = stop() }()

	resp, err := http.Get(url + "?state=ok-state&code=abc123")
	if err != nil {
		t.Fatalf("call callback: %v", err)
	}
	_ = resp.Body.Close()

	code, err := WaitForOAuthCode(context.Background(), resultCh)
	if err != nil {
		t.Fatalf("wait for code: %v", err)
	}
	if code != "abc123" {
		t.Fatalf("unexpected code: %q", code)
	}
}

func TestOAuthCallbackServerInvalidState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url, resultCh, stop, err := StartOAuthCallbackServer(ctx, "", "expected")
	if err != nil {
		t.Fatalf("start callback server: %v", err)
	}
	defer func() { _ = stop() }()

	resp, err := http.Get(url + "?state=wrong&code=abc123")
	if err != nil {
		t.Fatalf("call callback: %v", err)
	}
	_ = resp.Body.Close()

	_, err = WaitForOAuthCode(context.Background(), resultCh)
	if err == nil {
		t.Fatal("expected state validation error")
	}
}

func TestWaitForOAuthCodeTimesOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	ch := make(chan OAuthCallbackResult)
	_, err := WaitForOAuthCode(ctx, ch)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

// AC1: Invalid state returns HTTP 400 to the caller.
func TestOAuthCallbackServerInvalidStateReturns400(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url, _, stop, err := StartOAuthCallbackServer(ctx, "", "valid-state")
	if err != nil {
		t.Fatalf("start callback server: %v", err)
	}
	defer func() { _ = stop() }()

	resp, err := http.Get(url + "?state=bad-state&code=abc123")
	if err != nil {
		t.Fatalf("call callback: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid state, got %d", resp.StatusCode)
	}
}

// AC1: Missing code returns HTTP 400.
func TestOAuthCallbackServerMissingCodeReturns400(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url, _, stop, err := StartOAuthCallbackServer(ctx, "", "s1")
	if err != nil {
		t.Fatalf("start callback server: %v", err)
	}
	defer func() { _ = stop() }()

	resp, err := http.Get(url + "?state=s1")
	if err != nil {
		t.Fatalf("call callback: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing code, got %d", resp.StatusCode)
	}
}

// AC1: Error parameter in callback returns HTTP 400.
func TestOAuthCallbackServerErrorParamReturns400(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url, _, stop, err := StartOAuthCallbackServer(ctx, "", "s1")
	if err != nil {
		t.Fatalf("start callback server: %v", err)
	}
	defer func() { _ = stop() }()

	resp, err := http.Get(url + "?state=s1&error=access_denied")
	if err != nil {
		t.Fatalf("call callback: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for error param, got %d", resp.StatusCode)
	}
}

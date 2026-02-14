package runtime

import (
	"testing"
)

func TestHealthReport(t *testing.T) {
	r := New("/tmp/aios", NewMemoryTokenStore())
	h := r.Health()
	if h.Status != "ok" || !h.Ready {
		t.Fatalf("unexpected health: %#v", h)
	}
	if h.TokenStore != "memory" {
		t.Fatalf("unexpected token store: %s", h.TokenStore)
	}
}

// AC1: Health report must include runtime status and workspace.
func TestHealthReportIncludesWorkspace(t *testing.T) {
	r := New("/custom/workspace", NewMemoryTokenStore())
	h := r.Health()
	if h.Workspace != "/custom/workspace" {
		t.Fatalf("expected workspace '/custom/workspace', got %q", h.Workspace)
	}
	if h.Status != "ok" {
		t.Fatalf("expected status 'ok', got %q", h.Status)
	}
}

// AC1: Health report shows not ready when workspace is empty.
func TestHealthReportNotReadyWithoutWorkspace(t *testing.T) {
	r := New("", NewMemoryTokenStore())
	h := r.Health()
	if h.Ready {
		t.Fatal("expected ready=false when workspace is empty")
	}
}

// AC1: Health report shows keychain token store type for production runtime.
func TestHealthReportKeychainTokenStoreType(t *testing.T) {
	r, err := NewProductionRuntime(t.TempDir(), "aios")
	if err != nil {
		t.Fatalf("new production runtime: %v", err)
	}
	h := r.Health()
	if h.TokenStore != "keychain" {
		t.Fatalf("expected token store 'keychain', got %q", h.TokenStore)
	}
}

// AC1: Health report CheckedAt is populated.
func TestHealthReportCheckedAtPopulated(t *testing.T) {
	r := New("/tmp/aios", NewMemoryTokenStore())
	h := r.Health()
	if h.CheckedAt.IsZero() {
		t.Fatal("expected CheckedAt to be set")
	}
}

package runtime

import "testing"

func TestNewProductionRuntimeUsesKeychainStore(t *testing.T) {
	r, err := NewProductionRuntime(t.TempDir(), "aios")
	if err != nil {
		t.Fatalf("new production runtime: %v", err)
	}
	if !r.UsesSecureTokenStore() {
		t.Fatal("expected secure token store")
	}
}

func TestNewProductionRuntimeRejectsEmptyKeychainService(t *testing.T) {
	_, err := NewProductionRuntime(t.TempDir(), "")
	if err == nil {
		t.Fatal("expected error for empty keychain service")
	}
}

func TestMemoryTokenStoreIsNotSecure(t *testing.T) {
	store := NewMemoryTokenStore()
	r := New(t.TempDir(), store)
	if r.UsesSecureTokenStore() {
		t.Fatal("memory store must not be classified as secure")
	}
}

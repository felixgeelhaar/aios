package runtime

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNewKeychainTokenStoreRequiresService(t *testing.T) {
	if _, err := NewKeychainTokenStore(""); err == nil {
		t.Fatal("expected error")
	}
}

func TestKeychainTokenStorePutGet(t *testing.T) {
	oldSet := keychainSet
	oldGet := keychainGet
	t.Cleanup(func() {
		keychainSet = oldSet
		keychainGet = oldGet
	})

	saved := map[string]string{}
	keychainSet = func(service, user, password string) error {
		saved[service+":"+user] = password
		return nil
	}
	keychainGet = func(service, user string) (string, error) {
		v, ok := saved[service+":"+user]
		if !ok {
			return "", errors.New("missing")
		}
		return v, nil
	}

	store, err := NewKeychainTokenStore("aios")
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	if err := store.Put(context.Background(), "gdrive", "token123"); err != nil {
		t.Fatalf("put: %v", err)
	}
	got, err := store.Get(context.Background(), "gdrive")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "token123" {
		t.Fatalf("got %q", got)
	}
}

// AC3: Must fail securely with clear error if keychain is unavailable or locked.
func TestKeychainTokenStoreFailsOnKeychainError(t *testing.T) {
	oldSet := keychainSet
	oldGet := keychainGet
	t.Cleanup(func() {
		keychainSet = oldSet
		keychainGet = oldGet
	})

	keychainSet = func(_, _, _ string) error {
		return errors.New("keychain is locked")
	}
	keychainGet = func(_, _ string) (string, error) {
		return "", errors.New("keychain is locked")
	}

	store, err := NewKeychainTokenStore("aios")
	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	// Put must fail with wrapped error.
	err = store.Put(context.Background(), "gdrive", "token")
	if err == nil {
		t.Fatal("expected error on locked keychain put")
	}
	if !strings.Contains(err.Error(), "keychain") {
		t.Fatalf("expected keychain error context, got %q", err.Error())
	}

	// Get must fail with wrapped error.
	_, err = store.Get(context.Background(), "gdrive")
	if err == nil {
		t.Fatal("expected error on locked keychain get")
	}
	if !strings.Contains(err.Error(), "keychain") {
		t.Fatalf("expected keychain error context, got %q", err.Error())
	}
}

// AC3: Put rejects empty key.
func TestKeychainTokenStorePutRejectsEmptyKey(t *testing.T) {
	oldSet := keychainSet
	t.Cleanup(func() { keychainSet = oldSet })
	keychainSet = func(_, _, _ string) error { return nil }

	store, _ := NewKeychainTokenStore("aios")
	if err := store.Put(context.Background(), "", "value"); err == nil {
		t.Fatal("expected error for empty key")
	}
}

// AC3: Put rejects empty value.
func TestKeychainTokenStorePutRejectsEmptyValue(t *testing.T) {
	oldSet := keychainSet
	t.Cleanup(func() { keychainSet = oldSet })
	keychainSet = func(_, _, _ string) error { return nil }

	store, _ := NewKeychainTokenStore("aios")
	if err := store.Put(context.Background(), "gdrive", ""); err == nil {
		t.Fatal("expected error for empty value")
	}
}

// AC3: Get rejects empty key.
func TestKeychainTokenStoreGetRejectsEmptyKey(t *testing.T) {
	oldGet := keychainGet
	t.Cleanup(func() { keychainGet = oldGet })
	keychainGet = func(_, _ string) (string, error) { return "", nil }

	store, _ := NewKeychainTokenStore("aios")
	if _, err := store.Get(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty key")
	}
}

// AC4: Keychain store never writes plaintext to disk — Put delegates to keyring, not filesystem.
func TestKeychainTokenStoreNoDiskWrite(t *testing.T) {
	oldSet := keychainSet
	t.Cleanup(func() { keychainSet = oldSet })

	var calledService, calledUser string
	keychainSet = func(service, user, password string) error {
		calledService = service
		calledUser = user
		return nil
	}

	store, _ := NewKeychainTokenStore("aios-test")
	if err := store.Put(context.Background(), "gdrive", "secret-token"); err != nil {
		t.Fatalf("put: %v", err)
	}
	// Verify the keyring API was called with correct service/key (not a file path).
	if calledService != "aios-test" {
		t.Fatalf("expected service aios-test, got %q", calledService)
	}
	if calledUser != "gdrive" {
		t.Fatalf("expected user gdrive, got %q", calledUser)
	}
}

// AC6: Must support token refresh — Put overwrites existing token.
func TestKeychainTokenStoreRefreshOverwrite(t *testing.T) {
	oldSet := keychainSet
	oldGet := keychainGet
	t.Cleanup(func() {
		keychainSet = oldSet
		keychainGet = oldGet
	})

	saved := map[string]string{}
	keychainSet = func(service, user, password string) error {
		saved[service+":"+user] = password
		return nil
	}
	keychainGet = func(service, user string) (string, error) {
		v, ok := saved[service+":"+user]
		if !ok {
			return "", errors.New("missing")
		}
		return v, nil
	}

	store, _ := NewKeychainTokenStore("aios")
	ctx := context.Background()

	// Store initial token.
	if err := store.Put(ctx, "gdrive", "old-token"); err != nil {
		t.Fatalf("put old: %v", err)
	}
	// Refresh with new token.
	if err := store.Put(ctx, "gdrive", "new-token"); err != nil {
		t.Fatalf("put new: %v", err)
	}
	got, err := store.Get(ctx, "gdrive")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "new-token" {
		t.Fatalf("expected refreshed token 'new-token', got %q", got)
	}
}

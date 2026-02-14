package runtime

import (
	"context"
	"fmt"

	keyring "github.com/zalando/go-keyring"
)

var (
	keychainSet = keyring.Set
	keychainGet = keyring.Get
)

type KeychainTokenStore struct {
	service string
}

func NewKeychainTokenStore(service string) (*KeychainTokenStore, error) {
	if service == "" {
		return nil, fmt.Errorf("service is required")
	}
	return &KeychainTokenStore{service: service}, nil
}

func (k *KeychainTokenStore) Put(_ context.Context, key, value string) error {
	if key == "" {
		return fmt.Errorf("key is required")
	}
	if value == "" {
		return fmt.Errorf("value is required")
	}
	if err := keychainSet(k.service, key, value); err != nil {
		return fmt.Errorf("keychain set %q: %w", key, err)
	}
	return nil
}

func (k *KeychainTokenStore) Get(_ context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key is required")
	}
	value, err := keychainGet(k.service, key)
	if err != nil {
		return "", fmt.Errorf("keychain get %q: %w", key, err)
	}
	return value, nil
}

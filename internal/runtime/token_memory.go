package runtime

import (
	"context"
	"fmt"
)

type MemoryTokenStore struct {
	values map[string]string
}

func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{values: map[string]string{}}
}

func (m *MemoryTokenStore) Put(_ context.Context, key, value string) error {
	m.values[key] = value
	return nil
}

func (m *MemoryTokenStore) Get(_ context.Context, key string) (string, error) {
	value, ok := m.values[key]
	if !ok {
		return "", fmt.Errorf("key %q not found", key)
	}
	return value, nil
}

package runtime

import "fmt"

func NewProductionRuntime(workspace, keychainService string) (*Runtime, error) {
	store, err := NewKeychainTokenStore(keychainService)
	if err != nil {
		return nil, fmt.Errorf("create keychain store: %w", err)
	}
	return New(workspace, store), nil
}

func (r *Runtime) UsesSecureTokenStore() bool {
	_, ok := r.store.(*KeychainTokenStore)
	return ok
}

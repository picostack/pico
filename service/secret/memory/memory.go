package memory

import (
	"context"

	"github.com/picostack/picobot/service/secret"
)

// MemorySecrets implements a simple in-memory secret.Store for testing
type MemorySecrets struct {
	Secrets map[string]string
}

var _ secret.Store = &MemorySecrets{}

// GetSecretsForTarget implements secret.Store
func (v *MemorySecrets) GetSecretsForTarget(name string) (map[string]string, error) {
	return v.Secrets, nil
}

// Renew implements secret.Store
func (v *MemorySecrets) Renew(ctx context.Context) error {
	return nil
}

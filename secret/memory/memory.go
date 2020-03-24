package memory

import (
	"github.com/picostack/pico/secret"
)

// MemorySecrets implements a simple in-memory secret.Store for testing
type MemorySecrets struct {
	Secrets map[string]map[string]string
}

var _ secret.Store = &MemorySecrets{}

// GetSecretsForTarget implements secret.Store
func (v *MemorySecrets) GetSecretsForTarget(name string) (map[string]string, error) {
	table, ok := v.Secrets[name]
	if !ok {
		return nil, nil
	}
	return table, nil
}

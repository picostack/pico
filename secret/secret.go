// Package secret provides an interface and implementations for secret storage.
// A secret store is passed to the executor, which hydrates execution tasks with
// any secrets that match it.
package secret

// Store describes a type that can securely obtain secrets for services.
type Store interface {
	GetSecretsForTarget(name string) (map[string]string, error)
}

// Package secret provides an interface and implementations for secret storage.
// A secret store is passed to the executor, which hydrates execution tasks with
// any secrets that match it.
package secret

import "strings"

// Store describes a type that can securely obtain secrets for services.
type Store interface {
	GetSecretsForTarget(name string) (map[string]string, error)
}

// GetPrefixedSecrets uses a Store to get a set of secrets that use a prefix.
func GetPrefixedSecrets(s Store, path, prefix string) (map[string]string, error) {
	all, err := s.GetSecretsForTarget(path)
	if err != nil {
		return nil, err
	}
	pass := make(map[string]string)
	for k, v := range all {
		if strings.HasPrefix(k, prefix) {
			pass[k] = v
		}
	}
	return pass, nil
}

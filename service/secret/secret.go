package secret

import "context"

// Store describes a type that can securely obtain secrets for services.
type Store interface {
	GetSecretsForTarget(name string) (map[string]string, error)
	Renew(ctx context.Context) error
}

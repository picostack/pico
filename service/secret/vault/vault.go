package vault

import (
	"context"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/picostack/pico/service/secret"
)

// VaultSecrets implements a secret.Store backed by Hashicorp Vault
type VaultSecrets struct {
	client  *api.Client
	path    string
	renewal time.Duration
}

var _ secret.Store = &VaultSecrets{}

// New creates a new Vault client and pings the server
func New(addr, path, token string, renewal time.Duration) (*VaultSecrets, error) {
	client, err := api.NewClient(&api.Config{
		Address:    addr,
		HttpClient: cleanhttp.DefaultClient(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client")
	}
	client.SetToken(token)

	if _, err = client.Auth().Token().LookupSelf(); err != nil {
		return nil, errors.Wrap(err, "failed to connect to vault server")
	}

	return &VaultSecrets{
		client:  client,
		path:    path,
		renewal: renewal,
	}, nil
}

// GetSecretsForTarget implements secret.Store
func (v *VaultSecrets) GetSecretsForTarget(name string) (map[string]string, error) {
	path := filepath.Join(v.path, name)

	zap.L().Debug("looking for secrets in vault",
		zap.String("name", name),
		zap.String("path", path))

	secret, err := v.client.Logical().Read(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read secret")
	}
	if secret == nil {
		zap.L().Debug("did not find secrets in vault",
			zap.String("name", name),
			zap.String("path", path))
		return nil, nil
	}

	env := make(map[string]string)
	for k, v := range secret.Data {
		env[k] = v.(string)
	}

	zap.L().Debug("found secrets in vault",
		zap.Any("secrets", env),
		zap.Int("count", len(env)))

	return env, nil
}

// RenewEvery starts a renewal ticker and blocks until fatal error
// works well with github.com/eapache/go-resiliency
func (v *VaultSecrets) Renew(ctx context.Context) error {
	if ctx.Err() == context.Canceled {
		return ctx.Err()
	}

	renew := time.NewTicker(v.renewal)
	defer renew.Stop()
	for range renew.C {
		_, err := v.client.Auth().Token().RenewSelf(0)
		if err != nil {
			return errors.Wrap(err, "failed to renew vault token")
		}
	}
	return nil
}

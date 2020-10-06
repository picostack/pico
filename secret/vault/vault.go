package vault

import (
	"context"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/picostack/pico/secret"
)

// VaultSecrets implements a secret.Store backed by Hashicorp Vault
type VaultSecrets struct {
	client     *api.Client
	enginepath string
	path       string
	version    int
	renewal    time.Duration
}

var _ secret.Store = &VaultSecrets{}

// New creates a new Vault client and pings the server
func New(addr, basepath, token string, renewal time.Duration) (v *VaultSecrets, err error) {
	if strings.HasPrefix(basepath, "/") {
		basepath = basepath[1:]
	}

	v = &VaultSecrets{
		renewal: renewal,
	}

	if v.client, err = api.NewClient(&api.Config{
		Address:    addr,
		HttpClient: cleanhttp.DefaultClient(),
	}); err != nil {
		return nil, errors.Wrap(err, "failed to create vault client")
	}
	v.client.SetToken(token)

	if _, err = v.client.Auth().Token().LookupSelf(); err != nil {
		return nil, errors.Wrap(err, "failed to connect to vault server")
	}

	// engine is the first component of base, then the rest is the actual path.
	v.enginepath, v.path = splitPath(basepath)

	if v.version, err = getKVEngineVersion(v.client, v.enginepath); err != nil {
		return nil, errors.Wrapf(err, "failed to determine KV engine version at '/%s'", v.enginepath)
	}

	zap.L().Debug("created new vault client for secrets engine",
		zap.Int("kv_version", v.version),
		zap.String("basepath", basepath),
		zap.String("enginepath", v.enginepath))

	return v, nil
}

// GetSecretsForTarget implements secret.Store
func (v *VaultSecrets) GetSecretsForTarget(name string) (map[string]string, error) {
	path := v.buildPath(name)

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

	env, err := kvToMap(v.version, secret.Data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unwrap secret data")
	}

	zap.L().Debug("found secrets in vault",
		zap.Strings("secret", keys(env)))

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
		token, err := v.client.Auth().Token().RenewSelf(0)
		if err != nil {
			return errors.Wrap(err, "failed to renew vault token")
		}
		if _, err = v.client.Auth().Token().LookupSelf(); err != nil {
			return errors.Wrap(err, "failed to connect to vault server")
		}
		v.client.SetToken(token.Auth.ClientToken)
		zap.L().Debug("renewed fault token",
			zap.String("lease_id", token.LeaseID),
			zap.Int("lease_duration", token.LeaseDuration),
			zap.Bool("renewable", token.Renewable))
	}
	return nil
}

func splitPath(basepath string) (string, string) {
	basepath = strings.Trim(basepath, "/")
	s := strings.SplitN(basepath, "/", 2)
	if len(s[0]) == 0 {
		return basepath, "/"
	} else if len(s) == 1 {
		return basepath, "/"
	}
	return s[0], s[1]
}

// builds the correct path to a secret based on the kv version
func (v *VaultSecrets) buildPath(item string) string {
	if v.version == 1 {
		path.Split(v.path)
		return path.Join(v.enginepath, v.path, item)
	}
	return path.Join(v.enginepath, "data", v.path, item)
}

// pulls out the kv secret data for v1 and v2 secrets
func kvToMap(version int, data map[string]interface{}) (env map[string]string, err error) {
	if version == 1 {
		env = make(map[string]string)
		for k, v := range data {
			env[k] = v.(string) //nolint:err - we know it's a string already
		}
	} else if version == 2 {
		env = make(map[string]string)
		if kv, ok := data["data"].(map[string]interface{}); ok {
			for k, v := range kv {
				env[k] = v.(string) //nolint:err - we know it's a string already
			}
		} else {
			return nil, errors.New("could not interpret KV v2 response data as hashtable, this is likely a change in the KV v2 API, please open an issue")
		}
	} else {
		return nil, errors.Errorf("unrecognised KV version: %d", version)
	}
	return
}

func keys(m map[string]string) (k []string) {
	for x := range m {
		k = append(k, x)
	}
	return
}

// because Vault has no way to know if a kv engine is v1 or v2, we have to check
// for the /config path and if it doesn't exist, attempt to LIST the path, if
// that succeeds, it's a v1, if it doesn't succeed, it *might still* be a v1 but
// empty, and in that case there's no way to know so it just bails. Amazing.
func getKVEngineVersion(client *api.Client, basepath string) (int, error) {
	// only KV v2 has /config, /data, /metadata paths, so attempt to read one
	s, err := client.Logical().Read(path.Join(basepath, "config"))
	if err != nil {
		return 0, errors.Wrap(err, "failed to check engine config path for version query")
	}
	if s == nil {
		// no /config path present, now attempt to list the engine's base path
		l, err := client.Logical().List(path.Join(basepath))
		if err != nil {
			return 0, errors.Wrap(err, "failed to list possible KV v1 engine")
		}
		if l == nil {
			return 0, errors.New("could not read secrets engine, it's either an empty KV v1 engine or does not exist")
		}
		// engine does not have a /config but contains elements, it's a v1.
		return 1, nil
	}

	// has a /config, it's a v2
	return 2, nil
}

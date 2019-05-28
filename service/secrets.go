package service

import (
	"path/filepath"

	"github.com/pkg/errors"
)

func (app *App) getSecretsForTarget(name string) (env map[string]string, err error) {
	if app.vault != nil {
		return app.secretsFromVault(name)
	}
	return
}

func (app *App) secretsFromVault(name string) (env map[string]string, err error) {
	path := filepath.Join("/secret", app.config.VaultPath, name)
	secret, err := app.vault.Logical().Read(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read secret")
	}
	if secret == nil {
		return
	}

	env = make(map[string]string)
	for k, v := range secret.Data {
		env[k] = v.(string)
	}
	return
}

package service

import (
	"github.com/pkg/errors"
)

func (app *App) getSecretsForTarget(name string) (env map[string]string, err error) {
	path := "/secret/data/" + name
	secret, err := app.vault.Logical().Read(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read secret")
	}
	if secret == nil {
		return
	}

	data, ok := secret.Data["data"]
	if !ok {
		return nil, errors.New("data field missing from secret payload")
	}
	dataAsMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, errors.New("failed to cast data payload to map[string]interface{}")
	}

	env = make(map[string]string)
	for k, v := range dataAsMap {
		env[k] = v.(string)
	}
	return
}

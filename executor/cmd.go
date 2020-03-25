package executor

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/picostack/pico/secret"
	"github.com/picostack/pico/task"
)

var _ Executor = &CommandExecutor{}

// CommandExecutor handles command invocation targets
type CommandExecutor struct {
	secrets            secret.Store
	passEnvironment    bool   // pass the Pico process environment to children
	configSecretPath   string // path to global secrets to pass to children
	configSecretPrefix string // only pass secrets with this prefix, usually GLOBAL_
}

// NewCommandExecutor creates a new CommandExecutor
func NewCommandExecutor(
	secrets secret.Store,
	passEnvironment bool,
	configSecretPath string,
	configSecretPrefix string,
) CommandExecutor {
	return CommandExecutor{
		secrets:            secrets,
		passEnvironment:    passEnvironment,
		configSecretPath:   configSecretPath,
		configSecretPrefix: configSecretPrefix,
	}
}

// Subscribe implements executor.Executor
func (e *CommandExecutor) Subscribe(bus chan task.ExecutionTask) {
	for t := range bus {
		if err := e.execute(t.Target, t.Path, t.Shutdown, t.Env); err != nil {
			zap.L().Error("executor task unsuccessful",
				zap.String("target", t.Target.Name),
				zap.Bool("shutdown", t.Shutdown),
				zap.Error(err))
		}
	}
}

type exec struct {
	path            string
	env             map[string]string
	shutdown        bool
	passEnvironment bool
}

func (e *CommandExecutor) prepare(
	name string,
	path string,
	shutdown bool,
	execEnv map[string]string,
) (exec, error) {
	// get global secrets from the Pico config path in the secret store.
	// only secrets with the prefix are retrieved.
	global, err := secret.GetPrefixedSecrets(e.secrets, e.configSecretPath, e.configSecretPrefix)
	if err != nil {
		return exec{}, errors.Wrap(err, "failed to get global secrets for target")
	}

	secrets, err := e.secrets.GetSecretsForTarget(name)
	if err != nil {
		return exec{}, errors.Wrap(err, "failed to get secrets for target")
	}

	env := make(map[string]string)

	// merge execution environment with secrets in the following order:
	// globals first, then execution environment, then per-target secrets
	for k, v := range global {
		env[k] = v
	}
	for k, v := range execEnv {
		env[k] = v
	}
	for k, v := range secrets {
		env[k] = v
	}

	return exec{path, env, shutdown, e.passEnvironment}, nil
}

func (e *CommandExecutor) execute(
	target task.Target,
	path string,
	shutdown bool,
	execEnv map[string]string,
) (err error) {
	ex, err := e.prepare(target.Name, path, shutdown, execEnv)
	if err != nil {
		return err
	}

	zap.L().Debug("executing with secrets",
		zap.String("target", target.Name),
		zap.Strings("cmd", target.Up),
		zap.String("url", target.RepoURL),
		zap.String("dir", path),
		zap.Any("env", ex.env),
		zap.Bool("passthrough", e.passEnvironment))

	return target.Execute(ex.path, ex.env, ex.shutdown, ex.passEnvironment)
}

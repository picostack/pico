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
	secrets secret.Store
}

// NewCommandExecutor creates a new CommandExecutor
func NewCommandExecutor(secrets secret.Store) CommandExecutor {
	return CommandExecutor{
		secrets: secrets,
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

func (e *CommandExecutor) execute(
	target task.Target,
	path string,
	shutdown bool,
	execEnv map[string]string,
) (err error) {
	secrets, err := e.secrets.GetSecretsForTarget(target.Name)
	if err != nil {
		return errors.Wrap(err, "failed to get secrets for target")
	}

	env := make(map[string]string)

	// merge execution environment with secrets
	for k, v := range execEnv {
		env[k] = v
	}
	for k, v := range secrets {
		env[k] = v
	}

	zap.L().Debug("executing with secrets",
		zap.String("target", target.Name),
		zap.Strings("cmd", target.Up),
		zap.String("url", target.RepoURL),
		zap.String("dir", path),
		zap.Int("env", len(env)),
		zap.Int("secrets", len(secrets)))

	return target.Execute(path, env, shutdown)
}

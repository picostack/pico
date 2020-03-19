package executor

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/picostack/pico/service/secret"
	"github.com/picostack/pico/service/task"
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
func (e *CommandExecutor) Subscribe(bus chan task.ExecutionTask) error {
	for t := range bus {
		if err := e.execute(t.Target, t.Path, t.Shutdown); err != nil {
			return err
		}
	}
	return nil
}

func (e *CommandExecutor) execute(
	target task.Target,
	path string,
	shutdown bool,
) (err error) {
	env, err := e.secrets.GetSecretsForTarget(target.Name)
	if err != nil {
		return errors.Wrap(err, "failed to get secrets for target")
	}

	zap.L().Debug("executing with secrets",
		zap.String("target", target.Name),
		zap.Strings("cmd", target.Up),
		zap.String("url", target.RepoURL),
		zap.String("dir", path),
		zap.Int("secrets", len(env)))

	return target.Execute(path, env, shutdown)
}

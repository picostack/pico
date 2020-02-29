package service

import (
	"path/filepath"

	"github.com/Southclaws/gitwatch"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/picostack/picobot/service/task"
)

// handle takes an event from gitwatch and runs the event's triggers
func (app *App) handle(e gitwatch.Event) (err error) {
	target, exists := app.getTarget(e.URL)
	if !exists {
		return errors.Errorf("attempt to handle event for unknown target %s at %s", e.URL, e.Path)
	}
	zap.L().Debug("handling event",
		zap.String("target", target.Name),
		zap.String("url", target.RepoURL),
		zap.Time("timestamp", e.Timestamp))
	return app.executeWithSecrets(target, e.Path, false)
}

func (app *App) executeWithSecrets(target task.Target, path string, shutdown bool) (err error) {
	env, err := app.getSecretsForTarget(target.Name)
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

func (app App) getTarget(url string) (target task.Target, exists bool) {
	for _, t := range app.targets {
		if t.RepoURL == url {
			return t, true
		}
	}
	return
}

func (app App) executeTargets(targets []task.Target, shutdown bool) {
	zap.L().Debug("executing all targets", zap.Bool("shutdown", shutdown))
	for _, t := range targets {
		err := app.executeWithSecrets(
			t,
			filepath.Join(app.config.Directory, t.Name),
			shutdown,
		)
		if err != nil {
			zap.L().Error("failed to execute task after reconfigure",
				zap.Error(errors.Cause(err)))
			continue
		}
	}
	return
}

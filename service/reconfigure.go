package service

import (
	"github.com/Southclaws/gitwatch"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/Southclaws/wadsworth/service/config"
	"github.com/Southclaws/wadsworth/service/task"
)

// reconfigure will close the configuration watcher and target watcher (unless
// it's the first run) then create a watcher for the application's config target
// repo (the repo that contains the configuration file(s)) then wait for the
// first event (either from a fresh clone, a pull, or just a noop event) then
// read the configuration file(s) from the repository, gather all the targets
// and set up the target watcher. This should always happen in sync with the
// rest of the service to prevent a reconfiguration during an event handler.
func (app *App) reconfigure() (err error) {
	zap.L().Debug("reconfiguring")
	if app.configWatcher != nil {
		zap.L().Debug("closing existing watcher")
		app.configWatcher.Close()
	}

	app.configWatcher, err = gitwatch.New(
		app.ctx,
		[]string{app.config.Target},
		app.config.CheckInterval,
		app.config.Directory,
		nil,
		true)
	if err != nil {
		return errors.Wrap(err, "failed to watch config target")
	}
	go app.configWatcher.Run() //nolint:errcheck - no worthwhile errors returned
	zap.L().Debug("created new config watcher, awaiting setup")

	<-app.configWatcher.InitialDone
	zap.L().Debug("config initial setup done")

	path, err := gitwatch.GetRepoPath(app.config.Directory, app.config.Target)
	if err != nil {
		return
	}

	// TODO: if this fails, log an error and fall back to the old state
	state, err := config.ConfigFromDirectory(path)
	if err != nil {
		return errors.Wrap(err, "failed to construct config from repo")
	}
	zap.L().Debug("constructed desired state", zap.Int("targets", len(state.Targets)))

	if app.targetsWatcher != nil {
		app.targetsWatcher.Close()
	}

	// TODO: diff what changed, run the `down` command for those that were removed

	app.targets = make(map[string]task.Target)
	targets := make([]string, len(state.Targets))
	for i, t := range state.Targets {
		zap.L().Debug("assigned target", zap.String("url", t.RepoURL))
		targets[i] = t.RepoURL
		app.targets[t.RepoURL] = t
	}

	app.targetsWatcher, err = gitwatch.New(
		app.ctx,
		targets,
		app.config.CheckInterval,
		app.config.Directory,
		nil,
		true)
	if err != nil {
		return errors.Wrap(err, "failed to watch targets")
	}
	go app.targetsWatcher.Run() //nolint:errcheck - no worthwhile errors returned
	zap.L().Debug("created targets watcher, awaiting setup")

	<-app.targetsWatcher.InitialDone
	zap.L().Debug("targets initial setup done")

	return
}

package service

import (
	"reflect"

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

	err = app.watchConfig()
	if err != nil {
		return
	}

	// generate a new desired state from the config repo
	path, err := gitwatch.GetRepoPath(app.config.Directory, app.config.Target)
	if err != nil {
		return
	}
	state := getNewState(path, app.state)

	// diff targets
	additions, removals := diffTargets(app.targets, state.Targets)
	app.targets = state.Targets

	err = app.watchTargets()
	if err != nil {
		return
	}

	// out with the old, in with the new!
	app.executeTargets(removals, true)
	app.executeTargets(additions, false)

	zap.L().Debug("targets initial up done")

	app.state = state

	return
}

// watchConfig creates or restarts the watcher that reacts to changes to the
// repo that contains wadsworth configuration scripts
func (app *App) watchConfig() (err error) {
	if app.configWatcher != nil {
		zap.L().Debug("closing existing watcher")
		app.configWatcher.Close()
	}

	app.configWatcher, err = gitwatch.New(
		app.ctx,
		[]string{app.config.Target},
		app.config.CheckInterval,
		app.config.Directory,
		app.ssh,
		true)
	if err != nil {
		return errors.Wrap(err, "failed to watch config target")
	}
	go app.configWatcher.Run() //nolint:errcheck - no worthwhile errors returned
	zap.L().Debug("created new config watcher, awaiting setup")

	<-app.configWatcher.InitialDone
	zap.L().Debug("config initial setup done")

	return
}

// watchTargets creates or restarts the watcher that reacts to changes to target
// repositories that contain actual apps and services
func (app *App) watchTargets() (err error) {
	targetURLs := make([]string, len(app.targets))
	for _, t := range app.targets {
		zap.L().Debug("assigned target", zap.String("url", t.RepoURL))
		targetURLs = append(targetURLs, t.RepoURL)
	}

	if app.targetsWatcher != nil {
		app.targetsWatcher.Close()
	}
	app.targetsWatcher, err = gitwatch.New(
		app.ctx,
		targetURLs,
		app.config.CheckInterval,
		app.config.Directory,
		app.ssh,
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

// getNewState attempts to obtain a new desired state from the given path, if
// any failures occur, it simply returns a fallback state and logs an error
func getNewState(path string, fallback config.State) (state config.State) {
	state, err := config.ConfigFromDirectory(path)
	if err != nil {
		zap.L().Error("failed to construct config from repo, falling back to original state",
			zap.Error(err))

		state = fallback
	} else {
		zap.L().Debug("constructed desired state",
			zap.Int("targets", len(state.Targets)))
	}
	return
}

// diffTargets returns just the additions (also changes) and removals between
// the specified old targets and new targets
func diffTargets(oldTargets, newTargets []task.Target) (additions, removals []task.Target) {
	for _, newTarget := range newTargets {
		var exists bool
		for _, oldTarget := range oldTargets {
			if oldTarget.Name == newTarget.Name {
				exists = true
				break
			}
		}
		if !exists {
			removals = append(removals, newTarget)
		}
	}
	for _, oldTarget := range oldTargets {
		var exists bool
		var newTarget task.Target
		for _, newTarget = range newTargets {
			if newTarget.Name == oldTarget.Name {
				exists = true
				break
			}
		}
		if !exists {
			additions = append(additions, oldTarget)
		} else if !reflect.DeepEqual(oldTarget, newTarget) {
			additions = append(additions, oldTarget)
		}
	}
	return
}

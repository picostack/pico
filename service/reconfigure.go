package service

import (
	"context"
	"path/filepath"
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
func (app *App) reconfigure(hostname string) (err error) {
	zap.L().Debug("reconfiguring")

	err = app.watchConfig()
	if err != nil {
		return
	}

	// generate a new desired state from the config repo
	path, err := gitwatch.GetRepoDirectory(app.config.Target)
	if err != nil {
		return
	}
	state := getNewState(
		filepath.Join(app.config.Directory, path),
		hostname,
		app.state,
	)

	// Set the HOSTNAME config environment variable if necessary.
	if app.config.Hostname != "" {
		state.Env["HOSTNAME"] = app.config.Hostname
	}

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
		[]gitwatch.Repository{{URL: app.config.Target}},
		app.config.CheckInterval,
		app.config.Directory,
		app.ssh,
		false)
	if err != nil {
		return errors.Wrap(err, "failed to watch config target")
	}

	go func() {
		e := app.configWatcher.Run()
		if e != nil && !errors.Is(e, context.Canceled) {
			app.errors <- e
		}
	}()
	zap.L().Debug("created new config watcher, awaiting setup")

	return
}

// watchTargets creates or restarts the watcher that reacts to changes to target
// repositories that contain actual apps and services
func (app *App) watchTargets() (err error) {
	targetRepos := make([]gitwatch.Repository, len(app.targets))
	for i, t := range app.targets {
		zap.L().Debug("assigned target", zap.String("url", t.RepoURL))
		targetRepos[i] = gitwatch.Repository{
			URL:       t.RepoURL,
			Branch:    t.Branch,
			Directory: t.Name,
		}
	}

	if app.targetsWatcher != nil {
		app.targetsWatcher.Close()
	}
	app.targetsWatcher, err = gitwatch.New(
		app.ctx,
		targetRepos,
		app.config.CheckInterval,
		app.config.Directory,
		app.ssh,
		false)
	if err != nil {
		return errors.Wrap(err, "failed to watch targets")
	}

	go func() {
		e := app.targetsWatcher.Run()
		if e != nil && !errors.Is(e, context.Canceled) {
			app.errors <- e
		}
	}()
	zap.L().Debug("created targets watcher, awaiting setup")

	return
}

// getNewState attempts to obtain a new desired state from the given path, if
// any failures occur, it simply returns a fallback state and logs an error
func getNewState(path, hostname string, fallback config.State) (state config.State) {
	state, err := config.ConfigFromDirectory(path, hostname)
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
			additions = append(additions, newTarget)
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
			removals = append(removals, oldTarget)
		} else if !reflect.DeepEqual(oldTarget, newTarget) {
			additions = append(additions, newTarget)
		}
	}
	return
}

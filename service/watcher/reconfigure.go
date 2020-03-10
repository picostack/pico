package watcher

import (
	"context"
	"path/filepath"
	"reflect"

	"github.com/Southclaws/gitwatch"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/picostack/pico/service/config"
	"github.com/picostack/pico/service/task"
)

// reconfigure will close the configuration watcher and target watcher (unless
// it's the first run) then create a watcher for the application's config target
// repo (the repo that contains the configuration file(s)) then wait for the
// first event (either from a fresh clone, a pull, or just a noop event) then
// read the configuration file(s) from the repository, gather all the targets
// and set up the target watcher. This should always happen in sync with the
// rest of the service to prevent a reconfiguration during an event handler.
func (w *Watcher) reconfigure() (err error) {
	zap.L().Debug("reconfiguring")

	err = w.watchConfig()
	if err != nil {
		return
	}

	// generate a new desired state from the config repo
	path, err := gitwatch.GetRepoDirectory(w.configRepo)
	if err != nil {
		return
	}
	state := getNewState(
		filepath.Join(w.directory, path),
		w.hostname,
		w.state,
	)

	// Set the HOSTNAME config environment variable if necessary.
	if w.hostname != "" {
		state.Env["HOSTNAME"] = w.hostname
	}

	// diff targets
	additions, removals := diffTargets(w.targets, state.Targets)
	w.targets = state.Targets

	err = w.watchTargets()
	if err != nil {
		return
	}

	// out with the old, in with the new!
	w.executeTargets(removals, true)
	w.executeTargets(additions, false)

	zap.L().Debug("targets initial up done")

	w.state = state

	return
}

// watchConfig creates or restarts the watcher that reacts to changes to the
// repo that contains pico configuration scripts
func (w *Watcher) watchConfig() (err error) {
	if w.configWatcher != nil {
		zap.L().Debug("closing existing watcher")
		w.configWatcher.Close()
	}

	w.configWatcher, err = gitwatch.New(
		context.TODO(),
		[]gitwatch.Repository{{URL: w.configRepo}},
		w.checkInterval,
		w.directory,
		w.ssh,
		false)
	if err != nil {
		return errors.Wrap(err, "failed to watch config target")
	}

	go func() {
		e := w.configWatcher.Run()
		if e != nil && !errors.Is(e, context.Canceled) {
			w.errors <- e
		}
	}()
	zap.L().Debug("created new config watcher, awaiting setup")

	<-w.configWatcher.InitialDone

	return
}

// watchTargets creates or restarts the watcher that reacts to changes to target
// repositories that contain actual apps and services
func (w *Watcher) watchTargets() (err error) {
	targetRepos := make([]gitwatch.Repository, len(w.targets))
	for i, t := range w.targets {
		zap.L().Debug("assigned target", zap.String("url", t.RepoURL))
		targetRepos[i] = gitwatch.Repository{
			URL:       t.RepoURL,
			Branch:    t.Branch,
			Directory: t.Name,
		}
	}

	if w.targetsWatcher != nil {
		w.targetsWatcher.Close()
	}
	w.targetsWatcher, err = gitwatch.New(
		context.TODO(),
		targetRepos,
		w.checkInterval,
		w.directory,
		w.ssh,
		false)
	if err != nil {
		return errors.Wrap(err, "failed to watch targets")
	}

	go func() {
		e := w.targetsWatcher.Run()
		if e != nil && !errors.Is(e, context.Canceled) {
			w.errors <- e
		}
	}()
	zap.L().Debug("created targets watcher, awaiting setup")

	<-w.targetsWatcher.InitialDone

	return
}

// getNewState attempts to obtain a new desired state from the given path, if
// any failures occur, it simply returns a fallback state and logs an error
func getNewState(path, hostname string, fallback config.State) (state config.State) {
	state, err := config.ConfigFromDirectory(path, hostname)
	if err != nil {
		zap.L().Error("failed to construct config from repo, falling back to original state",
			zap.String("path", path),
			zap.String("hostname", hostname),
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

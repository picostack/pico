package watcher

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/Southclaws/gitwatch"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"

	"github.com/picostack/pico/service/config"
	"github.com/picostack/pico/service/task"
)

var _ Watcher = &GitWatcher{}

// GitWatcher implements a Watcher for monitoring Git repositories and executing
// tasks associated with those Git repositories when they receive commits.
type GitWatcher struct {
	directory     string
	bus           chan task.ExecutionTask
	checkInterval time.Duration
	ssh           transport.AuthMethod

	targetsWatcher *gitwatch.Session
	state          config.State

	initialised bool
	initialise  chan bool
	newState    chan config.State
	errors      chan error
}

// NewGitWatcher creates a new watcher with all necessary parameters
func NewGitWatcher(
	directory string,
	bus chan task.ExecutionTask,
	checkInterval time.Duration,
	ssh transport.AuthMethod,
) *GitWatcher {
	return &GitWatcher{
		directory:     directory,
		bus:           bus,
		checkInterval: checkInterval,
		ssh:           ssh,

		initialise: make(chan bool),
		newState:   make(chan config.State, 16),
		errors:     make(chan error, 16),
	}
}

// Start runs the watcher loop and blocks until a fatal error occurs
func (w *GitWatcher) Start() error {
	// wait for the first config event to set the initial state
	<-w.initialise

	zap.L().Debug("git watcher initialised", zap.Any("initial_state", w.state))

	f := func() (err error) {
		select {
		case newState := <-w.newState:
			zap.L().Debug("git watcher received new state",
				zap.Any("new_state", newState))

			return w.doReconfigure(newState)

		case event := <-w.targetsWatcher.Events:
			zap.L().Debug("git watcher received a target event",
				zap.Any("new_state", event))

			if e := w.handle(event); e != nil {
				zap.L().Error("failed to handle event",
					zap.String("url", event.URL),
					zap.Error(e))
			}

		case e := <-errorMultiplex(w.errors, w.targetsWatcher.Errors):
			zap.L().Error("git error",
				zap.Error(e))
		}
		return
	}

	for {
		err := f()
		if err != nil {
			return err
		}
	}
}

// performs a reconfigure:
//   - diffs the new state against the old state to build a list of +/-
//   - creates a new targets watcher
//   - executs the necessary targets - first shut down old ones, then create new
//   - sets the watcher state field to the new state
func (w *GitWatcher) doReconfigure(newState config.State) error {
	additions, removals := task.DiffTargets(w.state.Targets, newState.Targets)
	w.state = newState

	err := w.watchTargets()
	if err != nil {
		return err
	}

	// out with the old, in with the new!
	w.executeTargets(removals, true)
	w.executeTargets(additions, false)

	return nil
}

func (w *GitWatcher) doInit(state config.State) error {
	if err := w.doReconfigure(state); err != nil {
		return err
	}
	w.initialised = true
	w.initialise <- true
	return nil
}

// SetState implements Watcher
// Upon state being updated, the watcher dispatches an event to its own channel
// to instruct the daemon loop to reconfigure. The reason for this is that loop
// keeps the whole system in sync, so reconfigurations don't happen mid way
// through a target event.
func (w *GitWatcher) SetState(state config.State) error {
	if !w.initialised {
		return w.doInit(state)
	}
	w.newState <- state
	return nil
}

// GetState implements Watcher
func (w *GitWatcher) GetState() config.State {
	return w.state
}

// watchTargets creates or restarts the targets watcher.
func (w *GitWatcher) watchTargets() (err error) {
	targetRepos := make([]gitwatch.Repository, len(w.state.Targets))
	for i, t := range w.state.Targets {
		dir := t.Name
		if t.Branch != "" {
			dir = fmt.Sprintf("%s_%s", t.Name, t.Branch)
		}
		zap.L().Debug("assigned target", zap.String("url", t.RepoURL), zap.String("directory", dir))
		targetRepos[i] = gitwatch.Repository{
			URL:       t.RepoURL,
			Branch:    t.Branch,
			Directory: dir,
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

	zap.L().Debug("targets watcher initialised")

	return
}

func (w *GitWatcher) handle(e gitwatch.Event) (err error) {
	target, exists := w.getTarget(e.URL)
	if !exists {
		return errors.Errorf("attempt to handle event for unknown target %s at %s", e.URL, e.Path)
	}
	zap.L().Debug("handling event",
		zap.String("target", target.Name),
		zap.String("url", target.RepoURL),
		zap.Time("timestamp", e.Timestamp))
	w.send(target, e.Path, false)
	return nil
}

func (w GitWatcher) executeTargets(targets []task.Target, shutdown bool) {
	zap.L().Debug("executing all targets",
		zap.Bool("shutdown", shutdown),
		zap.Int("targets", len(targets)))

	for _, t := range targets {
		w.send(t, filepath.Join(w.directory, t.Name), shutdown)
	}
}

func (w GitWatcher) getTarget(url string) (target task.Target, exists bool) {
	for _, t := range w.state.Targets {
		if t.RepoURL == url {
			return t, true
		}
	}
	return
}

func (w GitWatcher) send(target task.Target, path string, shutdown bool) {
	w.bus <- task.ExecutionTask{
		Target:   target,
		Path:     path,
		Shutdown: shutdown,
		Env:      w.state.Env,
	}
}

func errorMultiplex(chans ...<-chan error) <-chan error {
	out := make(chan error)
	go func() {
		var wg sync.WaitGroup
		wg.Add(len(chans))

		for _, c := range chans {
			go func(c <-chan error) {
				for v := range c {
					out <- v
				}
				wg.Done()
			}(c)
		}

		wg.Wait()
		close(out)
	}()
	return out
}

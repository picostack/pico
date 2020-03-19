package watcher

import (
	"sync"
	"time"

	"github.com/Southclaws/gitwatch"
	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"

	"github.com/picostack/pico/service/config"
	"github.com/picostack/pico/service/task"
)

// Watcher is responsible for a number of things, it's quite a large system (the
// largest in the entire codebase!) and is quite tightly coupled for various
// reasons.
//
// It is in charge of:
//   - storing the configuration state
//   - watching for configuration state changes (hence the name)
//   - reacting to configuration changes by rebooting itself with the new state
//   - dispatching events for targets when their repositories receive commits
//
// The reactive code for executing targets is separated by an event bus.
// Unfortunately, because this system holds the configuration state, the same
// approach can't be applied to configuration state change events - the system
// itself reacts to those events so it can reconfigure itself. *That* is why it
// has such a large structure with so much state.
//
type Watcher struct {
	bus           chan task.ExecutionTask
	hostname      string
	directory     string
	configRepo    string
	checkInterval time.Duration
	ssh           transport.AuthMethod

	configWatcher  *gitwatch.Session
	targetsWatcher *gitwatch.Session

	targets []task.Target
	state   config.State

	errors chan error
}

// New creates a new watcher with all necessary parameters
func New(
	bus chan task.ExecutionTask,
	hostname string,
	directory string,
	configRepo string,
	checkInterval time.Duration,
	ssh transport.AuthMethod,
) *Watcher {
	return &Watcher{
		bus:           bus,
		hostname:      hostname,
		directory:     directory,
		configRepo:    configRepo,
		checkInterval: checkInterval,
		ssh:           ssh,

		errors: make(chan error, 16),
	}
}

// Start runs the watcher and blocks until a fatal error occurs
func (w *Watcher) Start() error {
	if err := w.reconfigure(); err != nil {
		return err
	}

	f := func() (err error) {
		select {
		case <-w.configWatcher.Events:
			err = w.reconfigure()

		case event := <-w.targetsWatcher.Events:
			if e := w.handle(event); e != nil {
				zap.L().Error("failed to handle event",
					zap.String("url", event.URL),
					zap.Error(e))
			}

		case e := <-errorMultiplex(w.configWatcher.Errors, w.targetsWatcher.Errors):
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

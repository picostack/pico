package watcher

import (
	"sync"
	"time"

	"github.com/Southclaws/gitwatch"
	"github.com/picostack/pico/service/config"
	"github.com/picostack/pico/service/secret"
	"github.com/picostack/pico/service/task"
	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
)

// Watcher handles Git events and dispatches config or task execution events.
type Watcher struct {
	secrets       secret.Store
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
	secrets secret.Store,
	hostname string,
	directory string,
	configRepo string,
	checkInterval time.Duration,
	ssh transport.AuthMethod,
) *Watcher {
	return &Watcher{
		secrets:       secrets,
		hostname:      hostname,
		directory:     directory,
		configRepo:    configRepo,
		checkInterval: checkInterval,
		ssh:           ssh,

		errors: make(chan error, 16),
	}
}

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

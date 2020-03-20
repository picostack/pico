package reconfigurer

import (
	"context"
	"path/filepath"
	"time"

	"github.com/Southclaws/gitwatch"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"

	"github.com/picostack/pico/config"
	"github.com/picostack/pico/watcher"
)

var _ Provider = &GitProvider{}

// GitProvider implements a Provider backed by Git. It will reconfigure its
// watcher process upon commits to its defined configuration repository.
type GitProvider struct {
	directory     string
	hostname      string
	configRepo    string
	checkInterval time.Duration
	authMethod    transport.AuthMethod

	configWatcher *gitwatch.Session
}

// New creates a new provider with all necessary parameters
func New(
	directory string,
	hostname string,
	configRepo string,
	checkInterval time.Duration,
	authMethod transport.AuthMethod,
) *GitProvider {
	return &GitProvider{
		directory:     directory,
		hostname:      hostname,
		configRepo:    configRepo,
		checkInterval: checkInterval,
		authMethod:    authMethod,
	}
}

// Configure implements Provider
func (p *GitProvider) Configure(w watcher.Watcher) error {
	if err := p.reconfigure(w); err != nil {
		return err
	}

	for range p.configWatcher.Events {
		if err := p.reconfigure(w); err != nil {
			return err
		}
	}

	return nil
}

// reconfigure will close the configuration watcher (unless it's the first run)
// then create a watcher for the application's config target repo then wait for
// the first event (either from a fresh clone, a pull, or just a noop event)
// then update the state of the watcher it's in charge of.
func (p *GitProvider) reconfigure(w watcher.Watcher) (err error) {
	zap.L().Debug("reconfiguring")

	err = p.watchConfig()
	if err != nil {
		return
	}

	// generate a new desired state from the config repo
	path, err := gitwatch.GetRepoDirectory(p.configRepo)
	if err != nil {
		return
	}
	state := getNewState(
		filepath.Join(p.directory, path),
		p.hostname,
		w.GetState(),
	)

	// Set the HOSTNAME config environment variable if necessary.
	if p.hostname != "" {
		state.Env["HOSTNAME"] = p.hostname
	}

	return w.SetState(state)
}

// watchConfig creates or restarts the watcher that reacts to changes to the
// repo that contains pico configuration scripts
func (p *GitProvider) watchConfig() (err error) {
	if p.configWatcher != nil {
		zap.L().Debug("closing existing watcher")
		p.configWatcher.Close()
	}

	p.configWatcher, err = gitwatch.New(
		context.TODO(),
		[]gitwatch.Repository{{URL: p.configRepo}},
		p.checkInterval,
		p.directory,
		p.authMethod,
		false)
	if err != nil {
		return errors.Wrap(err, "failed to watch config target")
	}

	go func() {
		e := p.configWatcher.Run()
		if e != nil && !errors.Is(e, context.Canceled) {
			zap.L().Error("config watcher failed", zap.Error(e))
		}
	}()
	zap.L().Debug("created new config watcher, awaiting setup")

	<-p.configWatcher.InitialDone

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

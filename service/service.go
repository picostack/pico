package service

import (
	"context"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"

	"github.com/picostack/picobot/service/secret"
	"github.com/picostack/picobot/service/secret/memory"
	"github.com/picostack/picobot/service/secret/vault"
	"github.com/picostack/picobot/service/watcher"
)

// Config specifies static configuration parameters (from CLI or environment)
type Config struct {
	Target        string
	Hostname      string
	NoSSH         bool
	Directory     string
	CheckInterval time.Duration
	VaultAddress  string
	VaultToken    string
	VaultPath     string
	VaultRenewal  time.Duration
}

// App stores application state
type App struct {
	config  Config
	watcher *watcher.Watcher
	secrets secret.Store
}

// Initialise prepares an instance of the app to run
func Initialise(c Config) (app *App, err error) {
	app = new(App)

	app.config = c

	var authMethod transport.AuthMethod
	if !c.NoSSH {
		authMethod, err = ssh.NewSSHAgentAuth("git")
		if err != nil {
			return nil, errors.Wrap(err, "failed to set up SSH authentication")
		}
	}

	var secretStore secret.Store
	if c.VaultAddress != "" {
		secretStore, err = vault.New(c.VaultAddress, c.VaultPath, c.VaultToken, c.VaultRenewal)
		if err != nil {
			return nil, err
		}
	} else {
		secretStore = &memory.MemorySecrets{
			// TODO: pull env vars with PICO_SECRET_* or something and shove em here
		}
	}

	app.secrets = secretStore

	app.watcher = watcher.New(
		secretStore,
		c.Hostname,
		c.Directory,
		c.Target,
		c.CheckInterval,
		authMethod,
	)

	return
}

// Start launches the app and blocks until fatal error
func (app *App) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	zap.L().Debug("starting service daemon")

	g.Go(app.watcher.Start)

	g.Go(func() error {
		return retrier.New(retrier.ConstantBackoff(3, 100*time.Millisecond), nil).
			RunCtx(ctx, app.secrets.Renew)
	})

	return g.Wait()
}

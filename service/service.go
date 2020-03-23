// Package service provides the runnable service that acts as the root of the
// overall system. It provides a configuration structure, a way to initialise a
// primed instance of the service which can then be start via the .Start() func.
package service

import (
	"context"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"

	"github.com/picostack/pico/executor"
	"github.com/picostack/pico/reconfigurer"
	"github.com/picostack/pico/secret"
	"github.com/picostack/pico/secret/memory"
	"github.com/picostack/pico/secret/vault"
	"github.com/picostack/pico/task"
	"github.com/picostack/pico/watcher"
)

// Config specifies static configuration parameters (from CLI or environment)
type Config struct {
	Target        task.Repo
	Hostname      string
	SSH           bool
	Directory     string
	CheckInterval time.Duration
	VaultAddress  string
	VaultToken    string
	VaultPath     string
	VaultRenewal  time.Duration
	VaultConfig   string
}

// App stores application state
type App struct {
	config       Config
	reconfigurer reconfigurer.Provider
	watcher      watcher.Watcher
	secrets      secret.Store
	bus          chan task.ExecutionTask
}

// Initialise prepares an instance of the app to run
func Initialise(c Config) (app *App, err error) {
	app = new(App)

	app.config = c

	var secretStore secret.Store
	if c.VaultAddress != "" {
		zap.L().Debug("connecting to vault",
			zap.String("address", c.VaultAddress),
			zap.String("path", c.VaultPath),
			zap.String("token", c.VaultToken),
			zap.Duration("renewal", c.VaultRenewal))

		secretStore, err = vault.New(c.VaultAddress, c.VaultPath, c.VaultToken, c.VaultRenewal)
		if err != nil {
			return nil, err
		}
	} else {
		secretStore = &memory.MemorySecrets{
			// TODO: pull env vars with PICO_SECRET_* or something and shove em here
		}
	}

	secretConfig, err := secretStore.GetSecretsForTarget(c.VaultConfig)
	if err != nil {
		zap.L().Info("could not read additional config from vault", zap.String("path", c.VaultConfig))
		err = nil
	}
	zap.L().Debug("read configuration secrets from secret store", zap.Strings("keys", getKeys(secretConfig)))

	authMethod, err := getAuthMethod(c, secretConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create an authentication method from the given config")
	}

	app.secrets = secretStore

	app.bus = make(chan task.ExecutionTask, 100)

	// reconfigurer
	app.reconfigurer = reconfigurer.New(
		c.Directory,
		c.Hostname,
		c.Target.URL,
		c.CheckInterval,
		authMethod,
	)

	// target watcher
	app.watcher = watcher.NewGitWatcher(
		app.config.Directory,
		app.bus,
		app.config.CheckInterval,
		authMethod,
	)

	return
}

// Start launches the app and blocks until fatal error
func (app *App) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	zap.L().Debug("starting service daemon")

	// TODO: Replace this errgroup with a more resilient solution.
	// Not all of these tasks fail in the same way. Some don't fail at all.
	// This needs to be rewritten to be more considerate of different failure
	// states and potentially retry in some circumstances. Pico should be the
	// kind of service that barely goes down, only when absolutely necessary.

	ce := executor.NewCommandExecutor(app.secrets)
	g.Go(func() error {
		ce.Subscribe(app.bus)
		return nil
	})

	// TODO: gw can fail when setting up the gitwatch instance, it should retry.
	gw := app.watcher.(*watcher.GitWatcher)
	g.Go(gw.Start)

	// TODO: reconfigurer can also fail when setting up gitwatch.
	g.Go(func() error {
		return app.reconfigurer.Configure(app.watcher)
	})

	if s, ok := app.secrets.(*vault.VaultSecrets); ok {
		g.Go(func() error {
			return retrier.New(retrier.ConstantBackoff(3, 100*time.Millisecond), nil).
				RunCtx(ctx, s.Renew)
		})
	}

	return g.Wait()
}

func getAuthMethod(c Config, secretConfig map[string]string) (transport.AuthMethod, error) {
	if c.SSH {
		authMethod, err := ssh.NewSSHAgentAuth("git")
		if err != nil {
			return nil, errors.Wrap(err, "failed to set up SSH authentication")
		}
		return authMethod, nil
	}

	if c.Target.User != "" && c.Target.Pass != "" {
		return &http.BasicAuth{
			Username: c.Target.User,
			Password: c.Target.Pass,
		}, nil
	}

	user, userok := secretConfig["GIT_USERNAME"]
	pass, passok := secretConfig["GIT_PASSWORD"]
	if userok && passok {
		return &http.BasicAuth{
			Username: user,
			Password: pass,
		}, nil
	}

	return nil, nil
}

func getKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

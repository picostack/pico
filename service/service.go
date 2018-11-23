package service

import (
	"context"
	"time"

	"github.com/Southclaws/gitwatch"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"

	"github.com/Southclaws/wadsworth/service/config"
	"github.com/Southclaws/wadsworth/service/task"
)

// App stores application state
type App struct {
	config         Config
	configWatcher  *gitwatch.Session
	targets        []task.Target
	targetsWatcher *gitwatch.Session
	ssh            transport.AuthMethod
	vault          *api.Client
	state          config.State
	ctx            context.Context
	cancel         context.CancelFunc
}

type Config struct {
	Target        string
	Directory     string
	CheckInterval time.Duration
	VaultAddress  string
	VaultToken    string
}

// Initialise prepares an instance of the app to run
func Initialise(ctx context.Context, c Config) (app *App, err error) {
	app = new(App)

	app.ctx, app.cancel = context.WithCancel(ctx)
	app.config = c

	app.ssh, err = ssh.NewSSHAgentAuth("git")
	if err != nil {
		return nil, errors.Wrap(err, "failed to set up SSH authentication")
	}

	vaultConfig := &api.Config{
		Address:    c.VaultAddress,
		HttpClient: cleanhttp.DefaultClient(),
	}
	app.vault, err = api.NewClient(vaultConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to vault")
	}
	app.vault.SetToken(c.VaultToken)

	_, err = app.vault.Logical().List("/secret/metadata")
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping secrets metadata endpoint")
	}

	err = app.reconfigure()
	if err != nil {
		return
	}

	return
}

// Start launches the app and blocks until fatal error
func (app *App) Start() (final error) {
	f := func() (err error) {
		select {
		case <-app.configWatcher.Events:
			err = app.reconfigure()

		case e := <-app.targetsWatcher.Events:
			err = app.handle(e)
		}
		return
	}

	zap.L().Debug("starting service daemon")

	for {
		final = f()
		if final != nil {
			break
		}
	}
	return
}

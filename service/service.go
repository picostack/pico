package service

import (
	"context"
	"time"

	"github.com/Southclaws/gitwatch"
	"go.uber.org/zap"
)

// App stores application state
type App struct {
	config         Config
	configWatcher  *gitwatch.Session
	targets        []string
	targetsWatcher *gitwatch.Session
	ctx            context.Context
	cancel         context.CancelFunc
}

type Config struct {
	Target        string
	Directory     string
	CheckInterval time.Duration
}

// Initialise prepares an instance of the app to run
func Initialise(ctx context.Context, config Config) (app *App, err error) {
	app = new(App)

	app.ctx, app.cancel = context.WithCancel(ctx)
	app.config = config

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

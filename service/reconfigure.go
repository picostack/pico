package service

import (
	"github.com/Southclaws/gitwatch"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// reconfigure will close the configuration watcher and target watcher (unless
// it's the first run) then create a watcher for the application's config target
// repo (the repo that contains the configuration file(s)) then wait for the
// first event (either from a fresh clone, a pull, or just a noop event) then
// read the configuration file(s) from the repository, gather all the targets
// and set up the target watcher. This should always happen in sync with the
// rest of the service to prevent a reconfiguration during an event handler.
func (app *App) reconfigure() (err error) {
	zap.L().Debug("reconfiguring")
	if app.configWatcher != nil {
		zap.L().Debug("closing existing watcher")
		app.configWatcher.Close()
	}

	app.configWatcher, err = gitwatch.New(
		app.ctx,
		[]string{app.config.Target},
		app.config.CheckInterval,
		app.config.Directory,
		nil,
		true)
	if err != nil {
		return errors.Wrap(err, "failed to watch config target")
	}
	go app.configWatcher.Run() //nolint:errcheck - no worthwhile errors returned
	zap.L().Debug("created new watcher, awaiting initial event")

	<-app.configWatcher.InitialDone
	zap.L().Debug("initial event received")

	// read config from repo
	// recreate targets gitwatch
	// diff?

	return
}

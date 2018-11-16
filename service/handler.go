package service

import (
	"github.com/Southclaws/gitwatch"
	"github.com/pkg/errors"
)

// handle takes an event from gitwatch and runs the event's triggers
func (app *App) handle(e gitwatch.Event) (err error) {
	target, exists := app.targets[e.URL]
	if !exists {
		return errors.Errorf("attempt to handle event for unknown target %s at %s", e.URL, e.Path)
	}

	return target.Execute(e.Path, nil, false)
}

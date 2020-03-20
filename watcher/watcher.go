package watcher

import (
	"github.com/picostack/pico/config"
)

// Watcher describes a type that receives a configuration and sets up the
// necessary providers to execute targets. It also provides a way to acquire its
// current state so systems such as the Reconfigurer can inspect it and decide
// if it needs updating.
type Watcher interface {
	SetState(config.State) error
	GetState() config.State
}

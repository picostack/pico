package reconfigurer

import (
	"github.com/picostack/pico/service/watcher"
)

// Provider describes a type that can provide config state events to a target
// watcher. It will reconfigure and restart the watcher whenever necessary.
type Provider interface {
	Configure(watcher.Watcher) error
}

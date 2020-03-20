// Package reconfigurer is responsible for providing configuration state to a
// watcher. It may acquire that state from somewhere such as a Git repository or
// an API. When state needs to be updated, a configuration provider will call
// the Watcher's SetState method. The state transition is then the reponsibility
// of the watcher. All a configuration provider does is acquire configuration
// from some source and ensure it's valid. Providers may store a fallback state
// if validation fails to ensure reliability.
package reconfigurer

import (
	"github.com/picostack/pico/watcher"
)

// Provider describes a type that can provide config state events to a target
// watcher. It will reconfigure and restart the watcher whenever necessary.
type Provider interface {
	Configure(watcher.Watcher) error
}

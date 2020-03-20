package reconfigurer

import (
	"github.com/picostack/pico/config"
	"github.com/picostack/pico/watcher"
)

var _ Provider = &Static{}

// Static implements a Provider with a static config state that only sets its
// watcher state once on initialisation.
type Static struct {
	state config.State
}

// NewStatic creates and calls SetState
func NewStatic(state config.State, w watcher.Watcher) Static {
	s := Static{state: state}
	s.Configure(w)
	return s
}

// Configure implements Provider
func (s *Static) Configure(w watcher.Watcher) error {
	return w.SetState(s.state)
}

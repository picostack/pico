package watcher

import "github.com/picostack/pico/config"

var _ Watcher = &MockWatcher{}

type MockWatcher struct {
	state config.State
}

func (m *MockWatcher) SetState(s config.State) error {
	m.state = s
	return nil
}
func (m *MockWatcher) GetState() config.State {
	return m.state
}

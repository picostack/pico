package watcher

import "github.com/picostack/pico/service/config"

var _ Watcher = &MockWatcher{}

type MockWatcher struct {
	state config.State
}

func (m *MockWatcher) SetState(s config.State) error {
	m.state = s
}
func (m *MockWatcher) GetState() config.State {
	return m.state
}

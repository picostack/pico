package watcher

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/picostack/pico/config"
	"github.com/picostack/pico/task"
)

func TestStateTransitions(t *testing.T) {
	// add target t01
	assert.NoError(t, w.SetState(config.State{
		Targets: []task.Target{{
			Name:    "t01",
			RepoURL: "https://github.com/picostack/pico-example-target",
			Up:      []string{"docker-compose", "up", "-d"},
		}},
		Env: map[string]string{
			"KEY": "VALUE",
		},
	}))
	// add target t02
	assert.NoError(t, w.SetState(config.State{
		Targets: []task.Target{{
			Name:    "t01",
			RepoURL: "https://github.com/picostack/pico-example-target",
			Up:      []string{"docker-compose", "up", "-d"},
		}, {
			Name:    "t02",
			RepoURL: "https://github.com/picostack/pico-example-target",
			Up:      []string{"git", "status"},
		}},
		Env: map[string]string{
			"KEY": "VALUE",
		},
	}))
	// remove target t01
	assert.NoError(t, w.SetState(config.State{
		Targets: []task.Target{{
			Name:    "t02",
			RepoURL: "https://github.com/picostack/pico-example-target",
			Up:      []string{"git", "status"},
		}},
		Env: map[string]string{
			"KEY": "VALUE",
		},
	}))
	// remove target t02
	assert.NoError(t, w.SetState(config.State{
		Targets: []task.Target{},
		Env: map[string]string{
			"KEY": "VALUE",
		},
	}))
	// remove env
	assert.NoError(t, w.SetState(config.State{
		Targets: []task.Target{},
		Env:     map[string]string{},
	}))

	assert.Equal(t, <-bus, task.ExecutionTask{
		Target: task.Target{
			Name:    "t01",
			RepoURL: "https://github.com/picostack/pico-example-target",
			Up:      []string{"docker-compose", "up", "-d"},
		},
		Path:     filepath.Join(".test", "t01"),
		Shutdown: false,
		Env: map[string]string{
			"KEY": "VALUE",
		},
	})
	assert.Equal(t, <-bus, task.ExecutionTask{
		Target: task.Target{
			Name:    "t02",
			RepoURL: "https://github.com/picostack/pico-example-target",
			Up:      []string{"git", "status"},
		},
		Path:     filepath.Join(".test", "t02"),
		Shutdown: false,
		Env: map[string]string{
			"KEY": "VALUE",
		},
	})
	assert.Equal(t, <-bus, task.ExecutionTask{
		Target: task.Target{
			Name:    "t01",
			RepoURL: "https://github.com/picostack/pico-example-target",
			Up:      []string{"docker-compose", "up", "-d"},
		},
		Path:     filepath.Join(".test", "t01"),
		Shutdown: true,
		Env: map[string]string{
			"KEY": "VALUE",
		},
	})
	assert.Equal(t, <-bus, task.ExecutionTask{
		Target: task.Target{
			Name:    "t02",
			RepoURL: "https://github.com/picostack/pico-example-target",
			Up:      []string{"git", "status"},
		},
		Path:     filepath.Join(".test", "t02"),
		Shutdown: true,
		Env: map[string]string{
			"KEY": "VALUE",
		},
	})
}

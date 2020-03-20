package watcher_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/picostack/pico/service/config"
	"github.com/picostack/pico/service/task"
	"github.com/picostack/pico/service/watcher"
)

var w watcher.Watcher
var bus chan task.ExecutionTask

func TestMain(m *testing.M) {
	os.Setenv("DEBUG", "1")
	bus = make(chan task.ExecutionTask, 16)
	gw := watcher.NewGitWatcher(".test", bus, time.Second, nil)

	go func() {
		if err := gw.Start(); err != nil {
			panic(err)
		}
	}()

	w = gw

	os.Exit(m.Run())
}

func TestStateTransitions(t *testing.T) {
	assert.Empty(t, w.GetState())

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

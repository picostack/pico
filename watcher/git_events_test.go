package watcher

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Southclaws/gitwatch"
	"github.com/stretchr/testify/assert"

	"github.com/picostack/pico/config"
	"github.com/picostack/pico/task"
)

func TestGitEvents(t *testing.T) {
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
	// assert receive
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

	assert.NoError(t, w.handle(gitwatch.Event{
		URL:       "https://github.com/picostack/pico-example-target",
		Path:      filepath.Join(".test", "t01"),
		Timestamp: time.Now(),
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
}

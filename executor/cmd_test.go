package executor

import (
	"os"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/picostack/pico/secret/memory"
	"github.com/picostack/pico/task"
	"github.com/stretchr/testify/assert"

	_ "github.com/picostack/pico/logger"
)

func TestMain(m *testing.M) {
	os.Mkdir(".test", os.ModePerm) //nolint:errcheck
	os.Exit(m.Run())
}

func TestCommandExecutor(t *testing.T) {
	ce := NewCommandExecutor(&memory.MemorySecrets{
		Secrets: map[string]map[string]string{
			"test": map[string]string{
				"SOME_SECRET": "123",
			},
		},
	}, false, "pico", "GLOBAL_")
	bus := make(chan task.ExecutionTask)

	g := errgroup.Group{}

	g.Go(func() error {
		bus <- task.ExecutionTask{
			Target: task.Target{
				Name: "test_executor",
				Up:   []string{"git", "init"},
			},
			Path: "./.test",
		}
		return nil
	})

	go ce.Subscribe(bus)

	if err := g.Wait(); err != nil {
		t.Error(err)
	}

	// wait for the task to be consumed and executed
	time.Sleep(time.Second)

	if _, err := os.Stat(".test/.git"); err != nil {
		t.Error("expected path .test/.git to exist:", err)
	}

	os.RemoveAll(".test/.git")
}

func TestCommandPrepareWithoutPassthrough(t *testing.T) {
	ce := NewCommandExecutor(&memory.MemorySecrets{
		Secrets: map[string]map[string]string{
			"test": map[string]string{
				"SOME_SECRET": "123",
			},
		},
	}, false, "pico", "GLOBAL_")

	ex, err := ce.prepare("test", "./", false, nil)
	assert.NoError(t, err)
	assert.Equal(t, exec{
		path: "./",
		env: map[string]string{
			"SOME_SECRET": "123",
		},
		shutdown:        false,
		passEnvironment: false,
	}, ex)
}

func TestCommandPrepareWithGlobal(t *testing.T) {
	ce := NewCommandExecutor(&memory.MemorySecrets{
		Secrets: map[string]map[string]string{
			"test": map[string]string{
				"SOME_SECRET": "123",
			},
			"pico": map[string]string{
				"GLOBAL_SECRET": "456",
				"IGNORE":        "this",
			},
		},
	}, false, "pico", "GLOBAL_")

	ex, err := ce.prepare("test", "./", false, nil)
	assert.NoError(t, err)
	assert.Equal(t, exec{
		path: "./",
		env: map[string]string{
			"SOME_SECRET":   "123",
			"GLOBAL_SECRET": "456",
		},
		shutdown:        false,
		passEnvironment: false,
	}, ex)
}

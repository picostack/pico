package executor_test

import (
	"os"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/picostack/pico/executor"
	"github.com/picostack/pico/secret/memory"
	"github.com/picostack/pico/task"
)

func TestMain(m *testing.M) {
	os.Mkdir(".test", os.ModePerm) //nolint:errcheck
	os.Exit(m.Run())
}

func TestCommandExecutor(t *testing.T) {
	ce := executor.NewCommandExecutor(&memory.MemorySecrets{
		Secrets: map[string]string{
			"SOME_SECRET": "123",
		},
	})
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

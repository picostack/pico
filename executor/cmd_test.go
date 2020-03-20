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
	ce := executor.NewCommandExecutor(&memory.MemorySecrets{})
	bus := make(chan task.ExecutionTask)

	g := errgroup.Group{}

	g.Go(func() error {
		bus <- task.ExecutionTask{
			Target: task.Target{
				Name: "test_executor",
				Up:   []string{"touch", "01"},
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

	if _, err := os.Stat(".test/01"); err == os.ErrNotExist {
		t.Error("expected file .test/01 to exist:", err)
	}
}

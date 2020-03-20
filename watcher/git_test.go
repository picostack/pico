package watcher

import (
	"os"
	"testing"
	"time"

	"github.com/picostack/pico/task"
)

var w *GitWatcher
var bus chan task.ExecutionTask

func TestMain(m *testing.M) {
	os.Setenv("DEBUG", "1")
	bus = make(chan task.ExecutionTask, 16)
	w = NewGitWatcher(".test", bus, time.Second, nil)

	go func() {
		if err := w.Start(); err != nil {
			panic(err)
		}
	}()

	os.Exit(m.Run())
}

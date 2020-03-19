package watcher

import (
	"path/filepath"

	"github.com/Southclaws/gitwatch"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/picostack/pico/service/task"
)

// handle takes an event from gitwatch and runs the event's triggers
func (w *Watcher) handle(e gitwatch.Event) (err error) {
	target, exists := w.getTarget(e.URL)
	if !exists {
		return errors.Errorf("attempt to handle event for unknown target %s at %s", e.URL, e.Path)
	}
	zap.L().Debug("handling event",
		zap.String("target", target.Name),
		zap.String("url", target.RepoURL),
		zap.Time("timestamp", e.Timestamp))
	w.send(target, e.Path, false)
	return nil
}

func (w Watcher) executeTargets(targets []task.Target, shutdown bool) {
	zap.L().Debug("executing all targets", zap.Bool("shutdown", shutdown))
	for _, t := range targets {
		w.send(t, filepath.Join(w.directory, t.Name), shutdown)
	}
	return
}

func (w Watcher) getTarget(url string) (target task.Target, exists bool) {
	for _, t := range w.targets {
		if t.RepoURL == url {
			return t, true
		}
	}
	return
}

func (w Watcher) send(target task.Target, path string, shutdown bool) {
	w.bus <- task.ExecutionTask{
		Target:   target,
		Path:     path,
		Shutdown: shutdown,
	}
}

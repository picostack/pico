package executor

import "github.com/picostack/pico/task"

// Executor describes a type that can handle events and react to them. An
// executor is also responsible for hydrating a target with secrets.
type Executor interface {
	Subscribe(chan task.ExecutionTask)
}

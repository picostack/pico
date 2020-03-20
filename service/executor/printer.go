package executor

import (
	"fmt"

	"github.com/picostack/pico/service/task"
)

// Printer implements an executor that doesn't actually execute, just prints.
type Printer struct{}

// Subscribe implements executor.Executor
func (p *Printer) Subscribe(bus chan task.ExecutionTask) {
	for t := range bus {
		fmt.Printf("received task: %s\n", t.Target.Name)
	}
}

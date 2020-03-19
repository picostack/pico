package task

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"go.uber.org/zap"
)

// ExecutionTask encodes a Target with additional execution-time information.
type ExecutionTask struct {
	Target   Target
	Path     string
	Shutdown bool
}

// Targets is just a list of target objects, to implement the Sort interface
type Targets []Target

// Target represents a repository and the task to perform when that repository
// is updated.
type Target struct {
	// An optional label for the target
	Name string `required:"true" json:"name"`

	// The repository URL to watch for changes, either http or ssh.
	RepoURL string `required:"true" json:"url"`

	// The git branch to use
	Branch string `json:"branch"`

	// The command to run on each new Git commit
	Up []string `required:"true" json:"up"`

	// Down specifies the command to run during either a graceful shutdown or when the target is removed
	Down []string `json:"down"`

	// Environment variables associated with the target - do not store credentials here!
	Env map[string]string `json:"env"`

	// Whether or not to run `Command` on first run, useful if the command is `docker-compose up`
	InitialRun bool `json:"initial_run"`
}

// Execute runs the target's command in the specified directory with the
// specified environment variables
func (t *Target) Execute(dir string, env map[string]string, shutdown bool) (err error) {
	if env == nil {
		env = make(map[string]string)
	}
	for k, v := range t.Env {
		env[k] = v
	}

	var command []string
	if shutdown {
		command = t.Down
	} else {
		command = t.Up
	}

	return execute(dir, env, command)
}

func execute(dir string, env map[string]string, command []string) (err error) {
	if len(command) == 0 {
		return errors.New("attempt to execute target with empty command")
	}

	cmd := exec.Command(command[0])
	if len(command) > 1 {
		cmd.Args = append(cmd.Args, command[1:]...)
	}
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	zap.L().Debug("executing target command",
		zap.String("command", command[0]),
		zap.Strings("args", command[1:]))

	return cmd.Run()
}

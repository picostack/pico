package task

import (
	"fmt"
	"os"
	"os/exec"
)

// Targets is just a list of target objects, to implement the Sort interface
type Targets []Target

// Target represents a repository and the task to perform when that repository
// is updated.
type Target struct {
	// An optional label for the target
	Name string `json:"name"`

	// The repository URL to watch for changes, either http or ssh.
	RepoURL string `required:"true" json:"url"`

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
		return
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

	err = cmd.Run()
	if err != nil {
		return err
	}

	return
}

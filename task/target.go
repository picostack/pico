package task

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

// ExecutionTask encodes a Target with additional execution-time information.
type ExecutionTask struct {
	Target   Target
	Path     string
	Shutdown bool
	Env      map[string]string
}

// Repo represents a Git repo with credentials
type Repo struct {
	URL  string
	User string
	Pass string `json:"-"`
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

	// Auth method to use from the auth store
	Auth string `json:"auth"`
}

// Execute runs the target's command in the specified directory with the
// specified environment variables
func (t *Target) Execute(dir string, env map[string]string, shutdown bool, inheritEnv bool) (err error) {
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

	c, err := prepare(dir, env, command, inheritEnv)
	if err != nil {
		return errors.Wrap(err, "failed to prepare command for execution")
	}

	return c.Run()
}

func prepare(dir string, env map[string]string, command []string, inheritEnv bool) (cmd *exec.Cmd, err error) {
	if len(command) == 0 {
		return nil, errors.New("attempt to execute target with empty command")
	}

	cmd = exec.Command(command[0])
	if len(command) > 1 {
		cmd.Args = append(cmd.Args, command[1:]...)
	}
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	var cmdEnv []string
	if inheritEnv {
		cmdEnv = os.Environ()
	}
	for k, v := range env {
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = cmdEnv

	return cmd, nil
}

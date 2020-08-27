package config

import (
	"os"
	"testing"

	"github.com/robertkrimen/otto"
	"github.com/stretchr/testify/assert"

	"github.com/picostack/pico/task"
)

func Test_applyFileTargets(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		wantTargets task.Targets
		wantErr     bool
	}{
		{"one", `T({
			name: "name",
			url:  "../test.local",
			up:   ["echo", "hello world"]
		})`, task.Targets{
			{
				Name:    "name",
				RepoURL: "../test.local",
				Up:      []string{"echo", "hello world"},
				Env:     map[string]string{},
			},
		}, false},
		{"variable", `
		var url = "https://github.com/Southclaws/";

		T({name: "1", url: url + "project1", up: ["sleep"]});
		T({name: "2", url: url + "project2", up: ["sleep"]});
		T({name: "3", url: url + "project3", up: ["sleep"]});

		console.log("done!");
		`, task.Targets{
			{Name: "1", RepoURL: "https://github.com/Southclaws/project1", Up: []string{"sleep"}, Env: map[string]string{}},
			{Name: "2", RepoURL: "https://github.com/Southclaws/project2", Up: []string{"sleep"}, Env: map[string]string{}},
			{Name: "3", RepoURL: "https://github.com/Southclaws/project3", Up: []string{"sleep"}, Env: map[string]string{}},
		}, false},
		{"auth", `
		var auther = A({
			name: "auth",
			path: "path",
			user_key: "user_key",
			pass_key: "pass_key"
		});

		T({
			name: "name",
			url:  "../test.local",
			up:   ["echo", "hello world"],
			auth: auther,
		});

		console.log("done!");
		`, task.Targets{
			{Name: "name", RepoURL: "../test.local", Up: []string{"echo", "hello world"}, Env: map[string]string{}, Auth: "auth"},
		}, false},
		{"envmap", `
		var url = "https://github.com/Southclaws/";

		T({name: "1", url: url + "project1", up: ["sleep"], env: {PASSWORD: "nope"}});
		T({name: "2", url: url + "project2", up: ["sleep"], env: {PASSWORD: "nope"}});
		T({name: "3", url: url + "project3", up: ["sleep"], env: {PASSWORD: "nope"}});

		console.log("done!");
		`, task.Targets{
			{Name: "1", RepoURL: "https://github.com/Southclaws/project1", Up: []string{"sleep"}, Env: map[string]string{"PASSWORD": "nope"}},
			{Name: "2", RepoURL: "https://github.com/Southclaws/project2", Up: []string{"sleep"}, Env: map[string]string{"PASSWORD": "nope"}},
			{Name: "3", RepoURL: "https://github.com/Southclaws/project3", Up: []string{"sleep"}, Env: map[string]string{"PASSWORD": "nope"}},
		}, false},
		{"envglobal", `
		E("GLOBAL", "readme");
		T({
			name: "name",
			url:  "../test.local",
			up:   ["sleep"],
			env:  {LOCAL: "hi"}
		})
		`, task.Targets{
			{Name: "name", RepoURL: "../test.local", Up: []string{"sleep"}, Env: map[string]string{"GLOBAL": "readme", "LOCAL": "hi"}},
		}, false},
		{"badtype", `T({name: "name", url: "../test.local", up: 1.23})`, task.Targets{}, true},
		{"missingkey", `T({name: "name", url: "../test.local"})`, task.Targets{}, true},
		{"env", `console.log(ENV["TEST_ENV_KEY"])`, task.Targets{}, false},
		{"hostname", `console.log(HOSTNAME)`, task.Targets{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := configBuilder{
				vm:      otto.New(),
				state:   new(State),
				scripts: []string{tt.script},
			}

			os.Setenv("TEST_ENV_KEY", "an environment variable inside the JS vm")

			err := cb.construct("host")
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantTargets, cb.state.Targets)
		})
	}
}

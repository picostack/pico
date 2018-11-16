package config

import (
	"testing"

	"github.com/robertkrimen/otto"
	"github.com/stretchr/testify/assert"

	"github.com/Southclaws/wadsworth/service/task"
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
		})`, Targets{
			{
				Name:    "name",
				RepoURL: "../test.local",
				Up:      []string{"echo", "hello world"},
			},
		}, false},
		{"variable", `
		var url = "https://github.com/Southclaws/";

		T({name: "1", url: url + "project1", up: ["sleep"]});
		T({name: "2", url: url + "project2", up: ["sleep"]});
		T({name: "3", url: url + "project3", up: ["sleep"]});

		console.log("done!");
		`, Targets{
			{Name: "1", RepoURL: "https://github.com/Southclaws/project1", Up: []string{"sleep"}},
			{Name: "2", RepoURL: "https://github.com/Southclaws/project2", Up: []string{"sleep"}},
			{Name: "3", RepoURL: "https://github.com/Southclaws/project3", Up: []string{"sleep"}},
		}, false},
		{"envmap", `
		var url = "https://github.com/Southclaws/";

		T({name: "1", url: url + "project1", up: ["sleep"], env: {PASSWORD: "nope"}});
		T({name: "2", url: url + "project2", up: ["sleep"], env: {PASSWORD: "nope"}});
		T({name: "3", url: url + "project3", up: ["sleep"], env: {PASSWORD: "nope"}});

		console.log("done!");
		`, Targets{
			{Name: "1", RepoURL: "https://github.com/Southclaws/project1", Up: []string{"sleep"}, Env: map[string]string{"PASSWORD": "nope"}},
			{Name: "2", RepoURL: "https://github.com/Southclaws/project2", Up: []string{"sleep"}, Env: map[string]string{"PASSWORD": "nope"}},
			{Name: "3", RepoURL: "https://github.com/Southclaws/project3", Up: []string{"sleep"}, Env: map[string]string{"PASSWORD": "nope"}},
		}, false},
		{"badtype", `T({name: "name", url: "../test.local", up: 1.23})`, Targets{{}}, true},
		{"missingkey", `T({name: "name", url: "../test.local"})`, Targets{{}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := configBuilder{
				vm:      otto.New(),
				state:   new(State),
				scripts: []string{tt.script},
			}

			err := cb.construct()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantTargets, cb.state.Targets)
		})
	}
}

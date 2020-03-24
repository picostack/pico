// Package config defines a configuration engine based on JavaScript. A
// configuration is built from a set of JavaScript source files and executed
// to generate a state object, which is provided to components such as the
// reconfigurer for resolving state changes. JavaScript is used so certain
// common expressions can be re-used, or targets can be conditionally resolved
// based on input variables such as the machine's hostname.
package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"

	"github.com/picostack/pico/task"
)

// State represents a desired system state
type State struct {
	Targets     task.Targets      `json:"targets"`
	AuthMethods []AuthMethod      `json:"auths"`
	Env         map[string]string `json:"env"`
}

// AuthMethod represents a method of authentication for a target
type AuthMethod struct {
	Name    string `json:"name"`     // name of the auth method
	Path    string `json:"path"`     // path within the secret store
	UserKey string `json:"user_key"` // key for username
	PassKey string `json:"pass_key"` // key for password
}

// ConfigFromDirectory searches a directory for configuration files and
// constructs a desired state from the declarations.
func ConfigFromDirectory(dir, hostname string) (state State, err error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		err = errors.Wrap(err, "failed to read config directory")
		return
	}

	sources := []string{}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) == ".js" {
			sources = append(sources, fileToString(filepath.Join(dir, file.Name())))
		}
	}

	cb := configBuilder{
		vm:      otto.New(),
		state:   new(State),
		scripts: sources,
	}

	err = cb.construct(hostname)
	if err != nil {
		return
	}

	state = *cb.state
	return
}

type configBuilder struct {
	vm      *otto.Otto
	state   *State
	scripts []string
}

func (cb *configBuilder) construct(hostname string) (err error) {
	//nolint:errcheck
	cb.vm.Run(`'use strict';
var STATE = {
	targets: [],
	env: {}
};

function T(t) {
	if(t.name === undefined) { throw "target name undefined"; }
	if(t.url === undefined) { throw "target url undefined"; }
	if(t.up === undefined) { throw "target up undefined"; }
	// if(t.down === undefined) { }
	// if(t.env) { }
	// if(t.initial_run) { }
	// if(t.shutdown_command) { }

	STATE.targets.push(t)
}

function E(k, v) {
	STATE.env[k] = v
}

function A(a) {
	if(a.name === undefined) { throw "auth name undefined"; }
	if(a.path === undefined) { throw "auth path undefined"; }
	if(a.user_key === undefined) { throw "auth user_key undefined"; }
	if(a.pass_key === undefined) { throw "auth pass_key undefined"; }

	STATE.auths.push(a);
}
`)

	cb.vm.Set("HOSTNAME", hostname) //nolint:errcheck

	env := make(map[string]string)
	for _, kv := range os.Environ() {
		d := strings.IndexRune(kv, '=')
		env[kv[:d]] = kv[d+1:]
	}
	cb.vm.Set("ENV", env) //nolint:errcheck

	for _, s := range cb.scripts {
		err = cb.applyFileTargets(s)
		if err != nil {
			return
		}
	}

	stateObj, err := cb.vm.Run(`JSON.stringify(STATE)`)
	if err != nil {
		return errors.Wrap(err, "failed to stringify STATE object")
	}
	stateRaw, err := stateObj.ToString()
	if err != nil {
		return errors.Wrap(err, "failed to get string representation of STATE")
	}
	err = json.Unmarshal([]byte(stateRaw), cb.state)

	for i := range cb.state.Targets {
		tmpEnv := cb.state.Targets[i].Env
		cb.state.Targets[i].Env = cb.state.Env
		for k, v := range tmpEnv {
			cb.state.Targets[i].Env[k] = v
		}
	}

	return
}

func (cb *configBuilder) applyFileTargets(script string) (err error) {
	_, err = cb.vm.Run(script)
	if err != nil {
		return
	}

	return
}

func fileToString(path string) (contents string) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	return string(b)
}

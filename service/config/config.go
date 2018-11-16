package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"

	"github.com/Southclaws/wadsworth/service/task"
)

// State represents a desired system state
type State struct {
	Targets task.Targets `json:"targets"`
}

// ConfigFromDirectory searches a directory for configuration files and
// constructs a desired state from the declarations.
func ConfigFromDirectory(dir string) (state State, err error) {
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

	err = cb.construct()
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

func (cb *configBuilder) construct() (err error) {
	//nolint:errcheck
	cb.vm.Run(`'use strict';
var STATE = {
	targets: []
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
`)

	for _, s := range cb.scripts {
		err = cb.applyFileTargets(s)
		if err != nil {
			return
		}
	}

	stateJsonObj, err := cb.vm.Run(`JSON.stringify(STATE)`)
	if err != nil {
		return errors.Wrap(err, "failed to stringify STATE object")
	}
	stateJson, err := stateJsonObj.ToString()
	if err != nil {
		return errors.Wrap(err, "failed to get string representation of STATE")
	}
	err = json.Unmarshal([]byte(stateJson), cb.state)

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

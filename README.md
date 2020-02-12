# Wadsworth

_The [Mister Handy robot][wadsworth] of automation!_

Wadsworth is a git-driven task runner to automate the application of configs.

## Install

### Linux

```sh
curl -s https://raw.githubusercontent.com/Southclaws/wadsworth/master/install.sh | bash
```

### Everything Else

It's primarily a server side tool aimed at Linux servers, so there aren't any install scripts for other platforms. Most
Windows/Mac usage is probably just local testing so just use `go get` for these use-cases.

## Overview

Wadsworth is a little tool for implementing [Git-Ops][git-ops] in single-server environments. It's not a cloud/cluster
tool however it could easily be used as one, but you'd probably be better off using something like
[kube-applier][kube-applier], [Terraform][terraform], [Ansible][ansible] or any of these more "serious" tools.

Instead, Wadsworth aims to be extremely simple. You give it some Git repositories and tell it to run commands when those
Git repositories receive commits and that's about it. It also provides a way of safely passing in credentials from
[Hashicorp's Vault][vault] so you can say goodbye to storing your MySQL password in a .env file!

## Usage

Currently, Wadsworth has a single command: `run` and it takes a single parameter: a Git URL. This Git URL defines the
"Config Repo" which contains Wadsworth configuration files. These configuration files declare where Wadsworth can find
"Target Repos" which are the repos that contain all the stuff you want to automate. The reason Wadsworth is designed
this way instead of just using the target repos to define what Wadsworth should do is 1. to consolidate Wadsworth config
into one place, 2. separate the config of the tools from the applications and 3. keep your target repos clean.

### Configuration

The precursor to Wadsworth used JSON for configuration, this was fine for simple tasks but the ability to provide a
little bit of logic and variables for repetitive configurations is very helpful. Inspired by [StackExchange's
dnscontrol][dnscontrol], Wadsworth uses JavaScript files as configuration. This provides a JSON-like environment with
the added benefit of conditional logic.

Here's a simple example of a configuration that should exist in the Wadsworth config repo that re-deploys a Docker
Compose stack whenever it changes:

```js
T({
  name: "my_app",
  url: "git@github.com:username/my-docker-compose-project",
  up: ["docker-compose", "up", "-d"],
  down: ["docker-compose", "down"]
});
```

You can also specify branches by suffixing the URL with a `#` followed by the branch name:

```js
...
  url: "git@github.com:username/my-docker-compose-project#development",
...
```

#### The `T` Function

The `T` function declares a "Target" which is essentially a Git repository. In this example, the repository
`git@github.com:username/my-docker-compose-project` would contain a `docker-compose.yml` file for some application
stack. Every time you make a change to this file and push it, Wadsworth will pull the new version and run the command
defined in the `up` attribute of the target, which is `docker-compose up -d`.

You can put as many target declarations as you want in the config file, and as many config files as you want in the
config repo. You can also use variables to cut down on repeated things:

```js
var GIT_HOST = "git@github.com:username/";
T({
  name: "my_app",
  url: GIT_HOST + "my-docker-compose-project",
  up: ["docker-compose", "up", "-d"]
});
```

Or, if you have a ton of Docker Compose projects and they all live on the same Git host, why not declare a function that
does all the hard work:

```js
var GIT_HOST = "git@github.com:username/";

function Compose(name) {
  return {
    name: name,
    url: GIT_HOST + name,
    up: ["docker-compose", "up", "-d"]
  };
}

T(Compose("homepage"));
T(Compose("todo-app"));
T(Compose("world-domination-scheme"));
```

The object passed to the `T` function accepts the following keys:

- `name`: The name of the target
- `url`: The Git URL (ssh or https)
- `up`: The command to run on first-run and on changes
- `down`: The command to run when the target is removed
- `env`: Environment variables to pass to the target

#### The `E` Function

The only other function available in the configuration runtime is `E`, this declares an environment variable that will
be passed to the `up` and `down` commands for all targets.

For example:

```js
E("MOUNT_POINT", "/data");
T({ name: "postgres", url: "...", up: "docker-compose", "up", "-d" });
```

This would pass the environment variable `MOUNT_POINT=/data` to the `docker-compose` invocation. This is useful if you
have a bunch of compose configs that all mount data to some path on the machine, you then use
`${MOUNT_POINT}/postgres:/var/lib/postgres/data` as a volume declaration in your `docker-compose.yml`.

[wadsworth]: https://i.imgur.com/RCYbkiq.png
[git-ops]: https://www.weave.works/blog/gitops-operations-by-pull-request
[kube-applier]: https://github.com/box/kube-applier
[terraform]: https://terraform.io
[ansible]: https://ansible.com
[vault]: https://vaultproject.io
[dnscontrol]: https://stackexchange.github.io/dnscontrol/

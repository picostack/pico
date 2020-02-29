# Picobot

_The little git robot of automation!_

[![Build Status](https://travis-ci.org/picostack/picobot.svg?branch=master)](https://travis-ci.org/picostack/picobot)

Picobot is a git-driven task runner to automate the application of configs.

## Overview

Picobot is a little tool for implementing [Git-Ops][git-ops] in single-server environments. It's analogous to
[kube-applier][kube-applier], [Terraform][terraform], [Ansible][ansible] but for automating lone servers that do not
need cluster-level orchestration.

Instead, Picobot aims to be extremely simple. You give it some Git repositories and tell it to run commands when those
Git repositories receive commits and that's about it. It also provides a way of safely passing in credentials from
[Hashicorp's Vault][vault].

## Install

### Linux

```sh
curl -s https://raw.githubusercontent.com/picostack/picobot/master/install.sh | bash
```

Or via Docker:

```sh
docker pull picostack/picobot:v1
```

See the docker section below and the image on [Docker Hub](https://hub.docker.com/r/picostack/picobot).

### Everything Else

It's primarily a server side tool aimed at Linux servers, so there aren't any install scripts for other platforms. Most
Windows/Mac usage is probably just local testing so just use `go get` for these use-cases.

## Usage

Currently, Picobot has a single command: `run` and it takes a single parameter: a Git URL. This Git URL defines the
"Config Repo" which contains Picobot configuration files. These configuration files declare where Picobot can find
"Target Repos" which are the repos that contain all the stuff you want to automate. The reason Picobot is designed
this way instead of just using the target repos to define what Picobot should do is 1. to consolidate Picobot config
into one place, 2. separate the config of the tools from the applications and 3. keep your target repos clean.

Picobot also has a Docker image - see below for docker-specific information.

### Configuration

The precursor to Picobot used JSON for configuration, this was fine for simple tasks but the ability to provide a
little bit of logic and variables for repetitive configurations is very helpful. Inspired by [StackExchange's
dnscontrol][dnscontrol], Picobot uses JavaScript files as configuration. This provides a JSON-like environment with
the added benefit of conditional logic.

Here's a simple example of a configuration that should exist in the Picobot config repo that re-deploys a Docker
Compose stack whenever it changes:

```js
T({
  name: "my_app",
  url: "git@github.com:username/my-docker-compose-project",
  branch: "prod",
  up: ["docker-compose", "up", "-d"],
  down: ["docker-compose", "down"]
});
```

#### The `T` Function

The `T` function declares a "Target" which is essentially a Git repository. In this example, the repository
`git@github.com:username/my-docker-compose-project` would contain a `docker-compose.yml` file for some application
stack. Every time you make a change to this file and push it, Picobot will pull the new version and run the command
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

## Usage as a Docker Container

See the `docker-compose.yml` file for an example and read below for details.

You can run Picobot as a Docker container. If you're using it to deploy Docker containers via compose, this makes the
most sense. This is quite simple and is best done by writing a Docker Compose configuration for Picobot in order to
bootstrap your deployment.

The Picobot image is built on the `docker/compose` image, since most use-cases will use Docker or Compose to deploy
services. This means you must mount the Docker API socket into the container, just like Portainer or cAdvisor or any of
the other Docker tools that also run inside a container.

The socket is located by default at `/var/run/docker.sock` and the `docker/compose` image expects this path too, so you
just need to add a volume mount to your compose that specifies `/var/run/docker.sock:/var/run/docker.sock`.

Another minor detail you should know is that Picobot exposes a `HOSTNAME` variable for the configuration script.
However, when in a container, this hostname is a randomised string such as `b50fa67783ad`. This means, if your
configuration performs checks such as `if (HOSTNAME === 'server031')`, this won't work. To resolve this, Picobot will
attempt to read the environment variable `HOSTNAME` and use that instead of using `/etc/hostname`.

This means, you can bootstrap a Picobot deployment with only two variables:

```env
VAULT_TOKEN=abcxyz
HOSTNAME=server012
```

### Docker Compose and `./` in Container Volume Mounts

Another caveat to running Picobot in a container to execute `docker-compose` is the container filesystem will not
match the host filesystem paths.

If you mount directories from your repository - a common strategy for versioning configuration - `./` will be expanded
by Docker compose running inside the container, but this path may not be valid in the context of the Docker daemon,
which will be running on the host.

The solution to this is both `DIRECTORY: "/cache"` and `/cache:/cache`: as long as the path used in the container also
exists on the host, Docker compose will expand `./` to the same path as the host and everything will work fine.

This also means your config and target configurations will be persisted on the host's filesystem.

<!-- Links -->

[wadsworth]: https://i.imgur.com/RCYbkiq.png
[git-ops]: https://www.weave.works/blog/gitops-operations-by-pull-request
[kube-applier]: https://github.com/box/kube-applier
[terraform]: https://terraform.io
[ansible]: https://ansible.com
[vault]: https://vaultproject.io
[dnscontrol]: https://stackexchange.github.io/dnscontrol/

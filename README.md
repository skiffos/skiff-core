# Skiff Core

[![GoDoc Widget]][GoDoc] [![Go Report Card Widget]][Go Report Card]

[GoDoc]: https://godoc.org/github.com/skiffos/skiff-core
[GoDoc Widget]: https://godoc.org/github.com/skiffos/skiff-core?status.svg
[Go Report Card Widget]: https://goreportcard.com/badge/github.com/skiffos/skiff-core
[Go Report Card]: https://goreportcard.com/report/github.com/skiffos/skiff-core

## Introduction

Skiff Core manages setting up user environment containers on embedded systems. It allows users to work inside familiar environments in a modular and easy to configure way.

Core works by reading a configuration file which defines **containers** and **users**. Core sets up the containers with Docker, creates users in the host system, and redirects SSH logins for those users into the containers as configured.

There are, therefore, two different modes that Core works in:

 - **setup**: at setup time, core reads the configuration, sets up the containers and users, and then exits.
 - **shell**: when SSHing in, core is used as a system shell (in /etc/passwd). Core redirects the IO and commands for the request into an exec session with the container.
 
## Configuration

SkiffOS configures Skiff Core at `/mnt/persist/skiff/core/config.yaml`.

The defaults are configured by the [core package] that you've chosen, the
[default] uses Ubuntu with a minimal desktop environment:

[core package]: https://github.com/skiffos/SkiffOS/tree/master/configs/core
[default]: https://github.com/skiffos/SkiffOS/blob/master/configs/skiff/core/buildroot_ext/package/skiff-core-defconfig/coreenv-defconfig.yaml

```yaml
# see other Skiff packages for more advanced defaults:
#  - core/alpine
#  - core/dietpi
#  - core/gentoo
#  - core/manjaro
#  - core/nixos
#  - core/ubuntu
containers:
  core: # name of the docker container
    image: skiffos/skiff-core-ubuntu:latest
    entrypoint: ["/lib/systemd/systemd"]
    # systemd: indicate this is a container
    env: ["container=docker"]
    stopSignal: RTMIN+3
    tty: true
    disableInit: true
    workingDirectory: /
    mounts:
      - /dev:/dev
      - /etc/resolv.conf:/etc/resolv.conf:ro
      - /etc/hostname:/etc/hostname:ro
      - /lib/modules:/lib/modules:ro
      - /mnt:/mnt
      - /run/udev:/run/udev
      - /mnt/persist/skiff/core/repos/apt:/var/lib/apt
      - /mnt/persist/skiff/core/repos/linux:/usr/src
      - /mnt/persist/skiff/core/repos/log:/var/log
      - /mnt/persist/skiff/core/repos/tmp:/var/tmp
    privileged: true
    startAfterCreate: true
    restartPolicy: "always"
    capAdd:
    - ALL
    hostIPC: true
    hostUTS: true
    hostNetwork: true
    securityOpt:
    - seccomp=unconfined
    tmpFs:
      /run: rw,noexec,nosuid,size=65536k
      /run/lock: rw,noexec,nosuid,size=65536k
users: # can add unlimited users 
  core:
    container: core
    containerUser: core
    containerShell:
    - "/bin/bash"
    auth:
      copyRootKeys: true
images:
  skiffos/skiff-core-ubuntu:latest:
    pull:
      # images are provided for arm64, arm, amd64
      # also an option: policy: ifbuildfails
      policy: ifnotexists
      # avoid docker hub rate limits
      registry: quay.io
    build:
      source: /opt/skiff/coreenv/base
```

The [full config structure] is in the config/ Go files.

[full config structure]: https://github.com/skiffos/skiff-core/blob/master/config/core_config.go#L43

# Skiff Core

[![GoDoc Widget]][GoDoc] [![Go Report Card Widget]][Go Report Card]

[GoDoc]: https://godoc.org/github.com/skiffos/skiff-core
[GoDoc Widget]: https://godoc.org/github.com/skiffos/skiff-core?status.svg
[Go Report Card Widget]: https://goreportcard.com/badge/github.com/skiffos/skiff-core
[Go Report Card]: https://goreportcard.com/report/github.com/skiffos/skiff-core

## Introduction

Skiff Core manages setting up user environment containers on embedded systems. It allows users to work inside familiar environments in a modular and easy to configure way.

Core works by reading a configuration file which defines **images**, **containers**, and **users**. Core sets up the containers with Docker, creates users in the host system, and redirects SSH logins for those users into the containers as configured.

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

[full config structure]: https://github.com/skiffos/skiff-core/blob/master/config/config.go#L43

### Detailed Configuration Reference

The Skiff Core configuration is defined in a YAML file, typically located at `/mnt/persist/skiff/core/config.yaml`. The structure of this file is described below.

#### Top-Level Configuration (`Config`)

The root of the configuration file can contain the following keys:

*   `containers` (`map[string]Container`): Defines named container configurations. Each key is a container name.
*   `users` (`map[string]User`): Defines named user configurations. Each key is a username.
*   `images` (`map[string]Image`): Defines named image configurations for pulling or building Docker images. Each key is an image name (e.g., `skiffos/skiff-core-ubuntu:latest`).

---

#### Container Configuration (`containers.<name>`)

Each entry under `containers` defines a Docker container.

*   `image` (`string`): The name of the image to use for this container. This image must be defined under the top-level `images` section or exist locally/on Docker Hub.
*   `tty` (`bool`, optional): Enable TTY (pseudo-terminal) for the container. Defaults to `false`.
*   `workingDirectory` (`string`, optional): Set the working directory inside the container.
*   `mounts` (`list[string]`, optional): A list of volume mounts, using Docker's colon-separated format (e.g., `/host/path:/container/path:ro`).
*   `disableInit` (`bool`, optional): Disable passing `--init` to the container. Defaults to `false`.
*   `privileged` (`bool`, optional): Run the container in privileged mode. Defaults to `false`.
*   `capAdd` (`list[string]`, optional): List of capabilities to add to the container (e.g., `["SYS_ADMIN"]`). Can also accept `["ALL"]`.
*   `hostIPC` (`bool`, optional): Use the host's IPC namespace. Defaults to `false`.
*   `hostPID` (`bool`, optional): Use the host's PID namespace. Defaults to `false`.
*   `hostUTS` (`bool`, optional): Use the host's UTS namespace. Defaults to `false`.
*   `hostNetwork` (`bool`, optional): Use the host's network stack. Defaults to `false`.
*   `securityOpt` (`list[string]`, optional): List of security options (e.g., `["seccomp=unconfined"]`).
*   `tmpFs` (`map[string]string`, optional): Tmpfs mounts. Keys are container paths, values are options (e.g., `"/run": "rw,noexec,nosuid,size=65536k"`).
*   `entrypoint` (`list[string]`, optional): Override the default entrypoint of the image.
*   `cmd` (`list[string]`, optional): Override the default command of the image.
*   `env` (`list[string]`, optional): A list of environment variables in `KEY=VALUE` format.
*   `ports` (`list[PortMapping]`, optional): Port mappings if not using `hostNetwork: true`.
    *   Each `PortMapping` object has:
        *   `hostPort` (`int`): Port on the host.
        *   `containerPort` (`int`): Port in the container.
*   `dns` (`list[string]`, optional): List of DNS server IP addresses.
*   `dnsSearch` (`list[string]`, optional): List of DNS search domains.
*   `hosts` (`list[string]`, optional): List of additional host entries in `hostname:IP` format.
*   `restartPolicy` (`string`, optional): Restart policy for the container (e.g., `always`, `on-failure`, `never`).
*   `startAfterCreate` (`bool`, optional): Start the container immediately after it's created. Defaults to `false`.
*   `stopSignal` (`string`, optional): Signal to use for stopping the container (e.g., `SIGTERM`, `RTMIN+3`).

---

#### User Configuration (`users.<name>`)

Each entry under `users` defines a system user and how their SSH sessions are handled.

*   `container` (`string`): The name of the container (defined under `containers`) this user's sessions should be directed to.
*   `auth` (`UserAuth`, optional): Authentication settings for the user.
    *   `copyRootKeys` (`bool`, optional): If `true`, copy the host's root user's SSH authorized keys for this user. Defaults to `false`.
    *   `sshKeys` (`list[string]`, optional): A list of public SSH keys (strings) to authorize for this user.
    *   `password` (`string`, optional): Set a password for the user. If empty, password login is typically disabled by setting a long random password.
    *   `allowEmptyPassword` (`bool`, optional): If `true`, allows an empty password (insecure). Defaults to `false`.
    *   `locked` (`bool`, optional): If `true`, the user account will be locked. Defaults to `false`.
*   `containerUser` (`string`, optional): The username to use inside the container when an SSH session starts.
*   `containerShell` (`list[string]`, optional): The shell and its arguments to execute inside the container (e.g., `["/bin/bash"]`).
*   `createContainerUser` (`bool`, optional): If `true`, attempt to create the `containerUser` inside the container if it doesn't exist. Defaults to `false`.

---

#### Image Configuration (`images.<name>`)

Each entry under `images` defines how a Docker image should be obtained, either by pulling or building. The key is the image name (e.g., `ubuntu:latest` or `myrepo/myimage:tag`).

*   `pull` (`ImagePull`, optional): Configuration for pulling the image.
    *   `pullPolicy` (`string`, optional): When to pull the image. Options:
        *   `always`: Always pull the image before container startup.
        *   `ifnotpresent`: Pull only if the image is not present locally (default).
        *   `ifbuildfails`: Pull if a configured build for this image fails.
    *   `registry` (`string`, optional): Specify a custom registry to pull from (e.g., `quay.io`). Defaults to Docker Hub.
*   `build` (`ImageBuild`, optional): Configuration for building the image.
    *   `source` (`string`, optional): Path to the directory containing the build context (source files and Dockerfile). Relative paths are typically resolved based on Skiff Core's configuration directory.
    *   `dockerfile` (`string`, optional): Path to the Dockerfile, relative to the `source` directory. Defaults to `Dockerfile` in the `source` directory.
    *   `root` (`string`, optional): Path to use as the root for the Dockerfile, if files outside the `source` directory are needed.
    *   `buildArgs` (`map[string]*string`, optional): Build-time variables (e.g., `HTTP_PROXY: "http://proxy.example.com"`). A `null` value for a key means the argument is passed without a value (e.g., `MY_FLAG: null` becomes `--build-arg MY_FLAG`).
    *   `preserveIntermediate` (`bool`, optional): If `true`, preserve intermediate build containers. Defaults to `false`.
    *   `squash` (`bool`, optional): If `true`, squash the image layers into a single layer after a successful build. Defaults to `false`.
    *   `scratchBuild` (`bool`, optional, **deprecated**): Previously used for patching image trees for arch-specific images. Defaults to `false`. Modern multi-arch images and Docker manifests are preferred.

# Scratch Build

[![GoDoc Widget]][GoDoc] [![Go Report Card Widget]][Go Report Card]

[GoDoc]: https://godoc.org/github.com/paralin/scratchbuild
[GoDoc Widget]: https://godoc.org/github.com/paralin/scratchbuild?status.svg
[Go Report Card Widget]: https://goreportcard.com/badge/github.com/paralin/scratchbuild
[Go Report Card]: https://goreportcard.com/report/github.com/paralin/scratchbuild

## Introduction

Scratch Build is a small CLI tool that can successfully compile **from scratch** the majority of the Docker library images.

Images are built by traversing the Dockerfile stack down to either scratch or a known working alternative for the target architecture, and then working up to the target.

This is VERY useful for rebuilding docker images on non-intel architectures.

## Getting Started

You use `scratchbuild build` just like `docker build`:

```
$ scratchbuild build -t "myname/myimage:latest-arm" -f Dockerfile .
```

Scratch Build will detect your target architecture (customizable with the flags), traverse the stack of `FROM` declarations, make a plan on how to build the container for the target arch, and then execute the plan in stages.

## Build Quirks

Note: you may have some errors when trying to build. In this case, try disabling CGO:

```
CGO_ENABLED=0 go build -v -o scratchbuild github.com/paralin/scratchbuild/cmd/scratchbuild
```

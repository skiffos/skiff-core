# Skiff Core

[![GoDoc Widget]][GoDoc] [![Go Report Card Widget]][Go Report Card]

[GoDoc]: https://godoc.org/github.com/paralin/skiff-core
[GoDoc Widget]: https://godoc.org/github.com/paralin/skiff-core?status.svg
[Go Report Card Widget]: https://goreportcard.com/badge/github.com/paralin/skiff-core
[Go Report Card]: https://goreportcard.com/report/github.com/paralin/skiff-core

## Introduction

Skiff Core manages setting up user environment containers on embedded systems. It allows users to work inside familiar environments in a modular and easy to configure way.

Core works by reading a configuration file which defines **containers** and **users**. Core sets up the containers with Docker, creates users in the host system, and redirects SSH logins for those users into the containers as configured.

There are, therefore, two different modes that Core works in:

 - **setup**: at setup time, core reads the configuration, sets up the containers and users, and then exits.
 - **shell**: when SSHing in, core is used as a system shell (in /etc/passwd). Core redirects the IO and commands for the request into an exec session with the container.

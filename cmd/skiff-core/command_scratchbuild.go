package main

import (
	sbcli "github.com/paralin/scratchbuild/cli"
	"github.com/urfave/cli"
)

// ScratchBuildCommands define the commands for "scratchbuild"
var ScratchBuildCommands cli.Commands = []cli.Command{
	{
		Name:        "scratchbuild",
		Usage:       "Use the scratch build tool.",
		Subcommands: sbcli.RootCommands,
	},
}

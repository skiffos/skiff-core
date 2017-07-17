package cli

import (
	"github.com/urfave/cli"
)

// RootCommands are the root level commands.
var RootCommands cli.Commands = []cli.Command{
	BuildCommand,
}

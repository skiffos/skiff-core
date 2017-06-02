package main

import (
	"errors"
	"os/user"

	"github.com/mgutz/str"
	"github.com/paralin/skiff-core/shell"
	"github.com/urfave/cli"
)

// ShellCommands define the commands for "shell"
var ShellCommands cli.Commands = []cli.Command{
	{
		Name:  "shell",
		Usage: "Runs skiff-core in shell mode.",
		Action: func(c *cli.Context) error {
			// Check the home directory
			currentUser, err := user.Current()
			if err != nil {
				return err
			}

			if currentUser.HomeDir == "" {
				return errors.New("Cannot determine home directory.")
			}

			sh := shell.NewShell(currentUser.HomeDir)
			var cmd []string
			if globalFlags.Command != "" {
				cmd = str.ToArgv(globalFlags.Command)
			}
			err = sh.Execute(cmd)
			if err != nil {
				return cli.NewExitError(err.Error(), 1)
			}
			return nil
		},
	},
}

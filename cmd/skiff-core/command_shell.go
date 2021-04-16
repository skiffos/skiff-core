package main

import (
	"errors"
	"os/user"

	"github.com/skiffos/skiff-core/shell"
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
			// cmd, if unset, defaults to config.UserShell
			// arg 2, execWithShell, indicates the cmd should be run in user shell.
			// ex: cmd={"rsync", "~/test", "remote:test"}, converts to:
			// docker exec -it container-id /bin/sh -c 'rsync ~/test remote:test'
			err = sh.Execute(globalFlags.Command, true)
			if err != nil {
				return cli.NewExitError(err.Error(), 1)
			}
			return nil
		},
	},
}

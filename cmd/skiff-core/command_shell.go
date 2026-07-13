package main

import (
	"os/user"

	"github.com/aperturerobotics/cli"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skiffos/skiff-core/shell"
)

func buildShellCommand(le *logrus.Entry, args *appArgs) *cli.Command {
	return &cli.Command{
		Name:  "shell",
		Usage: "Run skiff-core in shell mode.",
		Action: func(c *cli.Context) error {
			currentUser, err := user.Current()
			if err != nil {
				return err
			}
			if currentUser.HomeDir == "" {
				return errors.New("cannot determine home directory")
			}

			userShell := shell.NewShell(le, currentUser.HomeDir)
			err = userShell.Execute(c.Context, args.command, true)
			if err == nil {
				return nil
			}

			var exitErr interface{ ExitCode() int }
			if errors.As(err, &exitErr) {
				return cli.Exit("", exitErr.ExitCode())
			}
			return err
		},
	}
}

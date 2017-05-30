package main

import (
	"github.com/paralin/skiff-core/setup"
	"github.com/urfave/cli"
)

var setupArgs struct {
	CreateUsers bool
}

// SetupCommands define the commands for "setup"
var SetupCommands cli.Commands = []cli.Command{
	{
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:        "create-users",
				Usage:       "If set, core will attempt to create missing users.",
				Destination: &setupArgs.CreateUsers,
			},
		},
		Name:  "setup",
		Usage: "Sets up users and containers.",
		Action: func(c *cli.Context) error {
			// read the config
			conf, err := parseGlobalConfig()
			if err != nil {
				return err
			}

			s := setup.NewSetup(conf, setupArgs.CreateUsers)
			return s.Execute()
		},
	},
}

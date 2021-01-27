package main

import (
	"os"
	"strings"

	"github.com/paralin/skiff-core/setup"
	"github.com/urfave/cli"
)

var setupArgs struct {
	CreateUsers bool
	WorkDir     string
}

// SetupCommands define the commands for "setup"
var SetupCommands cli.Commands = []cli.Command{
	{
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:        "create-users",
				Usage:       "If set, core will attempt to create missing users.",
				Destination: &setupArgs.CreateUsers,
				EnvVar:      "SKIFF_CORE_CREATE_USERS",
			},
			cli.StringFlag{
				Name:        "work-dir",
				Usage:       "If set, core will use the directory for working files.",
				Destination: &setupArgs.WorkDir,
				EnvVar:      "SKIFF_CORE_WORK_DIR",
			},
		},
		Name:  "setup",
		Usage: "Sets up users and containers.",
		Action: func(c *cli.Context) error {
			// read the config
			conf, err := parseGlobalConfig()
			if err != nil {
				return cli.NewExitError("Unable to parse config: "+err.Error(), 1)
			}

			setupArgs.WorkDir = strings.TrimSpace(setupArgs.WorkDir)
			if setupArgs.WorkDir != "" {
				if _, err := os.Stat(setupArgs.WorkDir); err != nil {
					if os.IsNotExist(err) {
						// if we created the dir, remove it afterwards.
						defer os.RemoveAll(setupArgs.WorkDir)
					}
					err = os.Mkdir(setupArgs.WorkDir, 0755)
					if err != nil {
						return cli.NewExitError("Unable to create working directory: "+err.Error(), 1)
					}
				}
			}

			s := setup.NewSetup(conf, setupArgs.WorkDir, setupArgs.CreateUsers)

			err = s.Execute()
			if err != nil {
				return cli.NewExitError(err.Error(), 1)
			}
			return nil
		},
	},
}

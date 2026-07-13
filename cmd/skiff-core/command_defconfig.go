package main

import (
	"os"

	"github.com/aperturerobotics/cli"
	"github.com/pkg/errors"
	"github.com/skiffos/skiff-core/config"
)

func buildDefconfigCommand(args *appArgs) *cli.Command {
	return &cli.Command{
		Name:  "defconfig",
		Usage: "Write the default config.",
		Action: func(*cli.Context) error {
			_, err := os.Stat(args.configPath)
			switch {
			case err == nil:
				return errors.Errorf("path already exists, not overwriting: %s", args.configPath)
			case !os.IsNotExist(err):
				return errors.Wrap(err, "inspect config path")
			}

			return args.writeConfig(config.DefaultConfig())
		},
	}
}

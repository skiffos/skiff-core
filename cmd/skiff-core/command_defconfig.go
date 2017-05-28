package main

import (
	"fmt"
	"os"

	"github.com/paralin/skiff-core/config"
	"github.com/urfave/cli"
)

// DefconfigCommands define the commands for "defconfig"
var DefconfigCommands cli.Commands = []cli.Command{
	{
		Name:  "defconfig",
		Usage: "Writes the default config.",
		Action: func(c *cli.Context) error {
			path := globalFlags.ConfigPath
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				return fmt.Errorf("Path %s already exists, not overwriting.", path)
			}

			defConf := config.DefaultConfig()
			return writeGlobalConfig(defConf)
		},
	},
}

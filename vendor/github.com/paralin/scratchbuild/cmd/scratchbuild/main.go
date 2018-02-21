package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	sbcli "github.com/paralin/scratchbuild/cli"
	"github.com/urfave/cli"
)

func main() {
	log.SetLevel(log.DebugLevel)

	app := cli.NewApp()
	app.Name = "scratchbuild"
	app.Description = "Builds Docker images from scratch."
	app.Commands = sbcli.RootCommands
	app.Run(os.Args)
}

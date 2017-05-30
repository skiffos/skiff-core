package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	log.SetLevel(log.DebugLevel)

	app := cli.NewApp()
	app.Name = "scratchbuild"
	app.Description = "Builds Docker images from scratch."
	app.Commands = []cli.Command{
		BuildCommand,
	}
	app.Run(os.Args)
}

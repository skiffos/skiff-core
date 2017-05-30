package main

import (
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/paralin/skiff-core/config"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

var globalFlags struct {
	ConfigPath string
}

func parseGlobalConfig() (*config.Config, error) {
	configData, err := ioutil.ReadFile(globalFlags.ConfigPath)
	if err != nil {
		return nil, err
	}

	res := &config.Config{}
	if err := yaml.Unmarshal(configData, res); err != nil {
		return nil, err
	}
	res.FillPrivateFields()
	res.FillDefaults()
	return res, nil
}

func writeGlobalConfig(conf *config.Config) error {
	data, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(globalFlags.ConfigPath, data, 0644)
}

func main() {
	log.SetLevel(log.DebugLevel)

	app := cli.NewApp()
	app.Author = "Christian Stewart <christian@paral.in>"
	app.Description = "Manages user environment containers."
	app.Commands = append(app.Commands, SetupCommands...)
	app.Commands = append(app.Commands, DefconfigCommands...)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config",
			Usage:       "Config path (.yaml)",
			Destination: &globalFlags.ConfigPath,
			Value:       "config.yaml",
		},
	}
	app.Run(os.Args)
}

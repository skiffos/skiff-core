package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/skiffos/skiff-core/config"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var gitCommit string = "unknown"

var globalFlags struct {
	ConfigPath string
	Command    string
}

func parseGlobalConfig() (*config.Config, error) {
	configData, err := os.ReadFile(globalFlags.ConfigPath)
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
	return os.WriteFile(globalFlags.ConfigPath, data, 0644)
}

func main() {
	log.SetLevel(log.DebugLevel)

	app := cli.NewApp()
	app.Authors = []*cli.Author{{
		Name:  "Christian Stewart",
		Email: "christian@aperture.us",
	}}
	app.Usage = "Manages user environment containers."
	app.Version = gitCommit
	if gitCommit == "unknown" {
		app.HideVersion = true
	}
	app.Commands = append(app.Commands, SetupCommands...)
	app.Commands = append(app.Commands, DefconfigCommands...)
	app.Commands = append(app.Commands, ShellCommands...)
	app.Commands = append(app.Commands, SysInfoCommands...)
	app.Commands = append(app.Commands, ScratchBuildCommands...)
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Usage:       "skiff-core config yaml (.yaml)",
			Destination: &globalFlags.ConfigPath,
			Value:       "config.yaml",
		},
		&cli.StringFlag{
			Name:        "command",
			Aliases:     []string{"c"},
			Usage:       "Command override when calling as a shell.",
			Destination: &globalFlags.Command,
		},
	}
	app.Action = func(c *cli.Context) error {
		if []rune(os.Args[0])[0] != '-' && globalFlags.Command == "" {
			return cli.ShowAppHelp(c)
		}

		// Detected shell mode, execute as shell.
		return ShellCommands[0].Run(c)
	}
	if err := app.Run(os.Args); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

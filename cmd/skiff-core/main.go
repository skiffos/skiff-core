package main

import (
	"context"
	"fmt"
	"os"
	osSignal "os/signal"
	"strings"
	"syscall"

	"github.com/aperturerobotics/cli"
	"github.com/sirupsen/logrus"
	"github.com/skiffos/skiff-core/config"
	"gopkg.in/yaml.v3"
)

// gitCommit is set at build time.
var gitCommit = "unknown"

type appArgs struct {
	configPath string
	command    string
}

func (a *appArgs) parseConfig() (*config.Config, error) {
	configData, err := os.ReadFile(a.configPath)
	if err != nil {
		return nil, err
	}

	result := &config.Config{}
	if err := yaml.Unmarshal(configData, result); err != nil {
		return nil, err
	}
	result.FillPrivateFields()
	result.FillDefaults()
	return result, nil
}

func (a *appArgs) writeConfig(conf *config.Config) error {
	data, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	return os.WriteFile(a.configPath, data, 0o644)
}

func buildApp(le *logrus.Entry) *cli.App {
	args := &appArgs{}
	shellCommand := buildShellCommand(le, args)
	app := &cli.App{
		Name:    "skiff-core",
		Usage:   "Manage user environment containers.",
		Version: gitCommit,
		Authors: []*cli.Author{{
			Name:  "Christian Stewart",
			Email: "christian@aperture.us",
		}},
		Commands: []*cli.Command{
			buildSetupCommand(le, args),
			buildDefconfigCommand(args),
			shellCommand,
			buildSysInfoCommand(),
			buildScratchBuildCommand(le),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Usage:       "skiff-core config YAML.",
				Destination: &args.configPath,
				Value:       "config.yaml",
			},
			&cli.StringFlag{
				Name:        "command",
				Aliases:     []string{"c"},
				Usage:       "Command override when invoked as a login shell.",
				Destination: &args.command,
			},
		},
	}
	app.HideVersion = gitCommit == "unknown"
	app.Action = func(c *cli.Context) error {
		if !strings.HasPrefix(os.Args[0], "-") && args.command == "" {
			return cli.ShowAppHelp(c)
		}
		return shellCommand.Action(c)
	}
	return app
}

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(logger)

	ctx, stop := osSignal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := buildApp(le).RunContext(ctx, os.Args); err != nil {
		if _, writeErr := fmt.Fprintln(os.Stderr, err); writeErr != nil {
			os.Exit(1)
		}
		os.Exit(1)
	}
}

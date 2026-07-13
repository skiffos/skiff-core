package main

import (
	"os"
	"strings"

	"github.com/aperturerobotics/cli"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skiffos/skiff-core/setup"
)

type setupArgs struct {
	createUsers bool
	workDir     string
}

func buildSetupCommand(le *logrus.Entry, appArgs *appArgs) *cli.Command {
	args := &setupArgs{}
	return &cli.Command{
		Name:  "setup",
		Usage: "Set up users and containers.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "create-users",
				Usage:       "Create missing host users.",
				Destination: &args.createUsers,
				EnvVars:     []string{"SKIFF_CORE_CREATE_USERS"},
			},
			&cli.StringFlag{
				Name:        "work-dir",
				Usage:       "Use this directory for temporary build files.",
				Destination: &args.workDir,
				EnvVars:     []string{"SKIFF_CORE_WORK_DIR"},
			},
		},
		Action: func(c *cli.Context) error {
			conf, err := appArgs.parseConfig()
			if err != nil {
				return errors.Wrap(err, "parse config")
			}

			args.workDir = strings.TrimSpace(args.workDir)
			if args.workDir != "" {
				info, err := os.Stat(args.workDir)
				switch {
				case err == nil && !info.IsDir():
					return errors.Errorf("working path is not a directory: %s", args.workDir)
				case err == nil:
				case os.IsNotExist(err):
					if err := os.MkdirAll(args.workDir, 0o755); err != nil {
						return errors.Wrap(err, "create working directory")
					}
					defer func() {
						if err := os.RemoveAll(args.workDir); err != nil {
							le.WithError(err).WithField("path", args.workDir).Warn("remove working directory")
						}
					}()
				default:
					return errors.Wrap(err, "inspect working directory")
				}
			}

			setupRunner := setup.NewSetup(le, conf, args.workDir, args.createUsers)
			return setupRunner.Execute(c.Context)
		},
	}
}

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"
	"github.com/paralin/scratchbuild/arch"
	"github.com/paralin/scratchbuild/builder"
	"github.com/paralin/scratchbuild/library"
	"github.com/paralin/scratchbuild/stack"
	"github.com/urfave/cli"
)

var buildArgs struct {
	Dockerfile string
	Tag        string
	Arch       string
	Cleanup    bool
	CacheDir   string
}

// BuildCommand is the command to build the image.
var BuildCommand = cli.Command{
	Name:  "build",
	Usage: "Builds the image in the specified directory.",
	Action: func(c *cli.Context) (rerr error) {
		defer func() {
			if rerr != nil {
				rerr = cli.NewExitError(rerr.Error(), 1)
			}
		}()
		if c.NArg() != 1 {
			return errors.New("Only one positional argument for the build path is expected.")
		}

		if buildArgs.Tag == "" {
			return errors.New("You must specify a target tag with -t.")
		}

		// parse reference
		if _, err := stack.ParseNormalizedImageName(buildArgs.Tag); err != nil {
			return fmt.Errorf("Tag \"%s\" you specified is invalid: %v", buildArgs.Tag, err)
		}

		// build docker client
		dockerClient, err := client.NewEnvClient()
		if err != nil {
			return err
		}

		targetArch, ok := arch.ParseArch(buildArgs.Arch)
		if !ok {
			log.WithField("arch", buildArgs.Arch).Warn("Unknown architecture, defaulting to amd64")
		}

		buildPath := c.Args().Get(0)
		if _, err := os.Stat(buildPath); err != nil {
			log.WithError(err).Error("Cannot stat build path")
			return err
		}
		log.WithField("path", buildPath).Debug("Using build path")

		if buildArgs.CacheDir == "" {
			buildArgs.Cleanup = true
			cd, err := ioutil.TempDir("", "scratchbuild-")
			if err != nil {
				return err
			}
			buildArgs.CacheDir = cd
		}
		log.WithField("path", buildArgs.CacheDir).Debug("Using cache path")

		if buildArgs.Cleanup {
			defer func() {
				log.WithField("path", buildArgs.CacheDir).Debug("Cleaning up")
				os.RemoveAll(buildArgs.CacheDir)
			}()
		}

		// build image library
		lib, err := library.BuildLibraryResolver(buildArgs.CacheDir)
		if err != nil {
			return err
		}

		// build stack
		stk, err := stack.ImageStackFromPath(buildPath, buildArgs.Dockerfile, buildArgs.Tag, lib, targetArch)
		if err != nil {
			return err
		}

		// rewrite stack
		if targetArch != arch.AMD64 {
			if err := stk.RebaseOnArch(targetArch); err != nil {
				return err
			}
		}

		// execute the plan
		bldr := builder.NewBuilder(stk, dockerClient)
		return bldr.Build()
	},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "tag, t",
			Usage:       "Tag the final image will be labeled with.",
			Destination: &buildArgs.Tag,
		},
		cli.StringFlag{
			Name:        "arch, m",
			Usage:       "The target architecture. Default is to use the system arch.",
			Value:       runtime.GOARCH,
			Destination: &buildArgs.Arch,
		},
		cli.StringFlag{
			Name:        "dockerfile, f",
			Usage:       "The dockerfile to use. Relative to build directory if rel, otherwise absolute.",
			Value:       "Dockerfile",
			Destination: &buildArgs.Dockerfile,
		},
		cli.StringFlag{
			Name:        "cachedir",
			Usage:       "The cache directory used to download source repositories. Default is to use a temporary directory.",
			Destination: &buildArgs.CacheDir,
		},
		cli.BoolTFlag{
			Name:        "cleanup",
			Usage:       "Clean up the cache directory when done. Default is true.",
			Destination: &buildArgs.Cleanup,
		},
	},
}

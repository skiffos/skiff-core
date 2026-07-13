package main

import (
	"os"
	"runtime"

	"github.com/aperturerobotics/cli"
	"github.com/docker/docker/client"
	"github.com/paralin/scratchbuild/arch"
	"github.com/paralin/scratchbuild/library"
	"github.com/paralin/scratchbuild/stack"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	imageBuilder "github.com/skiffos/skiff-core/builder"
)

type scratchBuildArgs struct {
	dockerfile string
	tag        string
	arch       string
	cleanup    bool
	cacheDir   string
}

func buildScratchBuildCommand(le *logrus.Entry) *cli.Command {
	args := &scratchBuildArgs{}
	return &cli.Command{
		Name:  "scratchbuild",
		Usage: "Use the scratch build tool.",
		Subcommands: []*cli.Command{
			{
				Name:      "build",
				Usage:     "Build an image from a directory.",
				ArgsUsage: "<build-path>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "tag",
						Aliases:     []string{"t"},
						Usage:       "Tag for the built image.",
						Destination: &args.tag,
					},
					&cli.StringFlag{
						Name:        "arch",
						Aliases:     []string{"m"},
						Usage:       "Target architecture.",
						Value:       runtime.GOARCH,
						Destination: &args.arch,
					},
					&cli.StringFlag{
						Name:        "dockerfile",
						Aliases:     []string{"f"},
						Usage:       "Dockerfile path relative to the build directory.",
						Value:       "Dockerfile",
						Destination: &args.dockerfile,
					},
					&cli.StringFlag{
						Name:        "cache-dir",
						Usage:       "Cache directory; an empty value uses a temporary directory.",
						Destination: &args.cacheDir,
					},
					&cli.BoolFlag{
						Name:        "cleanup",
						Usage:       "Remove the cache directory after the build.",
						Value:       true,
						Destination: &args.cleanup,
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() != 1 {
						return errors.New("expected one build path")
					}
					if args.tag == "" {
						return errors.New("target tag is required")
					}
					if _, err := stack.ParseNormalizedImageName(args.tag); err != nil {
						return errors.Wrap(err, "parse target tag")
					}

					dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
					if err != nil {
						return err
					}
					defer dockerClient.Close()

					targetArch, ok := arch.ParseArch(args.arch)
					if !ok {
						targetArch = arch.AMD64
						le.WithField("arch", args.arch).Warn("unknown architecture; defaulting to amd64")
					}

					buildPath := c.Args().Get(0)
					if _, err := os.Stat(buildPath); err != nil {
						return errors.Wrap(err, "inspect build path")
					}
					le.WithField("path", buildPath).Debug("using build path")

					if args.cacheDir == "" {
						args.cacheDir, err = os.MkdirTemp("", "scratchbuild-")
						if err != nil {
							return err
						}
					}
					le.WithField("path", args.cacheDir).Debug("using cache path")
					if args.cleanup {
						defer func() {
							if err := os.RemoveAll(args.cacheDir); err != nil {
								le.WithError(err).WithField("path", args.cacheDir).Warn("remove cache directory")
							}
						}()
					}

					libraryResolver, err := library.BuildLibraryResolver(args.cacheDir)
					if err != nil {
						return err
					}
					imageStack, err := stack.ImageStackFromPath(
						buildPath,
						args.dockerfile,
						args.tag,
						libraryResolver,
						targetArch,
					)
					if err != nil {
						return err
					}
					if targetArch != arch.AMD64 {
						if err := imageStack.RebaseOnArch(targetArch); err != nil {
							return err
						}
					}

					builder := imageBuilder.NewScratchBuilder(le, imageStack, dockerClient)
					return builder.Build(c.Context)
				},
			},
		},
	}
}

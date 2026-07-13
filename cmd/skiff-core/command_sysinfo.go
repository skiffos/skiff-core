package main

import (
	"fmt"
	"runtime"

	"github.com/aperturerobotics/cli"
	"github.com/paralin/scratchbuild/arch"
)

func buildSysInfoCommand() *cli.Command {
	return &cli.Command{
		Name:  "sysinfo",
		Usage: "Print detected system information.",
		Action: func(c *cli.Context) error {
			if _, err := fmt.Fprintf(c.App.Writer, "GOARCH: %s\n", runtime.GOARCH); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(c.App.Writer, "GOOS: %s\n", runtime.GOOS); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(c.App.Writer, "GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0)); err != nil {
				return err
			}

			knownArch, ok := arch.ParseArch(runtime.GOARCH)
			if ok {
				_, err := fmt.Fprintf(c.App.Writer, "Detected arch: %v\n", knownArch)
				return err
			}
			_, err := fmt.Fprintf(c.App.Writer, "Unknown arch, defaulting to %v\n", knownArch)
			return err
		},
	}
}

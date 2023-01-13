package main

import (
	"fmt"
	"runtime"

	"github.com/paralin/scratchbuild/arch"
	"github.com/urfave/cli/v2"
)

// DefconfigCommands define the commands for "defconfig"
var SysInfoCommands cli.Commands = []cli.Command{
	{
		Name:  "sysinfo",
		Usage: "Prints detected system information.",
		Action: func(c *cli.Context) error {
			fmt.Printf("GOARCH: %s\n", runtime.GOARCH)
			fmt.Printf("GOOS: %s\n", runtime.GOOS)
			fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
			ka, ok := arch.ParseArch(runtime.GOARCH)
			if ok {
				fmt.Printf("Detected arch: %v\n", ka)
			} else {
				fmt.Printf("Unknown arch, defaulting to %v\n", ka)
			}
			return nil
		},
	},
}

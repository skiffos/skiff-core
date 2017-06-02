package shell

import (
	"fmt"
	"os"
)

// preservedEnvVars are environment variables to include in the shell.
var preservedEnvVars []string = []string{
	"SSH_CONNECTION",
	"SSH_CLIENT",
	"SSH_TTY",
	"TERM",
	"SHLVL",
}

// buildShellEnv loads ssh-related environment variables.
func buildShellEnv() []string {
	var env []string
	for _, name := range preservedEnvVars {
		val, ok := os.LookupEnv(name)
		if ok {
			env = append(env, fmt.Sprintf("%s=%s", name, val))
		}
	}
	return env
}

package shell

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/docker/docker/api"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/hpcloud/tail"
	"github.com/mgutz/str"
	"github.com/skiffos/skiff-core/config"
	"github.com/skiffos/skiff-core/util/execcmd"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Shell holds an instance of a user's interaction with a Docker container.
type Shell struct {
	homeDir string
}

// NewShell builds a new shell instance.
func NewShell(homeDir string) *Shell {
	return &Shell{homeDir: homeDir}
}

// buildDockerClient builds the docker client.
func (s *Shell) buildDockerClient() (client.APIClient, error) {
	return client.NewClient(client.DefaultDockerHost, api.DefaultVersion, nil, nil)
}

// loadUserConfig loads the information for this user.
func (s *Shell) loadUserConfig(configPath string) (*config.ConfigUserShell, error) {
	cf, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer cf.Close()

	data, err := ioutil.ReadAll(cf)
	if err != nil {
		return nil, err
	}
	return config.UnmarshalConfigUserShell(data)
}

// defaultShell is the default shell to use if nothing else is found.
var defaultShell = []string{"/bin/sh"}

// buildTargetCmd builds the full command to pass to docker, wrapping with shell.
func (s *Shell) buildTargetCmd(
	userConfig *config.ConfigUserShell,
	inputCmd string,
	execWithShell bool,
) ([]string, error) {
	var targetCmd []string
	userShell := userConfig.Shell
	if len(userShell) == 0 {
		userShell = defaultShell
	}

	// Setup the command based on the given.
	if len(inputCmd) != 0 {
		if execWithShell {
			targetCmd = make([]string, len(userShell)+2)
			copy(targetCmd, userShell)
			targetCmd[len(targetCmd)-2] = "-c"
			targetCmd[len(targetCmd)-1] = inputCmd
		} else {
			targetCmd = str.ToArgv(inputCmd)
		}
	}

	if len(targetCmd) == 0 {
		targetCmd = userShell // execute shell directly
	}
	return targetCmd, nil
}

// Execute executes the shell, redirecting stdin.
func (s *Shell) Execute(
	inputCmd string,
	execWithShell bool,
) error {
	dockerClient, err := s.buildDockerClient()
	if err != nil {
		return err
	}

	in := execcmd.NewInStream(os.Stdin, true)
	out := execcmd.NewOutStream(os.Stdout)
	errOut := execcmd.NewOutStream(os.Stderr)
	inStrm, _ := in.(*execcmd.InStream)
	useTty := inStrm != nil && inStrm.IsTty()
	outStrm, _ := out.(*execcmd.OutStream)
	// errStrm, _ := errOut.(*execcmd.OutStream)

	configPath := path.Join(s.homeDir, config.UserConfigFile)
	logPath := path.Join(s.homeDir, config.UserLogFile)
	completeCh := make(chan *config.ConfigUserShell, 1)
	checkFiles := func() {
		var err error
		userConfig, err := s.loadUserConfig(configPath)
		if err != nil || userConfig == nil || userConfig.ContainerId == "" {
			userConfig = nil
		}
		if userConfig != nil {
			select {
			case completeCh <- userConfig:
			default:
			}
			return
		}
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	var userConfig *config.ConfigUserShell
	checkFiles()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case userConfig = <-completeCh:
	default:
	}

	if userConfig == nil {
		errOut.Write([]byte("Container setup in progress:\n"))

		logTail, err := tail.TailFile(logPath, tail.Config{
			ReOpen: true,
			Follow: true,
		})
		if err != nil {
			return errors.Wrap(err, "tail setup logs")
		}
		defer logTail.Cleanup()

		pollTimer := time.NewTicker(time.Millisecond * 500)
		for userConfig == nil {
			select {
			case <-ctx.Done():
				pollTimer.Stop()
				return ctx.Err()
			case <-pollTimer.C:
				checkFiles()
			case line := <-logTail.Lines:
				errOut.Write([]byte(line.Text + "\n"))
			case userConfig = <-completeCh:
			}
		}
		pollTimer.Stop()
	}

	cmd, err := s.buildTargetCmd(userConfig, inputCmd, execWithShell)
	if err != nil {
		return err
	}

	// Probe the state of the container.
	ins, err := dockerClient.ContainerInspect(ctx, userConfig.ContainerId)
	if err != nil {
		return err
	}

	if ins.State == nil || !ins.State.Running {
		errOut.Write([]byte("Starting container " + userConfig.ContainerId + "...\n"))
		if err := execcmd.StartContainer(ctx, dockerClient, userConfig.ContainerId, 0); err != nil {
			if err == context.Canceled {
				return err
			}
			logsCloser, lerr := dockerClient.ContainerLogs(ctx, userConfig.ContainerId, types.ContainerLogsOptions{
				ShowStderr: true,
				ShowStdout: true,
			})
			if lerr == nil {
				_, _ = io.Copy(errOut, logsCloser)
				logsCloser.Close()
			}
			return fmt.Errorf("Unable to start container: %s", err.Error())
		}
	}

	execCreate, err := dockerClient.ContainerExecCreate(ctx, userConfig.ContainerId, types.ExecConfig{
		Tty:  useTty,
		User: userConfig.User,
		Cmd:  cmd,
		Env:  buildShellEnv(),

		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return err
	}

	conn, err := dockerClient.ContainerExecAttach(ctx, execCreate.ID, types.ExecStartCheck{
		Tty: useTty,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	// pipe os.stdin to the connection
	errCh := make(chan error, 1)
	go func() {
		streamer := execcmd.HijackedIOStreamer{
			InputStream:  in,
			OutputStream: out,
			ErrorStream:  errOut,
			Resp:         conn,
			Tty:          useTty,
		}
		if useTty {
			inStrm.SetRawMode()
			defer inStrm.RestoreTerminal()
		}

		errCh <- streamer.Stream(ctx)
	}()

	if useTty && inStrm != nil && inStrm.IsTerminal() && outStrm != nil {
		if err := MonitorTtySize(ctx, dockerClient, outStrm, execCreate.ID, true); err != nil {
			log.WithError(err).Error("Error monitoring TTY size")
		}
	}

	return <-errCh
}

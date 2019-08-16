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
	"github.com/paralin/skiff-core/config"
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

// Execute executes the shell, redirecting stdin.
func (s *Shell) Execute(cmd []string) error {
	dockerClient, err := s.buildDockerClient()
	if err != nil {
		return err
	}

	in := NewInStream(os.Stdin)
	out := NewOutStream(os.Stdout)
	errOut := NewOutStream(os.Stderr)
	useTty := in.IsTty() // Detect if this is necessary?

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

	if len(cmd) == 0 {
		cmd = userConfig.Shell
	}
	if len(cmd) == 0 {
		cmd = []string{"/bin/sh"}
	}

	// Probe the state of the container.
	ins, err := dockerClient.ContainerInspect(ctx, userConfig.ContainerId)
	if err != nil {
		return err
	}

	if ins.State == nil || !ins.State.Running {
		errOut.Write([]byte("Starting container " + userConfig.ContainerId + "...\n"))
		err = dockerClient.ContainerStart(ctx, userConfig.ContainerId, types.ContainerStartOptions{})
		if err != nil {
			return fmt.Errorf("Unable to start container: %s", err.Error())
		}

		// wait for the container to start
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(time.Duration(100) * time.Millisecond):
			}

			ctr, err := dockerClient.ContainerInspect(ctx, userConfig.ContainerId)
			if err != nil {
				return err
			}
			if ctr.State == nil {
				continue
			}
			if ctr.State.Dead ||
				(!ctr.State.Running && !ctr.State.Restarting && ctr.State.Status == "exited") {
				err = fmt.Errorf("Container failed to start with exit code: %d", ctr.State.ExitCode)
				logsCloser, lerr := dockerClient.ContainerLogs(ctx, userConfig.ContainerId, types.ContainerLogsOptions{
					ShowStderr: true,
					ShowStdout: true,
				})
				if lerr == nil {
					io.Copy(errOut, logsCloser)
					logsCloser.Close()
				}
				return err
			}

			health := ctr.State.Health
			if health != nil && (health.Status != "none" && health.Status != "healthy") {
				continue
			}

			if ctr.State.Running {
				break
			}
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
		streamer := hijackedIOStreamer{
			inputStream:  in,
			outputStream: out,
			errorStream:  errOut,
			resp:         conn,
			tty:          useTty,
		}
		if useTty {
			in.SetRawMode()
			defer in.RestoreTerminal()
		}

		errCh <- streamer.stream(ctx)
	}()

	if useTty && in.IsTerminal() {
		if err := MonitorTtySize(ctx, dockerClient, out, execCreate.ID, true); err != nil {
			log.WithError(err).Error("Error monitoring TTY size")
		}
	}

	return <-errCh
}

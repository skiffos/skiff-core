package shell

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/promise"
	"github.com/paralin/skiff-core/config"
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
func (s *Shell) buildDockerClient() (*client.Client, error) {
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

	configPath := path.Join(s.homeDir, config.UserConfigFile)
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			log.Debug("Container setup not complete, try again later.")
		}
		return err
	}

	userConfig, err := s.loadUserConfig(configPath)
	if err != nil {
		return err
	}

	if userConfig.ContainerId == "" {
		return errors.New("Container ID not set, setup failed.")
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	in := NewInStream(os.Stdin)
	out := NewOutStream(os.Stdout)
	errOut := NewOutStream(os.Stderr)

	useTty := in.IsTty() // Detect if this is necessary?
	execCreate, err := dockerClient.ContainerExecCreate(ctx, userConfig.ContainerId, types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          useTty,
		User:         userConfig.User,
		Cmd:          cmd,
	})
	if err != nil {
		return err
	}

	conn, err := dockerClient.ContainerExecAttach(ctx, execCreate.ID, types.ExecConfig{
		Tty:    useTty,
		Detach: false,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	// pipe os.stdin to the connection
	errCh := promise.Go(func() error {
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
		return streamer.stream(ctx)
	})

	if useTty && in.IsTerminal() {
		if err := MonitorTtySize(ctx, dockerClient, out, execCreate.ID, true); err != nil {
			log.WithError(err).Error("Error monitoring TTY size")
		}
	}

	return <-errCh
}

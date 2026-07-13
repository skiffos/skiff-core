package shell

import (
	"context"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/aperturerobotics/fsnotify"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skiffos/skiff-core/config"
	"github.com/skiffos/skiff-core/util/execcmd"
)

// Shell executes a host user's commands inside the configured Docker container.
type Shell struct {
	le      *logrus.Entry
	homeDir string
}

// NewShell constructs a shell for a host user home directory.
func NewShell(le *logrus.Entry, homeDir string) *Shell {
	return &Shell{le: le, homeDir: homeDir}
}

func (s *Shell) buildDockerClient() (client.APIClient, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

func (s *Shell) loadUserConfig(configPath string) (*config.ConfigUserShell, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	return config.UnmarshalConfigUserShell(data)
}

func (s *Shell) waitForUserConfig(
	ctx context.Context,
	configPath string,
	logPath string,
	logOut io.Writer,
) (*config.ConfigUserShell, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrap(err, "construct user config watcher")
	}
	defer watcher.Close()

	if err := watcher.Add(s.homeDir); err != nil {
		return nil, errors.Wrap(err, "watch user home directory")
	}

	loadConfig := func() (*config.ConfigUserShell, bool, error) {
		userConfig, err := s.loadUserConfig(configPath)
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		if err != nil {
			return nil, false, errors.Wrap(err, "load user shell config")
		}
		if userConfig.ContainerID == "" {
			return nil, false, errors.New("user shell config has no container ID")
		}
		return userConfig, true, nil
	}

	if userConfig, ok, err := loadConfig(); err != nil || ok {
		return userConfig, err
	}

	var logInfo os.FileInfo
	var logOffset int64
	copySetupLog := func() error {
		file, err := os.Open(logPath)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			return err
		}
		if logInfo == nil || !os.SameFile(logInfo, info) || info.Size() < logOffset {
			logOffset = 0
		}
		if _, err := file.Seek(logOffset, io.SeekStart); err != nil {
			return err
		}
		n, err := io.Copy(logOut, file)
		logOffset += n
		logInfo = info
		return err
	}

	if err := copySetupLog(); err != nil {
		return nil, errors.Wrap(err, "read setup log")
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil, errors.New("user config watcher stopped")
			}
			return nil, errors.Wrap(err, "watch user shell config")
		case event, ok := <-watcher.Events:
			if !ok {
				return nil, errors.New("user config watcher stopped")
			}

			eventPath := filepath.Clean(event.Name)
			if eventPath == filepath.Clean(logPath) {
				if err := copySetupLog(); err != nil {
					return nil, errors.Wrap(err, "read setup log")
				}
			}
			if eventPath != filepath.Clean(configPath) {
				continue
			}

			userConfig, ready, err := loadConfig()
			if err != nil {
				return nil, err
			}
			if ready {
				return userConfig, nil
			}
		}
	}
}

var defaultShell = []string{"/bin/sh"}

const sftpServerShim = `for p in /usr/lib/openssh/sftp-server /usr/libexec/sftp-server /usr/lib/ssh/sftp-server; do
	if [ -x "$p" ]; then
		exec "$p" "$@"
	fi
done
if command -v sftp-server >/dev/null 2>&1; then
	exec sftp-server "$@"
fi
echo "skiff-core: no sftp-server found in container" >&2
exit 127`

func buildSSHSubsystemCmd(inputCmd string) ([]string, bool, error) {
	inputArgv, err := shellquote.Split(inputCmd)
	if err != nil {
		return nil, false, err
	}
	if len(inputArgv) == 0 {
		return nil, false, nil
	}

	switch path.Base(inputArgv[0]) {
	case "sftp-server", "internal-sftp":
		targetCmd := []string{"/bin/sh", "-c", sftpServerShim, path.Base(inputArgv[0])}
		targetCmd = append(targetCmd, inputArgv[1:]...)
		return targetCmd, true, nil
	default:
		return nil, false, nil
	}
}

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

	if inputCmd != "" {
		subsystemCmd, subsystem, err := buildSSHSubsystemCmd(inputCmd)
		if err != nil {
			return nil, errors.Wrap(err, "parse SSH command")
		}
		if subsystem {
			return subsystemCmd, nil
		}

		if execWithShell {
			targetCmd = make([]string, len(userShell)+2)
			copy(targetCmd, userShell)
			targetCmd[len(targetCmd)-2] = "-c"
			targetCmd[len(targetCmd)-1] = inputCmd
		} else {
			targetCmd, err = shellquote.Split(inputCmd)
			if err != nil {
				return nil, errors.Wrap(err, "parse command")
			}
		}
	}

	if len(targetCmd) == 0 {
		targetCmd = userShell
	}
	return targetCmd, nil
}

// Execute runs a command inside the user's configured Docker container.
func (s *Shell) Execute(ctx context.Context, inputCmd string, execWithShell bool) error {
	dockerClient, err := s.buildDockerClient()
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	in := execcmd.NewInStream(os.Stdin, true)
	out := execcmd.NewOutStream(s.le, os.Stdout)
	errOut := execcmd.NewOutStream(s.le, os.Stderr)
	inStrm, _ := in.(*execcmd.InStream)
	useTTY := inStrm != nil && inStrm.IsTTY()
	outStrm, _ := out.(*execcmd.OutStream)

	configPath := filepath.Join(s.homeDir, config.UserConfigFile)
	logPath := filepath.Join(s.homeDir, config.UserLogFile)
	userConfig, err := s.loadUserConfig(configPath)
	if err != nil || userConfig.ContainerID == "" {
		if _, writeErr := errOut.Write([]byte("Container setup in progress:\n")); writeErr != nil {
			return writeErr
		}
		userConfig, err = s.waitForUserConfig(ctx, configPath, logPath, errOut)
		if err != nil {
			return err
		}
	}

	cmd, err := s.buildTargetCmd(userConfig, inputCmd, execWithShell)
	if err != nil {
		return err
	}

	inspection, err := dockerClient.ContainerInspect(ctx, userConfig.ContainerID)
	if err != nil {
		return err
	}
	if inspection.State == nil || !inspection.State.Running {
		if _, err := errOut.Write([]byte("Starting container " + userConfig.ContainerID + "...\n")); err != nil {
			return err
		}
		if err := dockerClient.ContainerStart(ctx, userConfig.ContainerID, container.StartOptions{}); err != nil {
			logs, logErr := dockerClient.ContainerLogs(ctx, userConfig.ContainerID, container.LogsOptions{
				ShowStderr: true,
				ShowStdout: true,
			})
			if logErr == nil {
				if _, copyErr := io.Copy(errOut, logs); copyErr != nil {
					s.le.WithError(copyErr).Debug("copy container logs")
				}
				if closeErr := logs.Close(); closeErr != nil {
					s.le.WithError(closeErr).Debug("close container logs")
				}
			}
			return errors.Wrap(err, "start container")
		}
	}

	execCreate, err := dockerClient.ContainerExecCreate(ctx, userConfig.ContainerID, container.ExecOptions{
		Tty:  useTTY,
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

	conn, err := dockerClient.ContainerExecAttach(ctx, execCreate.ID, container.ExecAttachOptions{
		Tty: useTTY,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	streamer := execcmd.NewHijackedIOStreamer(s.le, conn, useTTY)
	streamer.InputStream = in
	streamer.OutputStream = out
	streamer.ErrorStream = errOut

	if useTTY {
		if err := inStrm.SetRawMode(); err != nil {
			return err
		}
		defer func() {
			if err := inStrm.RestoreTerminal(); err != nil {
				s.le.WithError(err).Warn("restore terminal")
			}
		}()
	}

	monitorCtx, stopMonitor := context.WithCancel(ctx)
	defer stopMonitor()
	if useTTY && inStrm.IsTerminal() && outStrm != nil {
		if err := MonitorTTYSize(monitorCtx, s.le, dockerClient, outStrm, execCreate.ID, true); err != nil {
			return err
		}
	}

	if err := streamer.Stream(ctx); err != nil {
		return err
	}
	stopMonitor()

	return execcmd.InspectExecExit(ctx, dockerClient, execCreate.ID)
}

package setup

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skiffos/skiff-core/config"
	"github.com/skiffos/skiff-core/util/execcmd"
)

var createContainerUserMtx sync.Mutex
var createHostUserMtx sync.Mutex

// UserSetup configures one host user to execute commands in a container.
type UserSetup struct {
	le     *logrus.Entry
	config *config.ConfigUser
	waiter ContainerWaiter
	create bool
	done   chan struct{}
	err    error
}

// NewUserSetup constructs a host-user setup operation.
func NewUserSetup(
	le *logrus.Entry,
	conf *config.ConfigUser,
	waiter ContainerWaiter,
	createUsers bool,
) *UserSetup {
	return &UserSetup{
		le:     le.WithField("user", conf.Name()),
		config: conf,
		waiter: waiter,
		create: createUsers,
		done:   make(chan struct{}),
	}
}

// Execute configures the host user and publishes completion to waiters.
func (u *UserSetup) Execute(ctx context.Context) (executeError error) {
	defer func() {
		u.err = executeError
		close(u.done)
	}()

	if os.Geteuid() != 0 {
		return errors.Errorf("cannot set up user %s without root privileges", u.config.Name())
	}
	conf := u.config
	if conf.Container == "" {
		return errors.Errorf("user must specify a container: %s", conf.Name())
	}
	containerName := ensureSlashPrefix(conf.Container)
	if !u.waiter.CheckHasContainer(containerName) {
		return errors.Errorf("user %s references unknown container %s", conf.Name(), containerName)
	}

	shellPath, err := pathToSkiffCore()
	if err != nil {
		return err
	}
	hostUser, err := func() (*user.User, error) {
		createHostUserMtx.Lock()
		defer createHostUserMtx.Unlock()

		hostUser, lookupErr := user.Lookup(conf.Name())
		if lookupErr != nil {
			var unknownUser user.UnknownUserError
			if !errors.As(lookupErr, &unknownUser) {
				return nil, lookupErr
			}
		}
		if hostUser == nil {
			if !u.create {
				return nil, errors.Errorf("user %s does not exist and create-users is disabled", conf.Name())
			}
			u.le.Debug("creating host user")
			if err := execcmd.ExecCmd(
				ctx,
				"adduser",
				"-G",
				"docker",
				"-D",
				"-s",
				shellPath,
				conf.Name(),
			); err != nil {
				return nil, err
			}
			return user.Lookup(conf.Name())
		}

		u.le.WithField("path", shellPath).Debug("setting host user shell")
		if err := execcmd.ExecCmd(ctx, "chsh", "-s", shellPath, conf.Name()); err != nil {
			return nil, err
		}
		return hostUser, nil
	}()
	if err != nil {
		return err
	}

	uid, err := strconv.Atoi(hostUser.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(hostUser.Gid)
	if err != nil {
		return err
	}

	var nextPassword string
	var allowEmptyPassword bool
	var lockUser bool
	if conf.Auth != nil {
		nextPassword = conf.Auth.Password
		allowEmptyPassword = conf.Auth.AllowEmptyPassword
		lockUser = conf.Auth.Locked
	}
	if lockUser {
		u.le.Debug("locking host user")
		if err := execcmd.ExecCmd(ctx, "passwd", "-l", conf.Name()); err != nil {
			return err
		}
	} else {
		if nextPassword == "" && !allowEmptyPassword {
			u.le.Debug("setting generated host user password")
			nextPassword, err = randomPassword()
			if err != nil {
				return err
			}
		}
		if nextPassword == "" {
			u.le.Debug("setting empty host user password")
			if err := execcmd.ExecCmd(ctx, "passwd", "-d", conf.Name()); err != nil {
				return err
			}
		} else {
			u.le.Debug("setting host user password")
			password := strings.NewReplacer("\r", "", "\n", "").Replace(nextPassword)
			passwordInput := strings.NewReader(password + "\n" + password + "\n")
			cmd := exec.CommandContext(ctx, "passwd", conf.Name())
			cmd.Stdin = passwordInput
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}

	u.le.Debug("setting up SSH keys")
	if err := os.MkdirAll(hostUser.HomeDir, 0o755); err != nil {
		return err
	}
	if err := os.Chown(hostUser.HomeDir, uid, gid); err != nil {
		return err
	}
	sshDir := filepath.Join(hostUser.HomeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return err
	}
	if err := os.Chmod(sshDir, 0o700); err != nil {
		return err
	}
	if err := os.Chown(sshDir, uid, gid); err != nil {
		return err
	}

	authorizedKeysPath := filepath.Join(sshDir, "authorized_keys")
	if err := func() error {
		file, err := os.OpenFile(authorizedKeysPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
		if err != nil {
			return err
		}
		defer file.Close()

		if conf.Auth != nil && conf.Auth.CopyRootKeys {
			rootKeys, err := os.Open("/root/.ssh/authorized_keys")
			if err != nil {
				return err
			}
			if _, err := io.Copy(file, rootKeys); err != nil {
				rootKeys.Close()
				return err
			}
			if err := rootKeys.Close(); err != nil {
				return err
			}
			if _, err := file.WriteString("\n"); err != nil {
				return err
			}
		}
		if conf.Auth != nil {
			for _, key := range conf.Auth.SSHKeys {
				if _, err := file.WriteString(key + "\n"); err != nil {
					return err
				}
			}
		}
		return file.Sync()
	}(); err != nil {
		return err
	}
	if err := os.Chown(authorizedKeysPath, uid, gid); err != nil {
		return err
	}

	setupLogPath := filepath.Join(hostUser.HomeDir, config.UserLogFile)
	setupLog, err := os.OpenFile(setupLogPath, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer setupLog.Close()
	if err := setupLog.Chown(uid, gid); err != nil {
		return err
	}
	if err := setupLog.Sync(); err != nil {
		return err
	}

	containerID, err := u.waiter.WaitForContainer(ctx, containerName, setupLog)
	if err != nil {
		return err
	}
	if conf.ContainerUser != "" && conf.CreateContainerUser {
		if err := u.ensureContainerUser(ctx, containerID); err != nil {
			return err
		}
	}

	userConfigPath := filepath.Join(hostUser.HomeDir, config.UserConfigFile)
	u.le.WithField("path", userConfigPath).Debug("writing user config")
	userConfigData, err := conf.ToConfigUserShell(containerID).Marshal()
	if err != nil {
		return err
	}
	tempConfig, err := os.CreateTemp(hostUser.HomeDir, ".skiff-core-config-*")
	if err != nil {
		return err
	}
	tempConfigPath := tempConfig.Name()
	defer os.Remove(tempConfigPath)
	if err := tempConfig.Chmod(0o640); err != nil {
		tempConfig.Close()
		return err
	}
	if err := tempConfig.Chown(uid, gid); err != nil {
		tempConfig.Close()
		return err
	}
	if _, err := tempConfig.Write(userConfigData); err != nil {
		tempConfig.Close()
		return err
	}
	if err := tempConfig.Sync(); err != nil {
		tempConfig.Close()
		return err
	}
	if err := tempConfig.Close(); err != nil {
		return err
	}
	return os.Rename(tempConfigPath, userConfigPath)
}

func (u *UserSetup) ensureContainerUser(ctx context.Context, containerID string) error {
	createContainerUserMtx.Lock()
	defer createContainerUserMtx.Unlock()

	var stderr bytes.Buffer
	checkErr := u.waiter.ExecCmdContainer(
		ctx,
		containerID,
		"root",
		nil,
		nil,
		&stderr,
		"id",
		u.config.ContainerUser,
	)
	if checkErr == nil {
		return nil
	}
	if !strings.HasSuffix(strings.TrimSpace(stderr.String()), "no such user") {
		return errors.Wrap(checkErr, "check container user")
	}

	u.le.WithFields(logrus.Fields{
		"container-id":   containerID,
		"container-user": u.config.ContainerUser,
	}).Debug("creating container user")
	if err := u.waiter.ExecCmdContainer(
		ctx,
		containerID,
		"root",
		nil,
		os.Stderr,
		os.Stderr,
		"useradd",
		u.config.ContainerUser,
	); err != nil {
		return errors.Wrap(err, "create container user")
	}
	return nil
}

// Wait waits for user setup completion or context cancellation.
func (u *UserSetup) Wait(ctx context.Context, _ io.Writer) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-u.done:
		return u.err
	}
}

// _ is a type assertion.
var _ SetupJob = ((*UserSetup)(nil))

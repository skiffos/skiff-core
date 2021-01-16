package setup

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/paralin/skiff-core/config"
	log "github.com/sirupsen/logrus"
)

// globalCreateContainerUserMtx prevents creating multiple users simultaneously.
var globalCreateContainerUserMtx sync.Mutex

// UserSetup sets up a container.
type UserSetup struct {
	config *config.ConfigUser
	waiter ContainerWaiter
	create bool

	wg  sync.WaitGroup
	err error
}

// NewUserSetup creates a new UserSetup.
func NewUserSetup(config *config.ConfigUser, waiter ContainerWaiter, createUsers bool) *UserSetup {
	return &UserSetup{config: config, waiter: waiter, create: createUsers}
}

// Execute starts the user setup.
func (cs *UserSetup) Execute() (execError error) {
	cs.wg.Add(1)
	defer func() {
		cs.err = execError
		cs.wg.Done()
	}()

	// check if we are root
	if os.Geteuid() != 0 {
		return fmt.Errorf("Not running as root, cannot setup user %s", cs.config.Name())
	}

	conf := cs.config
	if conf.Container == "" {
		return fmt.Errorf("User %s must have container specified.", conf.Name())
	}

	conf.Container = ensureSlashPrefix(conf.Container)
	if !cs.waiter.CheckHasContainer(conf.Container) {
		return fmt.Errorf("User %s: no such container: %s", conf.Name(), conf.Container)
	}

	euser, eusererr := user.Lookup(conf.Name())
	if eusererr != nil {
		if _, ok := eusererr.(user.UnknownUserError); !ok {
			return eusererr
		}
	}

	le := log.WithField("user", conf.Name())
	shellPath, err := pathToSkiffCore()
	if err != nil {
		return err
	}
	if euser == nil {
		if !cs.create {
			return fmt.Errorf("User %s: not found, and create-users is not enabled.", conf.Name())
		}

		// attempt to create the user
		le.Debug("Creating user")
		err = execCmd(
			"adduser",
			"-G",
			"docker",
			"-D",
			//"-c",
			// fmt.Sprintf("Skiff-Core user %s", cs.config.Name()),
			// "-m",
			"-s",
			shellPath,
			cs.config.Name(),
		)
		if err != nil {
			return err
		}
		euser, err = user.Lookup(cs.config.Name())
		if err != nil {
			return err
		}
	} else {
		// Set the shell for the user
		le.WithField("path", shellPath).Debug("Setting shell")
		if err := execCmd("chsh", "-s", shellPath, cs.config.Name()); err != nil {
			return err
		}

		// Add to the Docker group
		/*
			le.WithField("path", shellPath).Debug("Adding to Docker group")
			if err := execCmd("chsh", "-s", shellPath, cs.config.Name()); err != nil {
				return err
			}
		*/
	}

	uid, err := strconv.Atoi(euser.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(euser.Gid)
	if err != nil {
		return err
	}

	// Set password
	var nextPassword string
	var allowEmptyPassword bool
	var lock bool
	if cs.config.Auth != nil {
		nextPassword = cs.config.Auth.Password
		allowEmptyPassword = cs.config.Auth.AllowEmptyPassword
		lock = cs.config.Auth.Locked
	}

	if lock {
		le.Debug("Locking user")
		if err := execCmd("passwd", "-l", cs.config.Name()); err != nil {
			le.WithError(err).Warn("error while locking user")
			return err
		}
	} else {
		if nextPassword == "" && !allowEmptyPassword {
			le.Debug("Setting password to a long random value due to AllowEmptyPassword=false")
			nextPassword = randomPassword()
		}
		if nextPassword == "" {
			le.Debug("Disabling password for user (setting to empty password)")
			if err := execCmd("passwd", "-d", cs.config.Name()); err != nil {
				le.WithError(err).Warn("error while unsetting user password")
				// return err
			}
		} else {
			le.Debug("Setting password")
			passwd := strings.Replace(nextPassword, "\n", "", -1)
			passwordSet := strings.NewReader(fmt.Sprintf("%s\n%s\n", passwd, passwd))
			cmd := exec.Command("passwd", cs.config.Name())
			cmd.Stdin = passwordSet
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}

	le.Debug("Setting up SSH keys")
	sshDir := path.Join(euser.HomeDir, ".ssh")
	authorizedKeysPath := path.Join(sshDir, "authorized_keys")
	if _, err := os.Stat(euser.HomeDir); os.IsNotExist(err) {
		if err := os.MkdirAll(euser.HomeDir, 0755); err != nil {
			return err
		}
		if err := os.Chown(euser.HomeDir, uid, gid); err != nil {
			return err
		}
	}
	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		if err := os.MkdirAll(sshDir, 0700); err != nil {
			return err
		}
	}
	if err := os.Chmod(sshDir, 0700); err != nil {
		return err
	}
	if err := os.Chown(sshDir, uid, gid); err != nil {
		return err
	}

	f, err := os.OpenFile(authorizedKeysPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	authConf := cs.config.Auth
	if authConf != nil {
		if authConf.CopyRootKeys {
			rkf, err := os.Open("/root/.ssh/authorized_keys")
			if err != nil {
				return err
			}
			_, err = io.Copy(f, rkf)
			rkf.Close()
			if err != nil {
				return err
			}
			f.WriteString("\n")
		}
		for _, key := range authConf.SSHKeys {
			f.WriteString(key)
			f.WriteString("\n")
		}
	}

	f.Sync()
	f.Close()
	if err := os.Chown(authorizedKeysPath, uid, gid); err != nil {
		return err
	}

	setupPath := path.Join(euser.HomeDir, config.UserLogFile)
	logFile, err := os.OpenFile(setupPath, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer logFile.Close()
	logFile.Sync()
	logFile.Chown(uid, gid)

	containerId, err := cs.waiter.WaitForContainer(cs.config.Container, logFile)
	if err != nil {
		return err
	}

	if conf.ContainerUser != "" && conf.CreateContainerUser {
		globalCreateContainerUserMtx.Lock()
		// Check if user exists.
		var outp bytes.Buffer
		err := cs.waiter.ExecCmdContainer(
			containerId,
			"root",
			nil, nil, &outp, // catch stderr only
			"id", conf.ContainerUser,
		)
		errStr := strings.TrimSpace(outp.String())
		if strings.HasSuffix(errStr, "no such user") {
			ule := le.
				WithField("container-user", conf.ContainerUser).
				WithField("container-id", containerId)
			ule.Debug("Creating container user...")
			err = cs.waiter.ExecCmdContainer(
				containerId, "root",
				nil, os.Stderr, os.Stderr,
				"useradd", conf.ContainerUser,
			)
			if err != nil {
				ule.
					WithError(err).
					Warn("Unable to create container user")
			}
		}

		globalCreateContainerUserMtx.Unlock()
	}

	userConfPath := path.Join(euser.HomeDir, config.UserConfigFile)
	le.WithField("path", userConfPath).Debug("Writing user config...")
	userConf := cs.config.ToConfigUserShell(containerId)
	userConfData, err := userConf.Marshal()
	if err != nil {
		return err
	}

	userConfFile, err := os.OpenFile(userConfPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	if _, err := userConfFile.Write(userConfData); err != nil {
		return err
	}
	userConfFile.Close()

	return os.Chown(userConfPath, uid, gid)
}

// Wait waits for Execute() to finish.
func (i *UserSetup) Wait(io.Writer) error {
	i.wg.Wait()
	return i.err
}

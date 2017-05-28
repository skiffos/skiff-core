package setup

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/paralin/skiff-core/config"
)

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

	if !cs.waiter.CheckHasContainer(conf.Name()) {
		return fmt.Errorf("User %s: no such container: %s", conf.Name(), conf.Container)
	}

	euser, eusererr := user.Lookup(conf.Name())
	if eusererr != nil {
		if _, ok := eusererr.(user.UnknownUserError); !ok {
			return eusererr
		}
	}

	le := log.WithField("user", conf.Name())
	if euser == nil {
		if !cs.create {
			return fmt.Errorf("User %s: not found, and create-users is not enabled.", conf.Name())
		}

		shellPath, err := pathToSkiffCore()
		if err != nil {
			return err
		}
		// attempt to create the user
		le.Debug("Creating user")
		err = execCmd(
			"adduser",
			"-D",
			"-c",
			fmt.Sprintf("Skiff-Core user %s", cs.config.Name()),
			"-m",
			"-s",
			shellPath,
		)
		if err != nil {
			return err
		}
		euser, err = user.Lookup(cs.config.Name())
		if err != nil {
			return err
		}
	}

	// Set password
	if cs.config.Auth == nil || cs.config.Auth.Password == "" {
		log.Debug("Disabling password login")
		if err := execCmd("passwd", "-d", cs.config.Auth.Password); err != nil {
			return err
		}
	} else {
		log.Debug("Setting password")
		passwd := strings.Replace(cs.config.Auth.Password, "\n", "", -1)
		passwordSet := strings.NewReader(fmt.Sprintf("%s\n%s\n", passwd, passwd))
		cmd := exec.Command("passwd", cs.config.Name())
		cmd.Stdin = passwordSet
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	log.Debug("Setting up SSH keys")
	sshDir := path.Join(euser.HomeDir, ".ssh")
	authorizedKeysPath := path.Join(sshDir, "authorized_keys")
	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		if err := os.Mkdir(sshDir, 0700); err != nil {
			return err
		}
	}
	if err := os.Chmod(sshDir, 0700); err != nil {
		return err
	}

	f, err := os.OpenFile(authorizedKeysPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

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

	return nil
}

// Wait waits for Execute() to finish.
func (i *UserSetup) Wait() error {
	i.wg.Wait()
	return i.err
}

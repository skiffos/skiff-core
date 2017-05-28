package setup

import (
	"context"
	"fmt"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/paralin/skiff-core/builder"
	"github.com/paralin/skiff-core/config"
	"golang.org/x/crypto/ssh/terminal"
)

// ImageSetup is responsible for setting up an image.
type ImageSetup struct {
	config *config.ConfigImage

	err error
	wg  sync.WaitGroup
}

// NewImageSetup builds a new ImageSetup.
func NewImageSetup(conf *config.ConfigImage) *ImageSetup {
	return &ImageSetup{config: conf}
}

// checkImageExists checks if an image exists on the machine.
func (i *ImageSetup) checkImageExists(dockerClient *client.Client, ref string) (bool, error) {
	summaries, err := dockerClient.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return false, err
	}

	for _, summ := range summaries {
		for _, name := range summ.RepoTags {
			if name == ref {
				return true, nil
			}
		}
	}
	return false, nil
}

// pull attempts to pull.
func (i *ImageSetup) pull(dockerClient *client.Client) (pullError error) {
	isTerminal := terminal.IsTerminal(int(os.Stdout.Fd()))
	conf := i.config.Pull
	ref := conf.ImageName()
	if conf.Registry != "" {
		ref = fmt.Sprintf("%s/%s", conf.Registry, ref)
	}
	defer func() {
		if pullError != nil {
			log.WithError(pullError).WithField("ref", ref).Error("Cannot pull")
		}
	}()
	rc, err := dockerClient.ImagePull(context.Background(), ref, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	return jsonmessage.DisplayJSONMessagesStream(rc, os.Stdout, os.Stdout.Fd(), isTerminal, nil)
}

// build attempts to build the image.
func (i *ImageSetup) build() (buildError error) {
	bc := i.config.Build
	defer func() {
		if buildError != nil {
			log.WithError(buildError).Error("Cannot build")
		}
	}()

	bldr, err := builder.NewBuilder(bc)
	if err != nil {
		return err
	}

	return bldr.Build()
}

// Execute executes the setup.
func (i *ImageSetup) Execute() (exError error) {
	i.wg.Add(1)
	defer func() {
		i.err = exError
		i.wg.Done()
	}()

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	if i.config.Pull == nil && i.config.Build == nil {
		exists, err := i.checkImageExists(dockerClient, i.config.Pull.ImageName())
		if err != nil {
			return nil
		}
		if !exists {
			return fmt.Errorf("Image %s not found and no pull or build config specified.", i.config.Name())
		}
	}

	var postBuildPull bool
	if i.config.Pull != nil {
		if i.config.Pull.Policy == config.ConfigPullPolicy_IfBuildFails {
			postBuildPull = true
		} else if i.config.Pull.Policy == config.ConfigPullPolicy_Always {
			err := i.pull(dockerClient)
			if err == nil {
				return nil
			}
		} else if i.config.Pull.Policy == config.ConfigPullPolicy_IfNotPresent {
			exists, err := i.checkImageExists(dockerClient, i.config.Pull.ImageName())
			if err != nil {
				return err
			}
			if exists {
				return nil
			}
		}
	}

	if i.config.Build != nil {
		err := i.build()
		if err != nil {
			if postBuildPull {
				if perr := i.pull(dockerClient); perr != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}

// Wait waits for Execute() to finish.
func (i *ImageSetup) Wait() error {
	i.wg.Wait()
	return i.err
}

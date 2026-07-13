package config

import "strings"

// Config describes skiff-core images, containers, and host users.
type Config struct {
	// containers are keyed by container name
	Containers map[string]*ConfigContainer `json:"containers" yaml:"containers"`
	// users are keyed by host user name
	Users map[string]*ConfigUser `json:"users" yaml:"users"`
	// images are keyed by Docker image reference
	Images map[string]*ConfigImage `json:"images,omitempty" yaml:"images,omitempty"`
}

// FillDefaults fills omitted configuration values.
func (c *Config) FillDefaults() {
	for _, image := range c.Images {
		if image.Build != nil {
			image.Build.FillDefaults()
		}
		if image.Pull != nil {
			image.Pull.FillDefaults()
		}
	}
}

// FillPrivateFields derives names that are encoded as map keys.
func (c *Config) FillPrivateFields() {
	for name, container := range c.Containers {
		if !strings.HasPrefix(name, "/") {
			name = "/" + name
		}
		container.name = name
	}
	for name, image := range c.Images {
		image.SetName(name)
	}
	for name, user := range c.Users {
		user.name = name
	}
}

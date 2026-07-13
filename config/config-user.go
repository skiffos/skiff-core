package config

// ConfigUser describes a host user backed by a Docker container.
type ConfigUser struct {
	name string

	// container names the user's configured container
	Container string `json:"container,omitempty" yaml:"container,omitempty"`
	// auth configures host authentication
	Auth *ConfigUserAuth `json:"auth,omitempty" yaml:"auth,omitempty"`
	// containerUser selects the user inside the container
	ContainerUser string `json:"containerUser,omitempty" yaml:"containerUser,omitempty"`
	// containerShell selects the command shell inside the container
	ContainerShell []string `json:"containerShell,omitempty" yaml:"containerShell,omitempty"`
	// createContainerUser creates a missing user inside the container
	CreateContainerUser bool `json:"createContainerUser,omitempty" yaml:"createContainerUser,omitempty"`
}

// Name returns the host user name.
func (u *ConfigUser) Name() string {
	return u.name
}

// ToConfigUserShell builds the shell runtime config for a container ID.
func (u *ConfigUser) ToConfigUserShell(containerID string) *ConfigUserShell {
	return &ConfigUserShell{
		ContainerID: containerID,
		User:        u.ContainerUser,
		Shell:       u.ContainerShell,
	}
}

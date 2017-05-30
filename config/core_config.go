package config

import (
	"path"
	"strings"
)

type Config struct {
	Containers map[string]*ConfigContainer `json:"containers" yaml:"containers"`
	Users      map[string]*ConfigUser      `json:"users" yaml:"users"`
	Images     map[string]*ConfigImage     `json:"images,omitempty" yaml:"images,omitempty"`
}

// FillDefaults fills the config with reasonable values where necessary.
func (c *Config) FillDefaults() {
	for _, img := range c.Images {
		if img.Build != nil {
			img.Build.FillDefaults()
		}
		if img.Pull != nil {
			img.Pull.FillDefaults()
		}
	}
}

// FillPrivateFields fills hidden fields on the config.
func (c *Config) FillPrivateFields() {
	for name, container := range c.Containers {
		container.name = name
	}
	for name, image := range c.Images {
		image.name = name
		image.FillChildPrivateFields()
	}
	for name, user := range c.Users {
		user.name = name
	}
}

// ConfigContainer is a container in the system.
type ConfigContainer struct {
	name string // populated by the Config
	// Image, must also exist in images list.
	Image string `json:"image" yaml:"image"`
	// Mounts. Colon separated, see Docker mount style
	Mounts []string `json:"mounts,omitempty" yaml:"mounts,omitempty"`
	// Disable passing --init to the container
	DisableInit bool `json:"disableInit,omitempty" yaml:"disableInit,omitempty"`
	// Privileged container?
	Privileged bool `json:"privileged,omitempty" yaml:"privileged,omitempty"`
	// List of capabilities to add. Can also accept ["ALL"]
	CapAdd []string `json:"capAdd,omitempty" yaml:"capAdd,omitempty"`
	// HostIPC?
	HostIPC bool `json:"hostIPC,omitempty" yaml:"hostIPC,omitempty"`
	// HostPID?
	HostPID bool `json:"hostPID,omitempty" yaml:"hostPID,omitempty"`
	// HostUTS?
	HostUTS bool `json:"hostUTS,omitempty" yaml:"hostUTS,omitempty"`
	// HostNetwork?
	HostNetwork bool `json:"hostNetwork,omitempty" yaml:"hostNetwork,omitempty"`
	// Security options
	SecurityOpt []string `json:"securityOpt,omitempty" yaml:"securityOpt,omitempty"`
	// TmpFs mounts. For example: "/run": "rw,noexec,nosuid,size=65536k"
	TmpFs map[string]string `json:"tmpFs,omitempty" yaml:"tmpFs,omitempty"`
	// Entrypoint override
	Entrypoint []string `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`
	// CMD override
	Cmd []string `json:"cmd,omitempty" yaml:"cmd,omitempty"`
	// Any additional environment variables
	Env []ConfigContainerEnvironmentVariable `json:"env,omitempty" yaml:"env,omitempty"`
	// Ports to expose, if not using net=host
	Ports []ConfigContainerPort `json:"ports,omitempty" yaml:"ports,omitempty"`
	// Additional DNS options
	DNS []string `json:"dns,omitempty" yaml:"dns,omitempty"`
	// DNSSearch - additional DNS search domains.
	DNSSearch []string `json:"dnsSearch,omitempty" yaml:"dnsSearch,omitempty"`
	// Hosts contains additional hosts used by the container
	Hosts []string `json:"hosts,omitempty" yaml:"hosts,omitempty"`
}

// ConfigContainerEnvironmentVariable configures an environment variable.
type ConfigContainerEnvironmentVariable struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

// ConfigContainerPort configures a port mapping for a container.
type ConfigContainerPort struct {
	HostPort      int `json:"hostPort" yaml:"hostPort"`
	ContainerPort int `json:"containerPort" yaml:"containerPort"`
}

// Name returns the container's name.
func (c *ConfigContainer) Name() string {
	if !strings.HasPrefix(c.name, "/") {
		c.name = "/" + c.name
	}
	return c.name
}

// ConfigPullPolicy describes when we should pull this image.
type ConfigPullPolicy string

const (
	ConfigPullPolicy_Always       ConfigPullPolicy = "always"
	ConfigPullPolicy_IfNotPresent ConfigPullPolicy = "ifnotpresent"
	ConfigPullPolicy_IfBuildFails ConfigPullPolicy = "ifbuildfails"
)

// ConfigImage is an image we will pull or build.
type ConfigImage struct {
	// name is the name and tag (skiff/core:latest) of the image.
	name string
	// Pull describes information about pulling this image.
	Pull *ConfigImagePull `json:"pull,omitempty" yaml:"pull,omitempty"`
	// Build describes information about building this image from source.
	Build *ConfigImageBuild `json:"build,omitempty" yaml:"build,omitempty"`
}

// Name gets the name of the ConfigImage.
func (c *ConfigImage) Name() string {
	return c.name
}

// SetName changes the name of the ConfigImage
func (c *ConfigImage) SetName(name string) {
	c.name = name
	c.FillChildPrivateFields()
}

// FillChildPrivateFields fills the child private fields with parent values.
func (c *ConfigImage) FillChildPrivateFields() {
	if c.Build != nil {
		c.Build.imageName = c.name
	}
	if c.Pull != nil {
		c.Pull.imageName = c.name
	}
}

// ConfigImagePull is information about how to pull an image.
type ConfigImagePull struct {
	imageName string
	// PullPolicy describes when we should try to pull the image.
	Policy ConfigPullPolicy `json:"pullPolicy,omitempty" yaml:"pullPolicy,omitempty"`
	// Registry to pull from.
	Registry string `json:"registry,omitempty" yaml:"registry,omitempty"`
}

// FillEmpty fills empty fields
func (c *ConfigImagePull) FillDefaults() {
	if c.Policy == ConfigPullPolicy("") {
		c.Policy = ConfigPullPolicy_IfNotPresent
	}
}

// ImageName returns the imageName
func (c *ConfigImagePull) ImageName() string {
	return c.imageName
}

// ConfigImageBuild is information about how to build an image.
type ConfigImageBuild struct {
	imageName string

	// Source controls where we get the source of this image. Supports file paths and git:// or .git URLs and tarballs.
	Source string `json:"source,omitempty" yaml:"source,omitempty"`
	// Root controls the path inside the repository to use.
	Root string `json:"root,omitempty" yaml:"root,omitempty"`
	// Dockerfile controls the path to the Dockerfile inside the repository/source.
	Dockerfile string `json:"dockerfile,omitempty" yaml:"dockerfile,omitempty"`
}

// ImageName returns the imageName
func (c *ConfigImageBuild) ImageName() string {
	return c.imageName
}

// FillDefaults fills the config with reasonable values.
func (b *ConfigImageBuild) FillDefaults() {
	// Force any ../../ trickery in Dockerfile or Root away
	// Join it with /. -> ./path
	if b.Root != "" {
		b.Root = "." + path.Join("/", b.Root)
	}
	if b.Dockerfile != "" {
		b.Dockerfile = "." + path.Join("/", b.Dockerfile)
	}
}

// ConfigUser is a user in the system.
type ConfigUser struct {
	// name is the name of the user
	name string
	// Container is the ID of the container for this user.
	Container string `json:"container,omitempty" yaml:"container,omitempty"`
	// Auth is the authentication config for the user.
	Auth *ConfigUserAuth `json:"auth,omitempty" yaml:"auth,omitempty"`
}

// Name returns the name of the user.
func (u *ConfigUser) Name() string {
	return u.name
}

// ConfigUserAuth is the user authentication configuration.
type ConfigUserAuth struct {
	// CopyRootKeys indicates we should copy the root's SSH access keys.
	CopyRootKeys bool `json:"copyRootKeys,omitempty" yaml:"copyRootKeys,omitempty"`
	// SSHKeys to allow authentication to the system.
	SSHKeys []string `json:"sshKeys,omitempty" yaml:"sshKeys,omitempty"`
	// Password. If empty, then password auth will be disabled.
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

package config

import (
	"path"
	"strings"

	"gopkg.in/yaml.v3"
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
	// Enable starting with a tty?
	Tty bool `json:"tty,omitempty" yaml:"tty,omitempty"`
	// WorkingDirectory override
	WorkingDirectory string `json:"workingDirectory,omitempty" yaml:"workingDirectory,omitempty"`
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
	// Additional environment variables in Key=Value form
	Env []string `json:"env,omitempty" yaml:"env,omitempty"`
	// Ports to expose, if not using net=host
	Ports []ConfigContainerPort `json:"ports,omitempty" yaml:"ports,omitempty"`
	// Additional DNS options
	DNS []string `json:"dns,omitempty" yaml:"dns,omitempty"`
	// DNSSearch - additional DNS search domains.
	DNSSearch []string `json:"dnsSearch,omitempty" yaml:"dnsSearch,omitempty"`
	// Hosts contains additional hosts used by the container
	Hosts []string `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	// RestartPolicy is the restart policy for the container.
	// Typically uses "always" or "on-failure" or "never"
	RestartPolicy string `json:"restartPolicy,omitempty" yaml:"restartPolicy,omitempty"`
	// StartAfterCreate indicates we should start the container immediately after creating it.
	StartAfterCreate bool `json:"startAfterCreate,omitempty" yaml:"startAfterCreate,omitempty"`
	// StopSignal contains the stop signal to use when stopping the container.
	StopSignal string `json:"stopSignal,omitempty" yaml:"stopSignal,omitempty"`
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
	// BuildArgs contains key/value build arguments to set in the Dockerfile.
	BuildArgs map[string]*string `json:"buildArgs,omitempty" yaml:"buildArgs,omitempty"`
	// PreserveIntermediate indicates we should preserve intermediate build containers.
	PreserveIntermediate bool `json:"preserveIntermediate,omitempty" yaml:"preserveIntermediate,omitempty"`
	// ScratchBuild indicates we need to patch the image tree to use arch-specific images.
	// NOTE: This is deprecated and defaults to false.
	// The "correct way" is to use Docker manifests and multi-arch images.
	ScratchBuild bool `json:"scratchBuild,omitempty" yaml:"scratchBuild,omitempty"`
	// Squash indicates we should squash the results into a single layer before committing.
	Squash bool `json:"squash,omitempty" yaml:"squash,omitempty"`
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
	// ContainerUser is the user to execute as inside the container.
	ContainerUser string `json:"containerUser,omitempty" yaml:"containerUser,omitempty"`
	// ContainerShell is the default shell to execute in the container.
	ContainerShell []string `json:"containerShell,omitempty" yaml:"containerShell,omitempty"`
	// CreateContainerUser indicates to create the container user if it doesn't exist.
	//
	// Note: if this step fails the error is logged but ignored.
	CreateContainerUser bool `json:"createContainerUser,omitempty" yaml:"createContainerUser,omitempty"`
}

// Name returns the name of the user.
func (u *ConfigUser) Name() string {
	return u.name
}

// ToConfigUserShell builds a ConfigUserShell from the ConfigUser.
func (u *ConfigUser) ToConfigUserShell(containerId string) *ConfigUserShell {
	return &ConfigUserShell{
		ContainerId: containerId,
		User:        u.ContainerUser,
		Shell:       u.ContainerShell,
	}
}

// ConfigUserAuth is the user authentication configuration.
type ConfigUserAuth struct {
	// CopyRootKeys indicates we should copy the root's SSH access keys.
	CopyRootKeys bool `json:"copyRootKeys,omitempty" yaml:"copyRootKeys,omitempty"`
	// SSHKeys to allow authentication to the system.
	SSHKeys []string `json:"sshKeys,omitempty" yaml:"sshKeys,omitempty"`
	// Password. If empty, then password will be set to very long random value.
	// This is the most fool-proof way to disable password login for the account.
	// Set AllowEmptyPassword if you want insecure login.
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	// AllowEmptyPassword allows an empty password field.
	AllowEmptyPassword bool `json:"allowEmptyPassword,omitempty" yaml:"allowEmptyPassword,omitempty"`
	// Locked indicates the user should be locked.
	Locked bool `json:"locked,omitempty" yaml:"locked,omitempty"`
}

// ConfigUserShell is the configuration file loaded from the users' home directory.
type ConfigUserShell struct {
	ContainerId string   `json:"containerId" yaml:"containerId"`
	User        string   `json:"user,omitempty" yaml:"user,omitempty"`
	Shell       []string `json:"shell,omitempty" yaml:"shell,omitempty"`
}

// Marshal encodes the user shell config as yaml.
func (s *ConfigUserShell) Marshal() ([]byte, error) {
	return yaml.Marshal(s)
}

// Unmarshal decodes the user shell config from yaml.
func UnmarshalConfigUserShell(data []byte) (*ConfigUserShell, error) {
	res := &ConfigUserShell{}
	if err := yaml.Unmarshal(data, res); err != nil {
		return nil, err
	}
	return res, nil
}

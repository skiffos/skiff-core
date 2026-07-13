package config

// ConfigContainer describes a Docker container managed by skiff-core.
type ConfigContainer struct {
	name string

	// image names the required Docker image
	Image string `json:"image" yaml:"image"`
	// tty allocates a terminal for the container
	Tty bool `json:"tty,omitempty" yaml:"tty,omitempty"`
	// workingDirectory overrides the image working directory
	WorkingDirectory string `json:"workingDirectory,omitempty" yaml:"workingDirectory,omitempty"`
	// mounts contains Docker bind-mount specifications
	Mounts []string `json:"mounts,omitempty" yaml:"mounts,omitempty"`
	// disableInit disables the container init process
	DisableInit bool `json:"disableInit,omitempty" yaml:"disableInit,omitempty"`
	// privileged enables privileged container execution
	Privileged bool `json:"privileged,omitempty" yaml:"privileged,omitempty"`
	// capAdd lists Linux capabilities to add
	CapAdd []string `json:"capAdd,omitempty" yaml:"capAdd,omitempty"`
	// hostIPC shares the host IPC namespace
	HostIPC bool `json:"hostIPC,omitempty" yaml:"hostIPC,omitempty"`
	// hostPID shares the host process namespace
	HostPID bool `json:"hostPID,omitempty" yaml:"hostPID,omitempty"`
	// hostUTS shares the host UTS namespace
	HostUTS bool `json:"hostUTS,omitempty" yaml:"hostUTS,omitempty"`
	// hostNetwork shares the host network namespace
	HostNetwork bool `json:"hostNetwork,omitempty" yaml:"hostNetwork,omitempty"`
	// securityOpt contains Docker security options
	SecurityOpt []string `json:"securityOpt,omitempty" yaml:"securityOpt,omitempty"`
	// tmpFs maps container paths to tmpfs mount options
	TmpFs map[string]string `json:"tmpFs,omitempty" yaml:"tmpFs,omitempty"`
	// entrypoint overrides the image entrypoint
	Entrypoint []string `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`
	// cmd overrides the image command
	Cmd []string `json:"cmd,omitempty" yaml:"cmd,omitempty"`
	// env contains environment variables in key=value form
	Env []string `json:"env,omitempty" yaml:"env,omitempty"`
	// ports maps host ports to container ports
	Ports []ConfigContainerPort `json:"ports,omitempty" yaml:"ports,omitempty"`
	// name servers contains container DNS servers
	DNS []string `json:"dns,omitempty" yaml:"dns,omitempty"`
	// search domains contains additional DNS search domains
	DNSSearch []string `json:"dnsSearch,omitempty" yaml:"dnsSearch,omitempty"`
	// hosts contains additional host-to-address mappings
	Hosts []string `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	// restartPolicy selects a Docker policy such as always or on-failure
	RestartPolicy string `json:"restartPolicy,omitempty" yaml:"restartPolicy,omitempty"`
	// startAfterCreate starts the container immediately after creation
	StartAfterCreate bool `json:"startAfterCreate,omitempty" yaml:"startAfterCreate,omitempty"`
	// stopSignal overrides the image stop signal
	StopSignal string `json:"stopSignal,omitempty" yaml:"stopSignal,omitempty"`
}

// Name returns the normalized Docker container name.
func (c *ConfigContainer) Name() string {
	return c.name
}

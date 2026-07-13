package config

// ConfigContainerPort maps one TCP host port to a container port.
type ConfigContainerPort struct {
	// hostPort selects the host TCP port; zero requests an ephemeral port
	HostPort int `json:"hostPort" yaml:"hostPort"`
	// containerPort is the container TCP port
	ContainerPort int `json:"containerPort" yaml:"containerPort"`
}

package config

import "gopkg.in/yaml.v3"

// ConfigUserShell describes the container command environment for a host user.
type ConfigUserShell struct {
	// containerID is the Docker container ID
	ContainerID string `json:"containerId" yaml:"containerId"`
	// user selects the user inside the container
	User string `json:"user,omitempty" yaml:"user,omitempty"`
	// shell is the argv used for interactive and command shells
	Shell []string `json:"shell,omitempty" yaml:"shell,omitempty"`
}

// Marshal encodes the shell config as YAML.
func (s *ConfigUserShell) Marshal() ([]byte, error) {
	return yaml.Marshal(s)
}

// UnmarshalConfigUserShell decodes a shell config from YAML.
func UnmarshalConfigUserShell(data []byte) (*ConfigUserShell, error) {
	result := &ConfigUserShell{}
	if err := yaml.Unmarshal(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

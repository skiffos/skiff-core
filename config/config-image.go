package config

// ConfigImage describes how an image is pulled or built.
type ConfigImage struct {
	name string

	// pull describes image pull behavior
	Pull *ConfigImagePull `json:"pull,omitempty" yaml:"pull,omitempty"`
	// build describes image build behavior
	Build *ConfigImageBuild `json:"build,omitempty" yaml:"build,omitempty"`
}

// Name returns the Docker image reference.
func (c *ConfigImage) Name() string {
	return c.name
}

// SetName sets the Docker image reference and propagates it to child configs.
func (c *ConfigImage) SetName(name string) {
	c.name = name
	if c.Build != nil {
		c.Build.imageName = name
	}
	if c.Pull != nil {
		c.Pull.imageName = name
	}
}

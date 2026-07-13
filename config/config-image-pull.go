package config

// ConfigImagePull describes how to pull a Docker image.
type ConfigImagePull struct {
	imageName string

	// policy controls when the image is pulled
	Policy ConfigPullPolicy `json:"pullPolicy,omitempty" yaml:"pullPolicy,omitempty"`
	// registry overrides the source registry
	Registry string `json:"registry,omitempty" yaml:"registry,omitempty"`
}

// FillDefaults fills the default image pull policy.
func (c *ConfigImagePull) FillDefaults() {
	if c.Policy == "" {
		c.Policy = ConfigPullPolicyIfNotPresent
	}
}

// ImageName returns the target Docker image reference.
func (c *ConfigImagePull) ImageName() string {
	return c.imageName
}

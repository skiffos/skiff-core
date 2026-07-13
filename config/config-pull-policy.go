package config

// ConfigPullPolicy describes when an image is pulled.
type ConfigPullPolicy string

const (
	// ConfigPullPolicyAlways pulls the image before falling back to a build.
	ConfigPullPolicyAlways ConfigPullPolicy = "always"
	// ConfigPullPolicyIfNotPresent pulls only when the image is absent.
	ConfigPullPolicyIfNotPresent ConfigPullPolicy = "ifnotpresent"
	// ConfigPullPolicyIfBuildFails pulls only after a failed image build.
	ConfigPullPolicyIfBuildFails ConfigPullPolicy = "ifbuildfails"
)

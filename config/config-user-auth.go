package config

// ConfigUserAuth describes host-user authentication settings.
type ConfigUserAuth struct {
	// copyRootKeys copies root's authorized SSH keys
	CopyRootKeys bool `json:"copyRootKeys,omitempty" yaml:"copyRootKeys,omitempty"`
	// authorized keys contains additional SSH public keys
	SSHKeys []string `json:"sshKeys,omitempty" yaml:"sshKeys,omitempty"`
	// password sets the host-user password; empty generates a random password
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	// allowEmptyPassword permits an explicitly empty password
	AllowEmptyPassword bool `json:"allowEmptyPassword,omitempty" yaml:"allowEmptyPassword,omitempty"`
	// locked locks the host-user password
	Locked bool `json:"locked,omitempty" yaml:"locked,omitempty"`
}

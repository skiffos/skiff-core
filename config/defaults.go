package config

func DefaultConfig() *Config {
	return &Config{
		Users: map[string]*ConfigUser{
			"core": {
				Auth: &ConfigUserAuth{
					Container:    "skiff_core",
					CopyRootKeys: true,
				},
			},
		},
		Images: map[string]*ConfigImage{
			"skiff/core:latest": {
				Build: &ConfigImageBuild{
					Source: "/opt/skiff/coreenv/user",
				},
			},
		},
		Containers: map[string]*ConfigContainer{
			"skiff_core": {
				Image:       "skiff/core:latest",
				Privileged:  true,
				CapAdd:      []string{"ALL"},
				HostIPC:     true,
				HostPID:     true,
				HostUTS:     true,
				HostNetwork: true,
				SecurityOpt: []string{"seccomp=unconfined"},
				Mounts: []string{
					"/lib/modules:/lib/modules",
					"/sys/fs/cgroup:/sys/fs/cgroup:ro",
					"/dev:/dev",
					"/mnt:/mnt",
				},
				TmpFs: map[string]string{
					"/run": "rw,noexec,nosuid,size=65536k",
				},
			},
		},
	}
}

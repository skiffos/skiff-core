module github.com/skiffos/skiff-core

go 1.18

// fix broken references to git.apache.org/thrift
replace (
	git.apache.org/thrift => github.com/apache/thrift v0.17.0 // latest
	go.opencensus.io => go.opencensus.io v0.24.0 // latest
)

require (
	github.com/docker/cli v24.0.4+incompatible
	github.com/docker/docker v24.0.4+incompatible
	github.com/hpcloud/tail v1.0.0
	github.com/mgutz/str v1.2.0
	github.com/moby/sys/signal v0.6.0
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587
	github.com/paralin/scratchbuild v1.1.3
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.0
	github.com/urfave/cli/v2 v2.24.4
	github.com/zloylos/grsync v1.7.0
	golang.org/x/crypto v0.6.0
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/containerd/containerd v1.6.18 // indirect
	github.com/containerd/typeurl v1.0.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/docker-library/go-dockerlibrary v0.0.0-20200821205225-669fbe5c1d52 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20190725054713-01f96b0aa0cd // indirect
	github.com/klauspost/compress v1.15.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moby/buildkit v0.10.0 // indirect
	github.com/moby/patternmatcher v0.5.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/sys/symlink v0.2.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.3-0.20211202183452-c5a74bcca799 // indirect
	github.com/opencontainers/runc v1.1.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/sergi/go-diff v1.0.0 // indirect
	github.com/src-d/gcfg v1.4.0 // indirect
	github.com/stretchr/testify v1.8.1 // indirect
	github.com/xanzy/ssh-agent v0.2.1 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/term v0.5.0 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	pault.ag/go/debian v0.12.0 // indirect
	pault.ag/go/topsort v0.0.0-20160530003732-f98d2ad46e1a // indirect
)

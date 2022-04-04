module github.com/skiffos/skiff-core

go 1.16

// fix broken references to git.apache.org/thrift
replace (
	git.apache.org/thrift => github.com/apache/thrift v0.16.0 // latest
	go.opencensus.io => go.opencensus.io v0.23.0 // latest
)

require (
	github.com/docker/cli v20.10.14+incompatible
	github.com/docker/docker v20.10.14+incompatible
	github.com/hpcloud/tail v1.0.0
	github.com/mgutz/str v1.2.0
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/paralin/scratchbuild v1.0.3
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli v1.22.5
	github.com/zloylos/grsync v1.6.1
	golang.org/x/crypto v0.0.0-20220331220935-ae2d96664a29
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
)

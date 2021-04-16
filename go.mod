module github.com/paralin/skiff-core

go 1.12

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.8.2-0.20210318095723-fdf1618bf743
	github.com/containerd/containerd => github.com/containerd/containerd v1.5.0-rc.1.0.20210415223756-dda530a75065
	github.com/docker/cli => github.com/docker/cli v20.10.3-0.20210413214420-04dad42c3c82+incompatible
	github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20210415221334-8eb947c5b1d7+incompatible // latest
	github.com/paralin/scratchbuild => github.com/paralin/scratchbuild v1.0.2-0.20191213202554-135ca366503a
	github.com/tonistiigi/fsutil => github.com/tonistiigi/fsutil v0.0.0-20191018213012-0f039a052ca1
)

require (
	github.com/docker/cli v20.10.3-0.20210413214420-04dad42c3c82+incompatible
	github.com/docker/docker v20.10.3-0.20210415221334-8eb947c5b1d7+incompatible
	github.com/hpcloud/tail v1.0.0
	github.com/mgutz/str v1.2.0
	github.com/moby/buildkit v0.8.2 // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/paralin/scratchbuild v1.0.2-0.20191213202554-135ca366503a
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/urfave/cli v1.22.2
	github.com/zloylos/grsync v0.0.0-20200204095520-71a00a7141be
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
)

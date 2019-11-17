module github.com/paralin/skiff-core

go 1.12

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.3-0.20191026113918-67a7fdcf741f
	github.com/containerd/containerd => github.com/containerd/containerd v1.3.1-0.20191114164420-d7ec45b172d9
	github.com/docker/cli => github.com/docker/cli v0.0.0-20191113002236-6c12a82f3306
	github.com/docker/docker => github.com/docker/engine v1.4.2-0.20190822180741-9552f2b2fdde // v18.09.9
	github.com/moby/buildkit => github.com/moby/buildkit v0.6.2-0.20191113225518-5c9365b6f4c2
	github.com/paralin/scratchbuild => github.com/paralin/scratchbuild v1.0.1
	github.com/tonistiigi/fsutil => github.com/tonistiigi/fsutil v0.0.0-20191018213012-0f039a052ca1
)

require (
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/docker/cli v0.0.0-20190815010145-aa097cf1aa19
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/hpcloud/tail v1.0.0
	github.com/mgutz/str v1.2.0
	github.com/paralin/scratchbuild v1.0.1
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/cli v1.21.0
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/grpc v1.25.1 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.2.2
)

module github.com/paralin/skiff-core

go 1.12

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.3-0.20190807103436-de736cf91b92
	github.com/docker/cli => github.com/docker/cli v0.0.0-20190124132759-af2647d55b1d
	github.com/docker/docker => github.com/docker/engine v0.0.0-20190206233949-eb137ff1765f
)

require (
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c
	github.com/mgutz/str v1.2.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/paralin/scratchbuild v0.0.0-20190816191839-75fd33db42d6
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/cli v1.21.0
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.2.2
)

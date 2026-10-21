package builder

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/docker/cli/cli/command/image/build"
	"github.com/moby/go-archive"
	"github.com/moby/moby/client"
	"github.com/moby/moby/client/pkg/jsonmessage"
	"github.com/moby/moby/client/pkg/progress"
	"github.com/moby/moby/client/pkg/streamformatter"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

// terminalOutput returns the file descriptor of output and whether it is a
// terminal, for rendering Docker progress.
func terminalOutput(output io.Writer) (uintptr, bool) {
	file, ok := output.(*os.File)
	if !ok {
		return 0, false
	}
	fd := file.Fd()
	return fd, terminal.IsTerminal(int(fd))
}

// dockerBuild sends directory to the Docker daemon as a build context, with
// dockerfileSource as its Dockerfile, and streams the build progress to
// output. dockerfileName is the Dockerfile's path in directory, which
// .dockerignore patterns do not exclude.
func dockerBuild(
	ctx context.Context,
	dockerClient client.APIClient,
	output io.Writer,
	directory string,
	dockerfileName string,
	dockerfileSource string,
	options client.ImageBuildOptions,
) error {
	// Exclude what .dockerignore names, except the Dockerfile, as the Docker
	// CLI does.
	excludes, err := build.ReadDockerignore(directory)
	if err != nil {
		return err
	}
	if err := build.ValidateContextDirectory(directory, excludes); err != nil {
		return errors.Wrap(err, "validate Docker build context")
	}
	excludes = build.TrimBuildFilesFromExcludes(excludes, dockerfileName, false)

	// Archive the context with the Dockerfile source added to it.
	buildContext, err := archive.TarWithOptions(directory, &archive.TarOptions{ExcludePatterns: excludes})
	if err != nil {
		return err
	}
	defer buildContext.Close()
	buildContext, options.Dockerfile, err = build.AddDockerfileToBuildContext(
		io.NopCloser(strings.NewReader(dockerfileSource)),
		buildContext,
	)
	if err != nil {
		return err
	}

	// Send it, and stream the daemon's progress.
	body := progress.NewProgressReader(
		buildContext,
		streamformatter.NewProgressOutput(output),
		0,
		"",
		"Sending build context to Docker daemon",
	)
	response, err := dockerClient.ImageBuild(ctx, body, options)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	outputFD, isTerminal := terminalOutput(output)
	return jsonmessage.DisplayJSONMessagesStream(response.Body, output, outputFD, isTerminal, nil)
}

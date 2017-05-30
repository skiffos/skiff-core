package library

import (
	"fmt"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker-library/go-dockerlibrary/manifest"
	"github.com/docker/distribution/reference"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// LibraryRepo is the Git URL to the official images repository.
var LibraryRepo = "https://github.com/docker-library/official-images.git"

// LibraryResolver attempts to resolve the source of a library Dockerfile.
type LibraryResolver struct {
	libraryPath   string
	repositoryDir string
}

// GetLibrarySource clones the Dockerfile source for a library image.
func (r *LibraryResolver) GetLibrarySource(ref reference.NamedTagged) (string, error) {
	le := log.WithField("ref", ref.String())
	refName := ref.Name()
	if strings.HasPrefix(refName, "library/") {
		refName = refName[len("library/"):]
	}
	le.Debug("Consulting manifest")
	_, _, man, err := manifest.Fetch(r.libraryPath, refName)
	if err != nil {
		return "", err
	}
	tag := man.GetTag(ref.Tag())
	if tag == nil {
		return "", fmt.Errorf("Cannot find tag %s in library repo %s", ref.Tag(), refName)
	}
	le.Debugf("Got tag:\n%s", tag.String())
	repoPath := path.Join(r.repositoryDir, "library-"+refName)
	gle := log.WithField("repo", tag.GitRepo)
	var repo *git.Repository
	if _, err := os.Stat(repoPath); err == nil {
		repo, _ = git.PlainOpen(repoPath)
	}
	if repo == nil {
		os.RemoveAll(repoPath)
		gle.WithField("path", repoPath).Debug("Cloning")
		repo, err = git.PlainClone(repoPath, false, &git.CloneOptions{
			URL:               tag.GitRepo,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			Progress:          os.Stdout,
		})
		if err != nil {
			return "", err
		}
	}
	wt, err := repo.Worktree()
	if err != nil {
		return "", err
	}
	if tag.GitCommit != "" {
		le.WithField("ref", tag.GitCommit).Debug("Checking-out")
		err = wt.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(tag.GitCommit),
		})
		if err != nil {
			status := "unknown"
			stat, err := wt.Status()
			if err != nil {
				status = stat.String()
			}
			le.WithError(err).Warnf("Unable to checkout commit, using:\n%s", status)
		}
	}
	dockerfileDir := path.Join(repoPath, tag.Directory)
	return dockerfileDir, nil
}

// NewLibraryResolver builds a library resolver.
func NewLibraryResolver(libraryPath string, repositoryDir string) *LibraryResolver {
	return &LibraryResolver{libraryPath: libraryPath, repositoryDir: repositoryDir}
}

// BuildLibraryResolver sets up a library resolver given a temporary cache directory.
func BuildLibraryResolver(cacheDir string) (*LibraryResolver, error) {
	repoDir := path.Join(cacheDir, "official-images")
	log.WithField("url", LibraryRepo).Debug("Cloning")
	var repo *git.Repository
	if st, err := os.Stat(repoDir); err == nil && st.IsDir() {
		r, err := git.PlainOpen(repoDir)
		if err == nil {
			repo = r
			r.Fetch(&git.FetchOptions{
				Progress:   os.Stdout,
				RemoteName: "origin",
			})
			wt, err := r.Worktree()
			if err == nil {
				wt.Checkout(&git.CheckoutOptions{
					Branch: plumbing.ReferenceName("master"),
					Force:  true,
				})
			}
		}
	} else {
		os.RemoveAll(repoDir)
	}
	if repo == nil {
		_, err := git.PlainClone(repoDir, false, &git.CloneOptions{
			Depth:    1,
			Progress: os.Stdout,
			URL:      LibraryRepo,
		})
		if err != nil {
			return nil, err
		}
		// repo = r
	}
	return NewLibraryResolver(path.Join(repoDir, "library"), cacheDir), nil
}

package builder

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// fetchSourceTarball tries to download (if a URL) and extract a tarball/zip archive.
func (b *Builder) fetchSourceTarball(destination, source string) error {
	le := log.WithField("source", "tarball")

	var tarGzReader io.Reader
	if strings.HasPrefix(source, "http") {
		le.WithField("url", source).Debug("Fetching & extracting")
		response, err := http.Get(source)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		tarGzReader = response.Body
	} else {
		le.WithField("path", source).Debug("Extracting")
		f, err := os.Open(source)
		if err != nil {
			return err
		}
		defer f.Close()
		tarGzReader = f
	}

	gzr, err := gzip.NewReader(tarGzReader)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tarr := tar.NewReader(gzr)
	for {
		hdr, err := tarr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// NOTE: avoid zip slip vulnerability
		name := hdr.Name
		if strings.Contains(name, "..") {
			return errors.Errorf("zip entry cannot contain ..: %s", name)
		}

		fdest := path.Join(destination, hdr.Name)
		info := hdr.FileInfo()
		if info.IsDir() {
			if err := os.MkdirAll(fdest, info.Mode()); err != nil {
				return err
			}
			continue
		}

		f, err := os.OpenFile(fdest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		_, err = io.Copy(f, tarr)
		f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

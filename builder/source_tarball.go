package builder

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type pendingArchiveLink struct {
	path       string
	target     string
	isSymlink  bool
	linkSource string
}

func (b *Builder) fetchSourceTarball(ctx context.Context, destination string, source string) error {
	var archiveReader io.Reader
	var sourceCloser io.Closer
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		b.le.WithField("url", source).Debug("fetch image source archive")
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
		if err != nil {
			return err
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			response.Body.Close()
			return errors.Errorf("fetch source archive: HTTP status %s", response.Status)
		}
		archiveReader = response.Body
		sourceCloser = response.Body
	} else {
		b.le.WithField("path", source).Debug("extract image source archive")
		file, err := os.Open(source)
		if err != nil {
			return err
		}
		archiveReader = file
		sourceCloser = file
	}
	defer sourceCloser.Close()

	gzipReader, err := gzip.NewReader(archiveReader)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	var pendingLinks []pendingArchiveLink
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		name := filepath.Clean(filepath.FromSlash(header.Name))
		if !filepath.IsLocal(name) {
			return errors.Errorf("archive entry escapes destination: %s", header.Name)
		}
		destinationPath := filepath.Join(destination, name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(destinationPath, header.FileInfo().Mode().Perm()); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
				return err
			}
			file, err := os.OpenFile(
				destinationPath,
				os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
				header.FileInfo().Mode().Perm(),
			)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(file, tarReader)
			closeErr := file.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		case tar.TypeSymlink:
			targetPath := filepath.Clean(filepath.Join(filepath.Dir(name), filepath.FromSlash(header.Linkname)))
			if !filepath.IsLocal(targetPath) {
				return errors.Errorf("archive symlink escapes destination: %s -> %s", header.Name, header.Linkname)
			}
			pendingLinks = append(pendingLinks, pendingArchiveLink{
				path:       destinationPath,
				target:     header.Linkname,
				isSymlink:  true,
				linkSource: header.Name,
			})
		case tar.TypeLink:
			targetName := filepath.Clean(filepath.FromSlash(header.Linkname))
			if !filepath.IsLocal(targetName) {
				return errors.Errorf("archive hard link escapes destination: %s -> %s", header.Name, header.Linkname)
			}
			pendingLinks = append(pendingLinks, pendingArchiveLink{
				path:       destinationPath,
				target:     filepath.Join(destination, targetName),
				linkSource: header.Name,
			})
		default:
			return errors.Errorf("unsupported archive entry type %d: %s", header.Typeflag, header.Name)
		}
	}

	for _, link := range pendingLinks {
		if err := os.MkdirAll(filepath.Dir(link.path), 0o755); err != nil {
			return err
		}
		if link.isSymlink {
			if err := os.Symlink(link.target, link.path); err != nil {
				return errors.Wrapf(err, "create archive symlink %s", link.linkSource)
			}
			continue
		}
		if err := os.Link(link.target, link.path); err != nil {
			return errors.Wrapf(err, "create archive hard link %s", link.linkSource)
		}
	}
	return nil
}

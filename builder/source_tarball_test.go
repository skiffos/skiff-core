package builder

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

type testArchiveEntry struct {
	header tar.Header
	body   string
}

func writeTestArchive(t *testing.T, entries []testArchiveEntry) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "source-*.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)
	for _, entry := range entries {
		header := entry.header
		header.Size = int64(len(entry.body))
		if err := tarWriter.WriteHeader(&header); err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(tarWriter, entry.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return file.Name()
}

func TestFetchSourceTarballRejectsEscapingPaths(t *testing.T) {
	testCases := []struct {
		name   string
		header tar.Header
	}{
		{
			name: "entry path",
			header: tar.Header{
				Name:     "../outside",
				Typeflag: tar.TypeReg,
				Mode:     0o600,
			},
		},
		{
			name: "symbolic link target",
			header: tar.Header{
				Name:     "link",
				Typeflag: tar.TypeSymlink,
				Linkname: "../outside",
			},
		},
		{
			name: "hard link target",
			header: tar.Header{
				Name:     "link",
				Typeflag: tar.TypeLink,
				Linkname: "../outside",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			archivePath := writeTestArchive(t, []testArchiveEntry{{header: testCase.header}})
			imageBuilder := &Builder{le: logrus.NewEntry(logrus.New())}
			err := imageBuilder.fetchSourceTarball(t.Context(), t.TempDir(), archivePath)
			if err == nil || !strings.Contains(err.Error(), "escapes destination") {
				t.Fatalf("expected destination escape error, got %v", err)
			}
		})
	}
}

func TestFetchSourceTarballCreatesLocalLinks(t *testing.T) {
	archivePath := writeTestArchive(t, []testArchiveEntry{
		{
			header: tar.Header{Name: "target", Typeflag: tar.TypeReg, Mode: 0o600},
			body:   "payload",
		},
		{
			header: tar.Header{Name: "sub/link", Typeflag: tar.TypeSymlink, Linkname: "../target"},
		},
		{
			header: tar.Header{Name: "hard", Typeflag: tar.TypeLink, Linkname: "target"},
		},
	})
	destination := t.TempDir()
	imageBuilder := &Builder{le: logrus.NewEntry(logrus.New())}
	if err := imageBuilder.fetchSourceTarball(t.Context(), destination, archivePath); err != nil {
		t.Fatal(err)
	}

	linkData, err := os.ReadFile(filepath.Join(destination, "sub", "link"))
	if err != nil {
		t.Fatal(err)
	}
	if string(linkData) != "payload" {
		t.Fatalf("expected symbolic link payload, got %q", linkData)
	}
	targetInfo, err := os.Stat(filepath.Join(destination, "target"))
	if err != nil {
		t.Fatal(err)
	}
	hardLinkInfo, err := os.Stat(filepath.Join(destination, "hard"))
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(targetInfo, hardLinkInfo) {
		t.Fatal("expected hard link to share the target inode")
	}
}

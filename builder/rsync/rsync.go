package rsync

import (
	"io"
	"io/ioutil"
	"os"
	"path"
)

// Copies file source to destination dest.
func CopyFile(source string, dest string) error {
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	if err != nil {
		return err
	}
	si, err := os.Stat(source)
	if err != nil {
		return err
	}
	err = os.Chmod(dest, si.Mode())
	if err != nil {
		return err
	}
	return err
}

// Recursively copies a directory tree, attempting to preserve permissions.
func CopyDir(source string, dest string) error {
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(source)
	for _, entry := range entries {
		sfp := path.Join(source, entry.Name())
		dfp := path.Join(dest, entry.Name())
		if entry.IsDir() {
			err = CopyDir(sfp, dfp)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(sfp, dfp)
			if err != nil {
				return err
			}
		}

	}
	return err
}

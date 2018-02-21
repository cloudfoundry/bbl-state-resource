package storage

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bbl-state-resource/concourse"
)

var ObjectNotFoundError = errors.New("Object not found")

type object interface {
	NewReader() (io.ReadCloser, error)
	NewWriter() io.WriteCloser
	Version() (string, error)
}

type tarrer interface {
	Write(io.Writer, []string) error
	Read(io.Reader, string) error
}

type Storage struct {
	DirectoryName string
	Object        object
	Archiver      tarrer
}

func (s Storage) Version() (concourse.Version, error) {
	version, err := s.Object.Version()
	if err != nil {
		return concourse.Version{}, err
	}
	return concourse.Version{Ref: version}, nil
}

func (s Storage) Download(targetDir string) error {
	reader, err := s.Object.NewReader()
	if err != nil {
		if err == ObjectNotFoundError {
			return s.Upload(targetDir)
		}
		return err
	}
	defer reader.Close() // what happens if this errors?

	tmpDir, err := ioutil.TempDir("", "")
	// defer os.Remove(tmpDir)

	err = s.Archiver.Read(reader, tmpDir)
	if err != nil {
		return err
	}

	err = os.Rename(filepath.Join(tmpDir, s.DirectoryName), targetDir)
	if err != nil {
		return err
	}

	return s.Upload(targetDir)
}

func (s Storage) Upload(filePath string) error {
	writer := s.Object.NewWriter()

	tempDir := filepath.Join(os.TempDir(), s.DirectoryName)
	defer os.RemoveAll(tempDir)

	err := CopyDir(filePath, tempDir)
	if err != nil {
		return err
	}

	err = s.Archiver.Write(writer, []string{tempDir})
	if err != nil {
		return err
	}

	return writer.Close()
}

func CopyDir(sourcePath, destPath string) error {
	return filepath.Walk(sourcePath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if sourcePath == path {
			return os.MkdirAll(destPath, f.Mode())
		}

		if f.IsDir() {
			return os.MkdirAll(filepath.Join(destPath, f.Name()), f.Mode())
		}

		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		subpath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(filepath.Join(destPath, subpath), contents, f.Mode())
	})
}

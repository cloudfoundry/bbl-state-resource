package storage

import (
	"errors"
	"io"
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
	Object   object
	Archiver tarrer
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

	err = os.MkdirAll(targetDir, 777)
	if err != nil {
		return err
	}

	err = s.Archiver.Read(reader, targetDir)
	if err != nil {
		return err
	}

	return s.Upload(targetDir)
}

func (s Storage) Upload(filePath string) error {
	writer := s.Object.NewWriter()
	paths := []string{}
	err := filepath.Walk(filePath, func(path string, f os.FileInfo, err error) error {
		if filePath == path {
			return nil
		}
		paths = append(paths, path)
		return err
	})
	if err != nil {
		return err
	}

	err = s.Archiver.Write(writer, paths)
	if err != nil {
		return err
	}

	return writer.Close()
}

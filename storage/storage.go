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

func (s Storage) Download(targetDir string) (concourse.Version, error) {
	reader, err := s.Object.NewReader()
	if err != nil {
		if err == ObjectNotFoundError {
			return s.Upload(targetDir)
		}
		return concourse.Version{}, err
	}
	defer reader.Close() // what happens if this errors?

	err = os.MkdirAll(targetDir, os.ModePerm)
	if err != nil {
		return concourse.Version{}, err
	}

	err = s.Archiver.Read(reader, targetDir)
	if err != nil {
		return concourse.Version{}, err
	}

	return s.Version()
}

func (s Storage) Upload(filePath string) (concourse.Version, error) {
	writer := s.Object.NewWriter()
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		return concourse.Version{}, err
	}

	paths := []string{}
	for _, file := range files {
		paths = append(paths, filepath.Join(filePath, file.Name()))
	}

	err = s.Archiver.Write(writer, paths)
	if err != nil {
		return concourse.Version{}, err
	}

	err = writer.Close()
	if err != nil {
		return concourse.Version{}, err
	}

	return s.Version()
}

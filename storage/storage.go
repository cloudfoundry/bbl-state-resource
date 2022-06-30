package storage

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/mholt/archiver/v4"
)

var ObjectNotFoundError = errors.New("Object not found")

type Version struct {
	Name    string    `json:"name"`
	Ref     string    `json:"ref"`
	Updated time.Time `json:"updated"`
}

// public only because []Object != []ObjectImpl :(
type Object interface {
	NewReader() (io.ReadCloser, error)
	NewWriter() io.WriteCloser
	Version() (Version, error)
}

type Bucket interface {
	GetAllObjects() ([]Object, error)
	Delete() error // test only
}

type tarrer interface {
	Archive(ctx context.Context, output io.Writer, files []archiver.File) error
	Extract(ctx context.Context, sourceArchive io.Reader, pathsInArchive []string, handleFile archiver.FileHandler) error
}

type Storage struct {
	Name     string
	Bucket   Bucket
	Object   Object
	Archiver tarrer
}

func (s Storage) GetAllNewerVersions(watermark Version) ([]Version, error) {
	objects, err := s.Bucket.GetAllObjects()
	if err != nil {
		return nil, err
	}
	versions := []Version{}
	for _, object := range objects {
		version, err := object.Version()
		if err != nil {
			return nil, err
		}
		if version.Updated.Before(watermark.Updated) {
			continue
		}
		versions = append(versions, version)
	}
	return versions, nil
}

func (s Storage) Version() (Version, error) {
	return s.Object.Version()
}

func (s Storage) Download(targetDir string) (Version, error) {
	reader, err := s.Object.NewReader()
	if err != nil {
		if err == ObjectNotFoundError {
			return s.Upload(targetDir)
		}
		return Version{}, err
	}
	defer reader.Close() // what happens if this errors?

	err = os.MkdirAll(targetDir, os.ModePerm)
	if err != nil {
		return Version{}, err
	}

	handler := func(ctx context.Context, f archiver.File) error {
		hdr, ok := f.Header.(*tar.Header)

		if !ok {
			return nil
		}

		var fpath = filepath.Join(targetDir, f.NameInArchive)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(fpath, 0755); err != nil {
				return fmt.Errorf("failed to make directory %s: %w", fpath, err)
			}
			return nil

		case tar.TypeReg, tar.TypeRegA, tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
			if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
				return fmt.Errorf("failed to make directory %s: %w", filepath.Dir(fpath), err)
			}

			out, err := os.Create(fpath)
			if err != nil {
				return fmt.Errorf("%s: creating new file: %v", fpath, err)
			}
			defer out.Close()

			err = out.Chmod(f.Mode())
			if err != nil && runtime.GOOS != "windows" {
				return fmt.Errorf("%s: changing file mode: %v", fpath, err)
			}

			in, err := f.Open()
			if err != nil {
				return err
			}

			_, err = io.Copy(out, in)
			if err != nil {
				return fmt.Errorf("%s: writing file: %v", fpath, err)
			}
			return nil

		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
				return fmt.Errorf("failed to make directory %s: %w", filepath.Dir(fpath), err)
			}

			err = os.Symlink(hdr.Linkname, fpath)
			if err != nil {
				return fmt.Errorf("%s: making symbolic link for: %v", fpath, err)
			}
			return nil

		case tar.TypeLink:
			if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
				return fmt.Errorf("failed to make directory %s: %w", filepath.Dir(fpath), err)
			}

			err = os.Link(filepath.Join(fpath, hdr.Linkname), fpath)
			if err != nil {
				return fmt.Errorf("%s: making symbolic link for: %v", fpath, err)
			}
			return nil

		case tar.TypeXGlobalHeader:
			return nil // ignore the pax global header from git-generated tarballs
		default:
			return fmt.Errorf("%s: unknown type flag: %c", hdr.Name, hdr.Typeflag)
		}
	}

	err = s.Archiver.Extract(context.Background(), reader, nil, handler)
	if err != nil {
		return Version{}, err
	}

	return s.Version()
}

func (s Storage) Upload(filePath string) (Version, error) {
	writer := s.Object.NewWriter()
	path := make(map[string]string)
	path[filePath+"/"] = ""
	diskfiles, err := archiver.FilesFromDisk(nil, path)
	if err != nil {
		return Version{}, err
	}

	err = s.Archiver.Archive(context.Background(), writer, diskfiles)
	if err != nil {
		return Version{}, err
	}

	err = writer.Close()
	if err != nil {
		return Version{}, err
	}

	return s.Version()
}

// test cleanup only
func (s Storage) DeleteBucket() error {
	return s.Bucket.Delete()
}

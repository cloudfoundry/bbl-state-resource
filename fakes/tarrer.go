package fakes

import (
	"context"
	"io"

	"github.com/mholt/archiver/v4"
)

type Tarrer struct {
	ArchiveCall struct {
		CallCount int
		Receives  struct {
			Output io.Writer
			Files  []archiver.File
		}
		Returns struct {
			Error error
		}
	}

	ExtractCall struct {
		CallCount int
		Receives  struct {
			SourceArchive  io.Reader
			PathsInArchive []string
			HandleFile     archiver.FileHandler
		}
		Returns struct {
			Error error
		}
	}
}

func (t *Tarrer) Archive(ctx context.Context, output io.Writer, files []archiver.File) error {
	t.ArchiveCall.CallCount++
	t.ArchiveCall.Receives.Output = output
	t.ArchiveCall.Receives.Files = files
	return t.ArchiveCall.Returns.Error
}

func (t *Tarrer) Extract(ctx context.Context, sourceArchive io.Reader, pathsInArchive []string, handleFile archiver.FileHandler) error {
	t.ExtractCall.CallCount++
	t.ExtractCall.Receives.SourceArchive = sourceArchive
	t.ExtractCall.Receives.PathsInArchive = pathsInArchive
	t.ExtractCall.Receives.HandleFile = handleFile
	return t.ExtractCall.Returns.Error
}

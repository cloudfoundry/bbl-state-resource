package fakes

import (
	"io"

	"github.com/cloudfoundry/bbl-state-resource/storage"
)

type Object struct {
	VersionCall struct {
		Returns struct {
			Version storage.Version
			Error   error
		}
	}

	NewReaderCall struct {
		CallCount int
		Returns   struct {
			ReadCloser io.ReadCloser
			Error      error
		}
	}

	NewWriterCall struct {
		CallCount int
		Returns   struct {
			WriteCloser io.WriteCloser
		}
	}
}

func (g *Object) Version() (storage.Version, error) {
	return g.VersionCall.Returns.Version, g.VersionCall.Returns.Error
}

func (g *Object) NewReader() (io.ReadCloser, error) {
	g.NewReaderCall.CallCount++
	return g.NewReaderCall.Returns.ReadCloser, g.NewReaderCall.Returns.Error
}

func (g *Object) NewWriter() io.WriteCloser {
	g.NewWriterCall.CallCount++
	return g.NewWriterCall.Returns.WriteCloser
}

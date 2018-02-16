package fakes

import "io"

type GCSBucketObject struct {
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

func (g *GCSBucketObject) NewReader() (io.ReadCloser, error) {
	g.NewReaderCall.CallCount++
	return g.NewReaderCall.Returns.ReadCloser, g.NewReaderCall.Returns.Error
}

func (g *GCSBucketObject) NewWriter() io.WriteCloser {
	g.NewWriterCall.CallCount++
	return g.NewWriterCall.Returns.WriteCloser
}

package fakes

type WriteCloser struct {
	WriteCall struct {
		Receives struct {
			Contents []byte
		}
		Returns struct {
			BytesWritten int
			Error        error
		}
	}

	CloseCall struct {
		CallCount int
		Returns   struct {
			Error error
		}
	}
}

func (r *WriteCloser) Write(p []byte) (n int, err error) {
	r.WriteCall.Receives.Contents = p
	return r.WriteCall.Returns.BytesWritten, r.WriteCall.Returns.Error
}

func (r *WriteCloser) Close() error {
	r.CloseCall.CallCount++
	return r.CloseCall.Returns.Error
}

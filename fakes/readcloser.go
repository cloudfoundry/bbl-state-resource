package fakes

type ReadCloser struct {
	ReadCall struct {
		Receives struct {
			Buffer []byte
		}
		Returns struct {
			BytesRead int
			Error     error
		}
	}

	CloseCall struct {
		CallCount int
		Returns   struct {
			Error error
		}
	}
}

func (r *ReadCloser) Read(p []byte) (n int, err error) {
	r.ReadCall.Receives.Buffer = p
	return r.ReadCall.Returns.BytesRead, r.ReadCall.Returns.Error
}

func (r *ReadCloser) Close() error {
	r.CloseCall.CallCount++
	return r.CloseCall.Returns.Error
}

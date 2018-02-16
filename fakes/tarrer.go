package fakes

import "io"

type Tarrer struct {
	ReadCall struct {
		CallCount int
		Receives  struct {
			Input       io.Reader
			Destination string
		}
		Returns struct {
			Error error
		}
	}

	WriteCall struct {
		CallCount int
		Receives  struct {
			Output  io.Writer
			Sources []string
		}
		Returns struct {
			Error error
		}
	}
}

func (t *Tarrer) Write(output io.Writer, sources []string) error {
	t.WriteCall.CallCount++
	t.WriteCall.Receives.Output = output
	t.WriteCall.Receives.Sources = sources
	return t.WriteCall.Returns.Error
}

func (t *Tarrer) Read(input io.Reader, destination string) error {
	t.ReadCall.CallCount++
	t.ReadCall.Receives.Input = input
	t.ReadCall.Receives.Destination = destination
	return t.ReadCall.Returns.Error
}

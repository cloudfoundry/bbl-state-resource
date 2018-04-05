package fakes

import "github.com/cloudfoundry/bbl-state-resource/outrunner"

type StateDir struct {
	ReadCall struct {
		CallCount int
		Returns   struct {
			BblState outrunner.BblState
			Error    error
		}
	}

	JumpboxSSHKeyCall struct {
		CallCount int
		Returns   struct {
			Key   string
			Error error
		}
	}

	PathCall struct {
		CallCount int
		Returns   struct {
			Path string
		}
	}

	WriteMetadataCall struct {
		CallCount int
		Receives  struct {
			Metadata string
		}
		Returns struct {
			Error error
		}
	}

	WriteNameCall struct {
		CallCount int
		Receives  struct {
			Name string
		}
		Returns struct {
			Error error
		}
	}
}

func (s *StateDir) Read() (outrunner.BblState, error) {
	s.ReadCall.CallCount++

	return s.ReadCall.Returns.BblState, s.ReadCall.Returns.Error
}

func (s *StateDir) JumpboxSSHKey() (string, error) {
	s.JumpboxSSHKeyCall.CallCount++

	return s.JumpboxSSHKeyCall.Returns.Key, s.JumpboxSSHKeyCall.Returns.Error
}

func (s *StateDir) Path() string {
	s.PathCall.CallCount++

	return s.PathCall.Returns.Path
}

func (s *StateDir) WriteMetadata(metadata string) error {
	s.WriteMetadataCall.CallCount++
	s.WriteMetadataCall.Receives.Metadata = metadata

	return s.WriteMetadataCall.Returns.Error
}

func (s *StateDir) WriteName(name string) error {
	s.WriteNameCall.CallCount++
	s.WriteNameCall.Receives.Name = name

	return s.WriteNameCall.Returns.Error
}

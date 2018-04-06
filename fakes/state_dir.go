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

	WriteInteropFilesCall struct {
		CallCount int
		Receives  struct {
			Config outrunner.BoshDeploymentResourceConfig
			Name   string
		}
		Returns struct {
			Error error
		}
	}

	ExpungeInteropFilesCall struct {
		CallCount int
		Returns   struct {
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

func (s *StateDir) ExpungeInteropFiles() error {
	s.ExpungeInteropFilesCall.CallCount++

	return s.ExpungeInteropFilesCall.Returns.Error
}

func (s *StateDir) WriteInteropFiles(name string, c outrunner.BoshDeploymentResourceConfig) error {
	s.WriteInteropFilesCall.CallCount++
	s.WriteInteropFilesCall.Receives.Config = c
	s.WriteInteropFilesCall.Receives.Name = name

	return s.WriteInteropFilesCall.Returns.Error
}

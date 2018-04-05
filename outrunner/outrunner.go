package outrunner

import (
	"fmt"
	"os"
	"os/exec"
)

type stateDir interface {
	Path() string
	Read() (BblState, error)
	JumpboxSSHKey() (string, error)
	WriteName(string) error
	WriteMetadata(string) error
}

func RunBBL(name string, stateDir stateDir, command string, flags map[string]interface{}) error {
	return RunInjected(bblRunner, name, stateDir, command, flags)
}

func RunInjected(r commandRunner, name string, stateDir stateDir, command string, flags map[string]interface{}) error {
	bblState, err := stateDir.Read()
	if err != nil {
		return err
	}

	_ = bblState.Director.ClientUsername
	_ = bblState.Director.ClientSecret
	_ = bblState.Director.Address
	_ = bblState.Director.CaCert

	_, err = stateDir.JumpboxSSHKey()
	if err != nil {
		return err
	}

	err = stateDir.WriteName(name)
	if err != nil {
		return err
	}

	err = stateDir.WriteMetadata(name)
	if err != nil {
		return err
	}

	args := []string{}
	args = append(args, fmt.Sprintf("--name=%s", name))
	args = append(args, fmt.Sprintf("--state-dir=%s", stateDir.Path()))
	for key, value := range flags {
		args = append(args, fmt.Sprintf("--%s=%s", key, value))
	}

	return r.Run(command, args)
}

type commandRunner interface {
	Run(string, []string) error
}

type bblRunnerT struct{}

var bblRunner = bblRunnerT{}

func (bblRunnerT) Run(command string, args []string) error {
	args = append([]string{"-n", command}, args...)
	cmd := exec.Command("bbl", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stderr

	return cmd.Run()
}

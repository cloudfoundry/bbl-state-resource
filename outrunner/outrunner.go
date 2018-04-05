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
	WriteBoshDeploymentResourceConfig(BoshDeploymentResourceConfig) error
}

func RunBBL(name string, stateDir stateDir, command string, flags map[string]interface{}) error {
	return RunInjected(bblRunner, name, stateDir, command, flags)
}

func RunInjected(r commandRunner, name string, stateDir stateDir, command string, flags map[string]interface{}) error {
	err := stateDir.WriteName(name)
	if err != nil {
		return err
	}

	args := []string{}
	args = append(args, fmt.Sprintf("--name=%s", name))
	args = append(args, fmt.Sprintf("--state-dir=%s", stateDir.Path()))
	for key, value := range flags {
		args = append(args, fmt.Sprintf("--%s=%s", key, value))
	}

	err = r.Run(command, args)
	if err != nil {
		return fmt.Errorf("failed running bbl %s --state-dir=%s <sensitive flags omitted>: %s", command, stateDir.Path(), err)
	}

	bblState, err := stateDir.Read()
	if err != nil {
		return fmt.Errorf("failed reading bbl state: %s", err)
	}

	sshKey, err := stateDir.JumpboxSSHKey()
	if err != nil {
		return fmt.Errorf("failed reading jumpbox ssh key: %s", err)
	}

	return stateDir.WriteBoshDeploymentResourceConfig(BoshDeploymentResourceConfig{
		Target:          bblState.Director.Address,
		Client:          bblState.Director.ClientUsername,
		ClientSecret:    bblState.Director.ClientSecret,
		CaCert:          bblState.Director.CaCert,
		JumpboxUrl:      bblState.Jumpbox.URL,
		JumpboxSSHKey:   sshKey,
		JumpboxUsername: "jumpbox",
	})
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

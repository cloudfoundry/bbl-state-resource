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
	WriteInteropFiles(name string, config BoshDeploymentResourceConfig) error
	ExpungeInteropFiles() error
}

func RunBBL(name string, stateDir stateDir, command string, flags map[string]interface{}) error {
	return RunInjected(bblRunner, name, stateDir, command, flags)
}

func RunInjected(r commandRunner, name string, stateDir stateDir, command string, flags map[string]interface{}) error {
	args := []string{}
	args = append(args, fmt.Sprintf("--name=%s", name))
	args = append(args, fmt.Sprintf("--state-dir=%s", stateDir.Path()))
	for key, value := range flags {
		args = append(args, fmt.Sprintf("--%s=%s", key, value))
	}

	err := r.Run(command, args)
	SyncInteropFiles(stateDir)
	if err != nil {
		return fmt.Errorf("failed running bbl %s --state-dir=%s <sensitive flags omitted>: %s", command, stateDir.Path(), err)
	}
	return nil
}

// best effort, log aggressively
func SyncInteropFiles(stateDir stateDir) {
	bblState, err := stateDir.Read()
	if os.IsNotExist(err) {
		err = stateDir.ExpungeInteropFiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to expunge interop files: %s\n", err)
		}
		return // quietly return if we've successfully expunged after bbl down
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed reading bbl-state: %s\n", err)
	}

	sshKey, err := stateDir.JumpboxSSHKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed reading jumpbox ssh key: %s\n", err)
	}

	err = stateDir.WriteInteropFiles(bblState.EnvID, BoshDeploymentResourceConfig{
		Target:          bblState.Director.Address,
		Client:          bblState.Director.ClientUsername,
		ClientSecret:    bblState.Director.ClientSecret,
		CaCert:          bblState.Director.CaCert,
		JumpboxUrl:      bblState.Jumpbox.URL,
		JumpboxSSHKey:   sshKey,
		JumpboxUsername: "jumpbox",
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write interop files: %s\n", err)
	}
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

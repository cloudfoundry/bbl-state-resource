package outrunner

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

func RunBBL(name, stateDir, command string, flags map[string]interface{}) error {
	return RunInjected(bblRunner, name, stateDir, command, flags)
}

func RunInjected(r commandRunner,
	name, stateDir, command string, flags map[string]interface{},
) error {
	args := []string{}
	addArg := func(key string, value interface{}) {
		args = append(args, fmt.Sprintf("--%s=%s", key, value))
	}

	addArg("name", name)
	addArg("state-dir", stateDir)

	for key, value := range flags {
		addArg(key, value)
	}

	err := ioutil.WriteFile(
		filepath.Join(stateDir, "name"),
		[]byte(name),
		os.ModePerm,
	)
	if err != nil {
		return err
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

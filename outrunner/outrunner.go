package outrunner

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cloudfoundry/bbl-state-resource/concourse"
	"github.com/fatih/structs"
)

func RunBBL(outRequest concourse.OutRequest, bblStateDir string) error {
	return RunInjected(bblRunner, outRequest, bblStateDir)
}

func RunInjected(r commandRunner, outRequest concourse.OutRequest, bblStateDir string) error {
	args := []string{}
	addArg := func(key string, value interface{}) {
		args = append(args, fmt.Sprintf("--%s=%s", key, value))
	}

	for key, value := range structs.Map(outRequest.Source) {
		outRequest.Params.Args[key] = value
	}

	for key, value := range outRequest.Params.Args {
		addArg(key, value)
	}

	addArg("state-dir", bblStateDir)

	return r.Run(outRequest.Params.Command, args)
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

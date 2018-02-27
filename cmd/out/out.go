package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bbl-state-resource/concourse"
	"github.com/cloudfoundry/bbl-state-resource/outrunner"
	"github.com/cloudfoundry/bbl-state-resource/storage"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr,
			"not enough args - usage: %s <target directory>\n",
			os.Args[0],
		)
		os.Exit(1)
	}
	stateDir := os.Args[1]

	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read configuration: %s\n", err)
		os.Exit(1)
	}

	outRequest, err := concourse.NewOutRequest(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid parameters: %s\n", err)
		os.Exit(1)
	}

	storageClient, err := storage.NewStorageClient(outRequest.Source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create storage client: %s", err.Error())
	}

	_, err = storageClient.Download(stateDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to download bbl state: %s", err.Error())
		os.Exit(1)
	}

	var bblError error
	defer func() { // upload state even on failure for introspectability
		version, err := storageClient.Upload(stateDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to upload bbl state: %s", err.Error())
			os.Exit(1)
		}

		outMap := map[string]concourse.Version{"version": version}
		err = json.NewEncoder(os.Stdout).Encode(outMap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal version: %s", err.Error())
			os.Exit(1)
		}
		if bblError != nil {
			os.Exit(1)
		}
	}()

	bblError = outrunner.RunBBL(outRequest, stateDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run bbl command: %s", err.Error())
	}
}

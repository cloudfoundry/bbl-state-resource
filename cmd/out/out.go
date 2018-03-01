package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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
	sourcesDir := os.Args[1]

	fmt.Fprintf(os.Stderr, "sourcesDir: %s\n", sourcesDir)

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
		fmt.Fprintf(os.Stderr, "failed to create storage client: %s\n", err)
		os.Exit(1)
	}

	stateDir := outRequest.Params.StateDir
	if stateDir == "" {
		stateDir = filepath.Join(sourcesDir, "bbl-state")
		err = os.Mkdir(stateDir, os.ModePerm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create %s directory: %s\n", stateDir, err)
			os.Exit(1)
		}

		_, err = storageClient.Download(stateDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to download bbl state: %s\n", err)
			os.Exit(1)
		}
	}

	fmt.Fprintf(os.Stderr, "running 'bbl %s --state-dir=%s'...\n", outRequest.Params.Command, sourcesDir)

	bblError := outrunner.RunBBL(outRequest, stateDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run bbl command: %s\n", err)
	}

	version, err := storageClient.Upload(stateDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to upload bbl state: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "successfully uploaded bbl state!\n")

	outMap := map[string]concourse.Version{"version": version}
	err = json.NewEncoder(os.Stdout).Encode(outMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal version: %s\n", err)
		os.Exit(1)
	}
	if bblError != nil {
		os.Exit(1)
	}
}

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
			"not enough args - usage: %s <sources directory>\n",
			os.Args[0],
		)
		os.Exit(1)
	}
	sourcesDir := os.Args[1]

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

	name, err := outrunner.Name(sourcesDir, outRequest.Params)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}

	storageClient, err := storage.NewStorageClient(
		outRequest.Source.GCPServiceAccountKey,
		name,
		outRequest.Source.Bucket,
	)
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

	fmt.Fprintf(os.Stderr, "running something like 'bbl %s --state-dir=%s'...\n", outRequest.Params.Command, stateDir)

	bblError := outrunner.RunBBL(name, stateDir, outRequest.Params.Command,
		outrunner.AppendSourceFlags(outRequest.Params.Args, outRequest.Source))
	if bblError != nil {
		fmt.Fprintf(os.Stderr, "failed to run bbl command: %s\n", err)
	}

	version, err := storageClient.Upload(stateDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to upload bbl state: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "successfully uploaded bbl state!\n")

	outMap := map[string]storage.Version{"version": version}
	err = json.NewEncoder(os.Stdout).Encode(outMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal version: %s\n", err)
		os.Exit(1)
	}
	if bblError != nil {
		os.Exit(1)
	}
}

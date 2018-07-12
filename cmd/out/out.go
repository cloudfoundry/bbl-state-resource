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

	req, err := concourse.NewOutRequest(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid parameters: %s\n", err)
		os.Exit(1)
	}

	name, err := outrunner.Name(sourcesDir, req.Params)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}

	storageClient, err := storage.NewStorageClient(req.Source.GCPServiceAccountKey, name, req.Source.Bucket)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create storage client: %s\n", err)
		os.Exit(1)
	}

	bblStateDir := filepath.Join(sourcesDir, req.Params.StateDir)
	if req.Params.StateDir == "" {
		bblStateDir = filepath.Join(sourcesDir, "bbl-state")
		err = os.Mkdir(bblStateDir, os.ModePerm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create %s directory: %s\n", bblStateDir, err)
			os.Exit(1)
		}

		_, err = storageClient.Download(bblStateDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to download bbl state: %s\n", err)
			os.Exit(1)
		}
	}

	fmt.Fprintf(os.Stderr, "running something like 'bbl %s --state-dir=%s'...\n", req.Params.Command, bblStateDir)

	stateDir := outrunner.NewStateDir(bblStateDir)

	err = stateDir.ApplyPlanPatches(req.Params.PlanPatches)

	bblError := outrunner.RunBBL(name, stateDir, req.Params.Command, outrunner.AppendSourceFlags(req.Params.Args, req.Source))
	if bblError != nil {
		fmt.Fprintf(os.Stderr, "failed to run bbl command: %s\n", bblError)
	}

	version, err := storageClient.Upload(bblStateDir)
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

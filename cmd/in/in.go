package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bbl-state-resource/concourse"
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

	rawBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read configuration: %s\n", err)
		os.Exit(1)
	}

	req, err := concourse.NewInRequest(rawBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid parameters: %s\n", err)
		os.Exit(1)
	}

	storageClient, err := storage.NewStorageClient(req.Source.GCPServiceAccountKey, req.Version.Name, req.Source.Bucket)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create storage client: %s\n", err)
		os.Exit(1)
	}

	version, err := storageClient.Download(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to download bbl state: %s\n", err)
		os.Exit(1)
	}

	outMap := map[string]storage.Version{"version": version}
	err = json.NewEncoder(os.Stdout).Encode(outMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal version: %s\n", err)
		os.Exit(1)
	}
}

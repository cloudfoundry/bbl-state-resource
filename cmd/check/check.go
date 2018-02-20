package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/cloudfoundry/bbl-state-resource/concourse"
	"github.com/cloudfoundry/bbl-state-resource/storage"
)

func main() {
	rawBytes, err := ioutil.ReadAll(os.Stdin)
	fmt.Fprintf(os.Stderr, "bytes passed: %s\n", rawBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read configuration: %s\n", err)
		os.Exit(1)
	}

	checkRequest, err := concourse.NewInRequest(rawBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid parameters: %s\n", err)
		os.Exit(1)
	}

	storageClient, err := storage.NewStorageClient(checkRequest.Source)
	if err != nil {
		log.Fatalf("failed to create storage client: %s", err.Error())
	}

	version, err := storageClient.Version()
	if err == storage.ObjectNotFoundError {
		fmt.Fprintf(os.Stdout, `[]`)
		os.Exit(0)
	} else if err != nil {
		log.Fatalf("failed to fetch bbl state version: %s", err.Error())
	}

	err = json.NewEncoder(os.Stdout).Encode([]concourse.Version{version})
	if err != nil {
		log.Fatalf("failed to marshal version: %s", err.Error())
	}
}

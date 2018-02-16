package main

import (
	"fmt"
	"io/ioutil"
	"log"
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

	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read configuration: %s\n", err)
		os.Exit(1)
	}

	inRequest, err := concourse.NewInRequest(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid parameters: %s\n", err)
		os.Exit(1)
	}

	storageClient, err := storage.NewStorageClient(inRequest.Source)
	if err != nil {
		log.Fatalf("failed to create storage client: %s", err.Error())
	}

	err = storageClient.Upload(os.Args[1])
	if err != nil {
		log.Fatalf("failed to upload bbl state: %s", err.Error())
	}
}

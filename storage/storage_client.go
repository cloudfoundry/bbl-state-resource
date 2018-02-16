package storage

import (
	"fmt"

	"github.com/cloudfoundry/bbl-state-resource/concourse"
)

type StorageClient interface {
	Download(filePath string) error
	Upload(filePath string) error
}

func NewStorageClient(source concourse.Source) (StorageClient, error) {
	return NewGCSStorage(source.GCPServiceAccountKey, fmt.Sprintf("bbl-state-for-%s", source.Name))
}

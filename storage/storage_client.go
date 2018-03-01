package storage

import (
	"github.com/cloudfoundry/bbl-state-resource/concourse"
)

type StorageClient interface {
	Download(filePath string) (concourse.Version, error)
	Upload(filePath string) (concourse.Version, error)
	Version() (concourse.Version, error)
}

func NewStorageClient(source concourse.Source) (StorageClient, error) {
	return NewGCSStorage(source.GCPServiceAccountKey, source.Name)
}

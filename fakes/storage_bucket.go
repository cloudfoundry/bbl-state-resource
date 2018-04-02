package fakes

import storage "github.com/cloudfoundry/bbl-state-resource/storage"

type Bucket struct {
	ObjectsCall struct {
		Returns struct {
			Objects []storage.Object
			Error   error
		}
	}
}

func (b *Bucket) GetAllObjects() ([]storage.Object, error) {
	return b.ObjectsCall.Returns.Objects, b.ObjectsCall.Returns.Error
}

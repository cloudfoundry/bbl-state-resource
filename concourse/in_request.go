package concourse

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry/bbl-state-resource/storage"
)

type InRequest struct {
	Source  Source          `json:"source"`
	Version storage.Version `json:"version"`
}

func NewInRequest(request []byte) (InRequest, error) {
	var inRequest InRequest
	if err := json.NewDecoder(bytes.NewReader(request)).Decode(&inRequest); err != nil {
		return InRequest{}, fmt.Errorf("These are invalid parameters: %s\n", err)
	}

	return inRequest, nil
}

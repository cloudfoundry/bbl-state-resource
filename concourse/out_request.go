package concourse

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type OutRequest struct {
	Source Source    `json:"source"`
	Params OutParams `json:"params"`
}

func NewOutRequest(request []byte) (OutRequest, error) {
	var outRequest OutRequest
	if err := json.NewDecoder(bytes.NewReader(request)).Decode(&outRequest); err != nil {
		return OutRequest{}, fmt.Errorf("These are invalid parameters: %s\n", err)
	}

	return outRequest, nil
}

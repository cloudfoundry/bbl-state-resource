package outrunner

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type StateDir struct {
	dir string
}

func NewStateDir(dir string) StateDir {
	return StateDir{
		dir: dir,
	}
}

func (b StateDir) Path() string {
	return b.dir
}

func (b StateDir) Read() (BblState, error) {
	_, err := os.Stat(b.dir)
	if err != nil {
		return BblState{}, err
	}

	stateFile := filepath.Join(b.dir, "bbl-state.json")

	file, err := os.Open(stateFile)
	if err != nil {
		return BblState{}, err
	}

	state := BblState{}

	err = json.NewDecoder(file).Decode(&state)
	if err != nil {
		return BblState{}, err
	}

	return state, nil
}

func (b StateDir) JumpboxSSHKey() (string, error) {
	path := filepath.Join(b.dir, "vars", "jumpbox-vars-store.yml")

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("Read jumpbox vars store: %s", err)
	}

	var p struct {
		JumpboxSSH struct {
			PrivateKey string `yaml:"private_key"`
		} `yaml:"jumpbox_ssh"`
	}

	err = yaml.Unmarshal(contents, &p)
	if err != nil {
		return "", err
	}

	return p.JumpboxSSH.PrivateKey, nil
}

func (b StateDir) WriteMetadata(metadata string) error {
	return ioutil.WriteFile(filepath.Join(b.dir, "metadata"), []byte(metadata), os.ModePerm)
}

func (b StateDir) WriteName(name string) error {
	return ioutil.WriteFile(filepath.Join(b.dir, "name"), []byte(name), os.ModePerm)
}

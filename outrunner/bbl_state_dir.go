package outrunner

import (
	"encoding/json"
	"fmt"
	"io"
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
	defer file.Close()

	state := BblState{}

	err = json.NewDecoder(file).Decode(&state)
	if err != nil {
		return BblState{}, err
	}

	return state, nil
}

func copyDir(source, dest string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dest, info.Mode())
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(source)
	if err != nil {
		return err
	}

	for _, file := range files {
		source := filepath.Join(source, file.Name())
		dest := filepath.Join(dest, file.Name())

		err := copyHelper(source, dest)
		if err != nil {
			return err
		}
	}

	return nil
}

func copyFile(source, dest string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func copyHelper(source, dest string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDir(source, dest)
	} else {
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		return copyFile(source, dest)
	}
}

func (b StateDir) ApplyPlanPatches(patchPaths []string) error {
	for _, patch := range patchPaths {
		err := copyHelper(patch, b.dir)
		if err != nil {
			return err
		}
	}

	return nil
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

type BoshDeploymentResourceConfig struct {
	Target          string `yaml:"target"`
	Client          string `yaml:"client"`
	ClientSecret    string `yaml:"client_secret"`
	CaCert          string `yaml:"ca_cert"`
	JumpboxUrl      string `yaml:"jumpbox_url"`
	JumpboxSSHKey   string `yaml:"jumpbox_ssh_key"`
	JumpboxUsername string `yaml:"jumpbox_username"`
}

func (b StateDir) ExpungeInteropFiles() error {
	files := []string{"name", "metadata", "bdr-source-file"}
	for _, filename := range files {
		err := os.Remove(filepath.Join(b.dir, filename))
		if !os.IsNotExist(err) && err != nil {
			return err
		}
	}
	return nil
}

func (b StateDir) WriteInteropFiles(name string, c BoshDeploymentResourceConfig) error {
	err := ioutil.WriteFile(filepath.Join(b.dir, "name"), []byte(name), os.ModePerm)
	if err != nil {
		return err
	}
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(b.dir, "bdr-source-file"), []byte(bytes), os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(b.dir, "metadata"), []byte(bytes), os.ModePerm)
}

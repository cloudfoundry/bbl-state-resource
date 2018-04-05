package outrunner_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bbl-state-resource/outrunner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const sampleBblState = `
{
	"jumpbox": {
		"url": "nope.com"
	},
	"bosh": {
		"directorAddress": "da-address"
	}
}
`

const sampleJumpboxVarsStore = `
jumpbox_ssh:
  private_key: da-key
`

var _ = Describe("StateDir", func() {
	var (
		stateDir outrunner.StateDir

		tmpDir string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		tmpJson := filepath.Join(tmpDir, "bbl-state.json")
		err = ioutil.WriteFile(tmpJson, []byte(sampleBblState), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		stateDir = outrunner.NewStateDir(tmpDir)
	})

	Describe("Read", func() {
		It("reads the bbl state directory and returns the bbl state object", func() {
			bblState, err := stateDir.Read()
			Expect(err).NotTo(HaveOccurred())

			Expect(bblState.Jumpbox.URL).To(Equal("nope.com"))
			Expect(bblState.Director.Address).To(Equal("da-address"))
		})
	})

	Describe("JumpboxSSHKey", func() {
		BeforeEach(func() {
			varsDir := filepath.Join(tmpDir, "vars")
			err := os.Mkdir(varsDir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			varsStore := filepath.Join(varsDir, "jumpbox-vars-store.yml")
			err = ioutil.WriteFile(varsStore, []byte(sampleJumpboxVarsStore), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the jumpbox ssh key", func() {
			key, err := stateDir.JumpboxSSHKey()
			Expect(err).NotTo(HaveOccurred())

			Expect(key).To(Equal("da-key"))
		})
	})

	Describe("WriteMetadata", func() {
		It("writes a metadata file with contents", func() {
			err := stateDir.WriteMetadata("banana")
			Expect(err).NotTo(HaveOccurred())

			contents, err := ioutil.ReadFile(filepath.Join(tmpDir, "metadata"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("banana"))
		})
	})

	Describe("WriteName", func() {
		It("writes a name file with contents", func() {
			err := stateDir.WriteName("banana")
			Expect(err).NotTo(HaveOccurred())

			contents, err := ioutil.ReadFile(filepath.Join(tmpDir, "name"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("banana"))
		})
	})
})
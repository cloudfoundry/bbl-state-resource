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

const sampleMetadata = `target: target
client: da-client
client_secret: da-secret
ca_cert: da-cert
jumpbox_url: da-url
jumpbox_ssh_key: |-
  a-key
  with two lines
jumpbox_username: da-username
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

	Describe("InteropFiles", func() {
		var boshConfig outrunner.BoshDeploymentResourceConfig
		BeforeEach(func() {
			boshConfig = outrunner.BoshDeploymentResourceConfig{
				Target:          "target",
				Client:          "da-client",
				ClientSecret:    "da-secret",
				CaCert:          "da-cert",
				JumpboxUrl:      "da-url",
				JumpboxUsername: "da-username",
				JumpboxSSHKey: `a-key
with two lines`,
			}
		})

		Describe("WriteInteropFiles", func() {
			It("writes bosh-deployment-resource files", func() {
				err := stateDir.WriteInteropFiles("banana", boshConfig)
				Expect(err).NotTo(HaveOccurred())

				contents, err := ioutil.ReadFile(filepath.Join(tmpDir, "bdr-source-file"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal(sampleMetadata))

				contents, err = ioutil.ReadFile(filepath.Join(tmpDir, "metadata"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal(sampleMetadata))
			})

			It("writes pool-resource files", func() {
				err := stateDir.WriteInteropFiles("banana", boshConfig)
				Expect(err).NotTo(HaveOccurred())

				contents, err := ioutil.ReadFile(filepath.Join(tmpDir, "name"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("banana"))

				_, err = ioutil.ReadFile(filepath.Join(tmpDir, "metadata"))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("ExpungeInteropFiles", func() {
			Context("when the interop files are present", func() {
				BeforeEach(func() {
					err := stateDir.WriteInteropFiles("banana", boshConfig)
					Expect(err).NotTo(HaveOccurred())
				})

				It("deletes the interop files", func() {
					err := stateDir.ExpungeInteropFiles()
					Expect(err).NotTo(HaveOccurred())

					_, err = ioutil.ReadFile(filepath.Join(tmpDir, "bdr-source-file"))
					Expect(err).To(HaveOccurred())

					_, err = ioutil.ReadFile(filepath.Join(tmpDir, "metadata"))
					Expect(err).To(HaveOccurred())

					_, err = ioutil.ReadFile(filepath.Join(tmpDir, "name"))
					Expect(err).To(HaveOccurred())
				})
			})

			It("doesn't error when the interop files are not present", func() {
				err := stateDir.ExpungeInteropFiles()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

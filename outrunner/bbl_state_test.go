package outrunner_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bbl-state-resource/outrunner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StateDir", func() {
	var (
		stateDir outrunner.StateDir

		tmpDir string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		tmpState := filepath.Join(tmpDir, "bbl-state.json")
		err = ioutil.WriteFile(tmpState, []byte(`{ "jumpbox": { "url": "nope.com" } }`), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		stateDir = outrunner.NewStateDir(tmpDir)
	})

	It("reads the bbl state directory and returns the bbl state object", func() {
		bblState, err := stateDir.Read()
		Expect(err).NotTo(HaveOccurred())

		Expect(bblState.Jumpbox.URL).To(Equal("nope.com"))
	})
})

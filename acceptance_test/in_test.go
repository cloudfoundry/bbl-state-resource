package acceptance_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/cloudfoundry/bbl-state-resource/concourse"
	"github.com/cloudfoundry/bbl-state-resource/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("in", func() {
	var (
		targetDir        string
		version          concourse.Version
		bblStateContents string
		inInput          *bytes.Buffer
	)

	BeforeEach(func() {
		inRequest := fmt.Sprintf(`{
			"source": {
				"name": "in-test-test-env",
				"iaas": "gcp",
				"gcp-region": "us-east1",
				"gcp-service-account-key": %s
			},
			"version": {"ref": "the-greatest"}
		}`, strconv.Quote(serviceAccountKey))

		var req concourse.InRequest
		err := json.Unmarshal([]byte(inRequest), &req)
		Expect(err).NotTo(HaveOccurred())
		// this client isn't well tested, so we're going
		// to violate some abstraction layers to test it here
		// against the real api
		client, err := storage.NewStorageClient(req.Source)
		Expect(err).NotTo(HaveOccurred())

		By("uploading a bogus bbl state with some unique contents", func() {
			uploadDir, err := ioutil.TempDir("", "upload_dir")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(uploadDir)
			filename := filepath.Join(uploadDir, "bbl-state.json")
			f, err := os.Create(filename)
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()

			bblStateContents = fmt.Sprintf(`{"version": 14, "randomDir": "%s"}`, uploadDir)
			_, err = f.Write([]byte(bblStateContents))
			Expect(err).NotTo(HaveOccurred())

			version, err = client.Upload(uploadDir)
			Expect(err).NotTo(HaveOccurred())
		})

		targetDir, err = ioutil.TempDir("", "in_test")
		Expect(err).NotTo(HaveOccurred())

		inInput = bytes.NewBuffer([]byte(inRequest))
	})

	AfterEach(func() {
		os.RemoveAll(targetDir) // ignore the error
	})

	It("downloads the latest specified version of the resource", func() {
		cmd := exec.Command(inBinaryPath, targetDir)
		cmd.Stdin = inInput
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 10).Should(gexec.Exit(0))
		Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf(`[{"ref":"%s"}]`, version.Ref)))
		f, err := os.Open(filepath.Join(targetDir, "bbl-state.json"))
		Expect(err).NotTo(HaveOccurred())
		Eventually(gbytes.BufferReader(f)).Should(gbytes.Say(bblStateContents))
	})
})

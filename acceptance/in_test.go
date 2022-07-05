package acceptance_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/bbl-state-resource/concourse"
	"github.com/cloudfoundry/bbl-state-resource/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("in", func() {
	var (
		targetDir        string
		version          storage.Version
		bblStateContents string
		inInput          *bytes.Buffer
		marshalledSource []byte
		name             string
	)

	BeforeEach(func() {
		bucket := fmt.Sprintf("bsr-acc-tests-%s", projectId)
		source := concourse.Source{
			Bucket:               bucket,
			IAAS:                 "gcp",
			GCPRegion:            "us-east-1",
			GCPServiceAccountKey: serviceAccountKey,
		}
		var err error
		marshalledSource, err = json.Marshal(source)
		Expect(err).NotTo(HaveOccurred())

		// this client isn't well tested, so we're going
		// to violate some abstraction layers to test it here
		// against the real api
		name = fmt.Sprintf("bsr-test-in-%d-%s", GinkgoParallelProcess(), projectId)
		client, err := storage.NewStorageClient(serviceAccountKey, name, bucket)
		Expect(err).NotTo(HaveOccurred())

		By("uploading a bogus bbl state with some unique contents", func() {
			uploadDir, err := ioutil.TempDir("", "upload_dir")
			Expect(err).NotTo(HaveOccurred())
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
	})

	Context("when name is provided via the version", func() {
		BeforeEach(func() {
			inInput = bytes.NewBuffer([]byte(fmt.Sprintf(`{
				"source": %s,
				"params": {},
				"version": {
					"name": "%s",
					"ref": "the-greatest"
				}
			}`, marshalledSource, name)))
		})

		It("downloads the latest specified version of the resource", func() {
			cmd := exec.Command(inBinaryPath, targetDir)
			cmd.Stdin = inInput
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 10).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf(`{"version":{"name":"%s","ref":"%s","updated":".+"}}`, name, version.Ref)))
			f, err := os.Open(filepath.Join(targetDir, "bbl-state.json"))
			Expect(err).NotTo(HaveOccurred())
			Eventually(gbytes.BufferReader(f)).Should(gbytes.Say(bblStateContents))
		})
	})
})

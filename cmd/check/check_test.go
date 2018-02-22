package main_test

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

var _ = Describe("check", func() {
	var (
		targetDir         string
		bblStateContents  string
		serviceAccountKey string
		check             *bytes.Buffer
		version           concourse.Version
	)

	BeforeEach(func() {
		Expect(os.Getenv("BBL_GCP_SERVICE_ACCOUNT_KEY")).NotTo(Equal(""))

		var err error
		serviceAccountKey, err = readGCPServiceAccountKey(os.Getenv("BBL_GCP_SERVICE_ACCOUNT_KEY"))
		Expect(err).NotTo(HaveOccurred())

		checkRequest := fmt.Sprintf(`{
			"source": {
				"name": "check-test-test-env",
				"iaas": "gcp",
				"gcp-region": "us-east1",
				"gcp-service-account-key": %s
			},
			"version": {"ref": "the-greatest"}
		}`, strconv.Quote(serviceAccountKey))

		var req concourse.InRequest
		err = json.Unmarshal([]byte(checkRequest), &req)
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

			bblStateContents = fmt.Sprintf(`{"randomDir": "%s"}`, uploadDir)
			_, err = f.Write([]byte(bblStateContents))
			Expect(err).NotTo(HaveOccurred())

			version, err = client.Upload(uploadDir)
			Expect(err).NotTo(HaveOccurred())
		})

		targetDir, err = ioutil.TempDir("", "check_test")
		Expect(err).NotTo(HaveOccurred())

		check = bytes.NewBuffer([]byte(checkRequest))
	})

	AfterEach(func() {
		os.RemoveAll(targetDir) // ignore the error
	})

	It("prints the latest version of the resource", func() {
		cmd := exec.Command(checkBinaryPath, targetDir)
		cmd.Stdin = check
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 10).Should(gexec.Exit(0))
		Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf(`[{"ref":"%s"}]`, version.Ref)))
	})

	Context("when there is nothing stored in gcp", func() {
		BeforeEach(func() {
			checkRequest := fmt.Sprintf(`{
				"source": {
					"name": "empty-bucket-check-test",
					"iaas": "gcp",
					"gcp-region": "us-east1",
					"gcp-service-account-key": %s
				},
				"version": {"ref": "the-greatest"}
			}`, strconv.Quote(serviceAccountKey))
			check = bytes.NewBuffer([]byte(checkRequest))
		})

		It("prints an empty json list", func() {
			cmd := exec.Command(checkBinaryPath, targetDir)
			cmd.Stdin = check

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 10).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(Equal(`[]`))
		})
	})
})

func getGCPServiceAccountKey(key string) (string, error) {
	if _, err := os.Stat(key); err != nil {
		return key, nil
	}
	return readGCPServiceAccountKey(key)
}

func readGCPServiceAccountKey(path string) (string, error) {
	keyBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("Reading service account key: %v", err)
	}
	return string(keyBytes), nil
}

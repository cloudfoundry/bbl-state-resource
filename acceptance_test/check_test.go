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
	"time"

	"github.com/cloudfoundry/bbl-state-resource/concourse"
	"github.com/cloudfoundry/bbl-state-resource/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("check", func() {
	Context("when there is an environment in gcp", func() {
		var (
			bblStateContents string
			version          storage.Version
			marshalledSource []byte
			name             string
			timestamp        string
		)

		buildStorageClient := func(envName string) storage.StorageClient {
			bucketName := fmt.Sprintf("bsr-check-test-%d", GinkgoParallelNode())
			source := concourse.Source{
				Bucket:               bucketName,
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
			client, err := storage.NewStorageClient(serviceAccountKey, envName, bucketName)
			Expect(err).NotTo(HaveOccurred())
			return client
		}

		uploadBogusState := func(envName string) storage.Version {
			client := buildStorageClient(envName)
			var result storage.Version
			By("uploading a bogus bbl state with some unique contents", func() {
				uploadDir, err := ioutil.TempDir("", "upload_dir")
				Expect(err).NotTo(HaveOccurred())
				filename := filepath.Join(uploadDir, "bbl-state.json")
				f, err := os.Create(filename)
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()

				bblStateContents = fmt.Sprintf(`{"randomDir": "%s"}`, uploadDir)
				_, err = f.Write([]byte(bblStateContents))
				Expect(err).NotTo(HaveOccurred())

				result, err = client.Upload(uploadDir)
				Expect(err).NotTo(HaveOccurred())
			})
			return result
		}

		BeforeEach(func() {
			name = fmt.Sprintf("a-bsr-test-check-%d-%s", GinkgoParallelNode(), projectId)
			version = uploadBogusState(name)
			timestamp = version.Updated.Format(time.RFC3339Nano)
		})

		AfterEach(func() {
			err := buildStorageClient(name).DeleteBucket()
			Expect(err).NotTo(HaveOccurred())
		})

		It("prints the latest version of the environment", func() {
			checkRequest := fmt.Sprintf(`{
				"source": %s,
				"version": {
					"name": "%s",
					"ref": "the-greatest"
				}
			}`, marshalledSource, name)
			checkInput := bytes.NewBuffer([]byte(checkRequest))

			cmd := exec.Command(checkBinaryPath)
			cmd.Stdin = checkInput
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 10).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say(
				fmt.Sprintf(
					`\[{"name":"%s","ref":"%s","updated":"%s"}\]`,
					name, version.Ref, timestamp,
				),
			))
		})

		Context("when there are multiple environments in gcp", func() {
			var (
				newerName      string
				newerVersion   storage.Version
				newerTimestamp string
			)
			BeforeEach(func() {
				newerName = fmt.Sprintf("b-bsr-test-check-%d-%s", GinkgoParallelNode(), projectId)
				newerVersion = uploadBogusState(newerName)
				newerTimestamp = newerVersion.Updated.Format(time.RFC3339Nano)
			})

			AfterEach(func() {
				err := buildStorageClient(name).DeleteBucket()
				Expect(err).NotTo(HaveOccurred())
			})

			It("prints the latest version of each environment", func() {
				checkRequest := fmt.Sprintf(`{
					"source": %s,
					"version": {
						"name": "%s",
						"ref": "the-greatest"
					}
				}`, marshalledSource, name)
				checkInput := bytes.NewBuffer([]byte(checkRequest))

				cmd := exec.Command(checkBinaryPath)
				cmd.Stdin = checkInput
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10).Should(gexec.Exit(0))
				Eventually(session.Out).Should(gbytes.Say(
					fmt.Sprintf(
						`\[{"name":"%s","ref":"%s","updated":"%s"},{"name":"%s","ref":"%s","updated":"%s"}\]`,
						name, version.Ref, timestamp, newerName, newerVersion.Ref, newerTimestamp,
					),
				))
			})

			It("doesn't print versions for environments that are older than the one in the check request", func() {
				checkRequest := fmt.Sprintf(`{
					"source": %s,
					"version": {
						"name": "%s",
						"ref": "the-greatest",
						"updated": "%s"
					}
				}`, marshalledSource, newerName, newerTimestamp)
				checkInput := bytes.NewBuffer([]byte(checkRequest))

				cmd := exec.Command(checkBinaryPath)
				cmd.Stdin = checkInput
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10).Should(gexec.Exit(0))
				Eventually(session.Out).Should(gbytes.Say(
					fmt.Sprintf(
						`\[{"name":"%s","ref":"%s","updated":"%s"}\]`,
						newerName, newerVersion.Ref, newerTimestamp,
					),
				))
			})
		})
	})

	Context("when there is nothing stored in gcp", func() {
		It("prints an empty json list", func() {
			checkRequest := fmt.Sprintf(`{
				"source": {
					"bucket": "bsr-test-empty-%s",
					"iaas": "gcp",
					"gcp-region": "us-east1",
					"gcp-service-account-key": %s
				},
				"version": {"ref": "the-greatest"}
			}`, projectId, strconv.Quote(serviceAccountKey))

			cmd := exec.Command(checkBinaryPath)
			cmd.Stdin = bytes.NewBuffer([]byte(checkRequest))

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 10).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(`[]`))
		})
	})
})

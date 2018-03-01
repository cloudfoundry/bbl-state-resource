package acceptance_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("out", func() {
	Context("bbl succeeds", func() {
		var (
			upTargetDir string
			upOutInput  *bytes.Buffer

			downTargetDir string
			downOutInput  *bytes.Buffer
		)

		BeforeEach(func() {
			upRequest := fmt.Sprintf(`{
				"source": {
					"name": "%s-bsr-test-out",
					"iaas": "gcp",
					"gcp-region": "us-east1",
					"gcp-service-account-key": %s
				},
				"params": {
					"command": "up"
				}
			}`, projectId, strconv.Quote(serviceAccountKey))

			var err error
			upTargetDir, err = ioutil.TempDir("", "up_out_test")
			Expect(err).NotTo(HaveOccurred())
			upOutInput = bytes.NewBuffer([]byte(upRequest))

			downRequest := fmt.Sprintf(`{
				"source": {
					"name": "%s-bsr-test-out",
					"iaas": "gcp",
					"gcp-region": "us-east1",
					"gcp-service-account-key": %s
				},
				"params": {
					"command": "down"
				}
			}`, projectId, strconv.Quote(serviceAccountKey))

			downTargetDir, err = ioutil.TempDir("", "down_out_test")
			Expect(err).NotTo(HaveOccurred())
			downOutInput = bytes.NewBuffer([]byte(downRequest))
		})

		AfterEach(func() {
			By("bbling down", func() {
				cmd := exec.Command(outBinaryPath, downTargetDir)
				cmd.Stdin = downOutInput
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 40*time.Minute).Should(gexec.Exit(0), "bbl down should've suceeded!")
				Eventually(session.Out).Should(gbytes.Say(`{"version":{"ref":"[0-9a-f]+"}}`))
				_, err = os.Stat(filepath.Join(downTargetDir, "bbl-state.json"))
				Expect(err).To(HaveOccurred())
			})
		})

		It("bbls up and down successfully from different dirs configured with the same source", func() {
			By("bbling up", func() {
				cmd := exec.Command(outBinaryPath, upTargetDir)
				cmd.Stdin = upOutInput
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 40*time.Minute).Should(gexec.Exit(0), "bbl up should've suceeded!")
				Eventually(session.Out).Should(gbytes.Say(`{"version":{"ref":"[0-9a-f]+"}}`))
				_, err = os.Open(filepath.Join(upTargetDir, "bbl-state.json"))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	// at the moment, this is not easily testable.
	// it would require a way for bbl to consistently fail on the first command but also
	// write files, OR a way to delete uploaded state in the aftereach so that we can
	// inject a fake file like the test does as it is written now
	PContext("bbl exits 1 due to misconfiguration", func() {
		It("still uploads the failed state", func() {
			var (
				badInput io.Reader
			)
			badRequest := fmt.Sprintf(`{
				"source": {
					"name": "%s-bsr-test-bad",
					"iaas": "gcp",
					"gcp-region": "us-east1",
					"gcp-service-account-key": %s
				},
				"params": {
					"command": "make-bbl-fail"
				}
			}`, projectId, strconv.Quote(serviceAccountKey))
			badInput = bytes.NewBuffer([]byte(badRequest))
			putTargetDir, err := ioutil.TempDir("", "bad_out_test")
			Expect(err).NotTo(HaveOccurred())

			bblStateContents := fmt.Sprintf(`{ "fake-json": "%s" }`, putTargetDir)
			By("putting a bogus bbl-state.json into a bblStateContents", func() {
				bblStatePath := filepath.Join(putTargetDir, "not-bbl-state.json")
				err = ioutil.WriteFile(bblStatePath, []byte(bblStateContents), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			By("bbling up", func() {
				cmd := exec.Command(outBinaryPath, putTargetDir)
				cmd.Stdin = badInput
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10).Should(gexec.Exit(1), "bbl up should've failed when we misconfigured it")
				Eventually(session.Out).Should(gbytes.Say(`{"version":{"ref":"[0-9a-f]+"}}`))
			})

			By("getting the resource again", func() {
				inRequest := fmt.Sprintf(`{
					"source": {
						"name": "%s-bsr-test-bad",
						"iaas": "gcp",
						"gcp-region": "us-east1",
						"gcp-service-account-key": %s
					},
					"version": {"ref": "the-greatest"}
				}`, projectId, strconv.Quote(serviceAccountKey))

				getTargetDir, err := ioutil.TempDir("", "bad_test")
				Expect(err).NotTo(HaveOccurred())

				cmd := exec.Command(inBinaryPath, getTargetDir)
				cmd.Stdin = bytes.NewBuffer([]byte(inRequest))
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10).Should(gexec.Exit(0))
				Eventually(session.Out).Should(gbytes.Say(`{"version":{"ref":"[0-9a-f]+"}}`))

				f, err := os.Open(filepath.Join(getTargetDir, "not-bbl-state.json"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(gbytes.BufferReader(f)).Should(gbytes.Say(bblStateContents))
			})
		})
	})
})

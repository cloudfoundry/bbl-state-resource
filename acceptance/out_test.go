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
			name         string
			upSourcesDir string
			upOutInput   *bytes.Buffer

			downSourcesDir string
			downOutInput   *bytes.Buffer
		)

		BeforeEach(func() {
			name = fmt.Sprintf("bsr-test-out-%d-%s", GinkgoParallelNode(), projectId)
			upRequest := fmt.Sprintf(`{
				"source": {
					"bucket": "bsr-acc-tests-%s",
					"iaas": "gcp",
					"gcp-region": "us-east1",
					"gcp-service-account-key": %s
				},
				"params": {
					"name": "%s",
					"command": "up"
				}
			}`, projectId, strconv.Quote(serviceAccountKey), name)

			var err error
			upSourcesDir, err = ioutil.TempDir("", "up_out_test")
			Expect(err).NotTo(HaveOccurred())
			upOutInput = bytes.NewBuffer([]byte(upRequest))

			downRequest := fmt.Sprintf(`{
				"source": {
					"bucket": "bsr-acc-tests-%s",
					"iaas": "gcp",
					"gcp-region": "us-east1",
					"gcp-service-account-key": %s
				},
				"params": {
					"name": "%s",
					"command": "down"
				}
			}`, projectId, strconv.Quote(serviceAccountKey), name)

			downSourcesDir, err = ioutil.TempDir("", "down_out_test")
			Expect(err).NotTo(HaveOccurred())
			downOutInput = bytes.NewBuffer([]byte(downRequest))
		})

		AfterEach(func() {
			By("bbling down", func() {
				cmd := exec.Command(outBinaryPath, downSourcesDir)
				cmd.Stdin = downOutInput
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 40*time.Minute).Should(gexec.Exit(0), "bbl down should've suceeded!")
				Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf(`{"version":{"name":"%s","ref":".+","updated":".+"}}`, name)))
				_, err = os.Stat(filepath.Join(downSourcesDir, "bbl-state", "bbl-state.json"))
				Expect(err).To(HaveOccurred())

				_, err = ioutil.ReadFile(filepath.Join(downSourcesDir, "bbl-state", "metadata"))
				Expect(err).To(HaveOccurred())
			})
		})

		It("bbls up and down successfully from different dirs configured with the same source", func() {
			By("bbling up", func() {
				cmd := exec.Command(outBinaryPath, upSourcesDir)
				cmd.Stdin = upOutInput
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 40*time.Minute).Should(gexec.Exit(0), "bbl up should've suceeded!")
				Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf(`{"version":{"name":"%s","ref":".+","updated":".+"}}`, name)))
				_, err = os.Open(filepath.Join(upSourcesDir, "bbl-state", "bbl-state.json"))
				Expect(err).NotTo(HaveOccurred())

				bytes, err := ioutil.ReadFile(filepath.Join(upSourcesDir, "bbl-state", "metadata"))
				Expect(err).NotTo(HaveOccurred())
				Expect(bytes).To(ContainSubstring("target:"))
				Expect(bytes).To(ContainSubstring("client_secret:"))
				Expect(bytes).To(ContainSubstring("jumpbox_ssh_key:"))
			})
		})
	})

	Context("bbl exits 1 due to misconfiguration", func() {
		var (
			sourcesDir string
			name       string
			badInput   io.Reader

			badTerraformContents string
		)
		BeforeEach(func() {
			var err error
			sourcesDir, err = ioutil.TempDir("", "bad_out_test")
			Expect(err).NotTo(HaveOccurred())
			stateDir := filepath.Join(sourcesDir, "bbl-state")
			err = os.MkdirAll(filepath.Join(stateDir, "terraform"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			name = fmt.Sprintf("bsr-test-bad-out-%d-%s", GinkgoParallelNode(), projectId)
			badRequest := fmt.Sprintf(`{
				"source": {
					"bucket": "bsr-acc-tests-%s",
					"iaas": "gcp",
					"gcp-region": "us-east1",
					"gcp-service-account-key": %s
				},
				"params": {
					"command": "up",
					"name": "%s",
					"state_dir": "bbl-state"
				}
			}`, projectId, strconv.Quote(serviceAccountKey), name)
			badInput = bytes.NewBuffer([]byte(badRequest))

			badTerraformContents = fmt.Sprintf(`trololololol {}{{{{ %s`, sourcesDir)
			By("putting some bad terraform into a terraform override", func() {
				bblStatePath := filepath.Join(stateDir, "terraform", "broken_override.tf")
				err = ioutil.WriteFile(bblStatePath, []byte(badTerraformContents), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("still uploads the failed state", func() {
			By("bbling up", func() {
				cmd := exec.Command(outBinaryPath, sourcesDir)
				cmd.Stdin = badInput
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10).Should(gexec.Exit(1), "bbl up should've failed when we misconfigured it")
				Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf(`{"version":{"name":"%s","ref":".+","updated":".+"}}`, name)))

				_, err = os.Open(filepath.Join(sourcesDir, "bbl-state", "bdr-source-file"))
				Expect(err).NotTo(HaveOccurred())
			})

			By("getting the resource again", func() {
				inRequest := fmt.Sprintf(`{
					"source": {
						"bucket": "bsr-acc-tests-%s",
						"iaas": "gcp",
						"gcp-region": "us-east1",
						"gcp-service-account-key": %s
					},
					"version": {
						"name": "%s",
						"ref": "the-greatest"
					}
				}`, projectId, strconv.Quote(serviceAccountKey), name)

				getTargetDir, err := ioutil.TempDir("", "bad_test")
				Expect(err).NotTo(HaveOccurred())

				cmd := exec.Command(inBinaryPath, getTargetDir)
				cmd.Stdin = bytes.NewBuffer([]byte(inRequest))
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10).Should(gexec.Exit(0))
				Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf(`{"version":{"name":"%s","ref":".+","updated":".+"}}`, name)))

				f, err := os.Open(filepath.Join(getTargetDir, "terraform", "broken_override.tf"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(gbytes.BufferReader(f)).Should(gbytes.Say(badTerraformContents))
			})
		})
	})
})

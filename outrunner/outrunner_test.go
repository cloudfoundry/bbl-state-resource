package outrunner_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"

	"github.com/cloudfoundry/bbl-state-resource/concourse"
	"github.com/cloudfoundry/bbl-state-resource/fakes"
	"github.com/cloudfoundry/bbl-state-resource/outrunner"
	"github.com/fatih/structs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Run", func() {
	var (
		commandRunner *fakes.CommandRunner
		outRequest    concourse.OutRequest
		stateDir      string
	)

	BeforeEach(func() {
		commandRunner = &fakes.CommandRunner{}
		var err error
		stateDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	It("runs bbl up with the appropriate inputs", func() {
		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				IAAS:                 "some-iaas",
				LBType:               "some-lb-type",
				LBDomain:             "some-lb-domain",
				GCPServiceAccountKey: strconv.Quote(`{"some-json": "object"}`),
				GCPRegion:            "some-region",
			},
			Params: concourse.OutParams{
				Command: "up",
				Args: structs.Map(concourse.UpArgs{
					LBCert: "some-lb-cert",
					LBKey:  "some-lb-key",
				}),
			},
		}

		err := outrunner.RunInjected(commandRunner, "some-env-name", stateDir, outRequest.Params.Command, outRequest.Params.Args)
		Expect(err).NotTo(HaveOccurred())

		Expect(commandRunner.RunCall.Receives.Command).To(Equal("up"))
		Expect(commandRunner.RunCall.Receives.Args).To(ConsistOf(
			"--name=some-env-name",
			"--lb-cert=some-lb-cert",
			"--lb-key=some-lb-key",
			fmt.Sprintf("--state-dir=%s", stateDir),
		))

		contents, err := ioutil.ReadFile(filepath.Join(stateDir, "name"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(ContainSubstring("some-env-name"))

		contents, err = ioutil.ReadFile(filepath.Join(stateDir, "metadata"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(ContainSubstring("some-env-name"))
	})

	Context("without optional args", func() {
		It("omits the corresponding flags", func() {
			outRequest = concourse.OutRequest{
				Source: concourse.Source{
					IAAS:                 "some-iaas",
					GCPServiceAccountKey: strconv.Quote(`{"some-json": "object"}`),
					GCPRegion:            "some-region",
				},
				Params: concourse.OutParams{
					Command: "up",
					Args:    structs.Map(concourse.UpArgs{}),
				},
			}
			err := outrunner.RunInjected(commandRunner, "some-env-name", stateDir, outRequest.Params.Command, outRequest.Params.Args)
			Expect(err).NotTo(HaveOccurred())

			Expect(commandRunner.RunCall.Receives.Command).To(Equal("up"))
			Expect(commandRunner.RunCall.Receives.Args).To(ConsistOf(
				"--name=some-env-name",
				fmt.Sprintf("--state-dir=%s", stateDir),
			))
		})
	})
})

package outrunner_test

import (
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
	)

	BeforeEach(func() {
		commandRunner = &fakes.CommandRunner{}
	})

	It("runs bbl up with the appropriate inputs", func() {
		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				Name:                 "some-env-name",
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

		err := outrunner.RunInjected(commandRunner, outRequest, "some-state-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(commandRunner.RunCall.Receives.Command).To(Equal("up"))
		Expect(commandRunner.RunCall.Receives.Args).To(ConsistOf(
			"--name=some-env-name",
			"--lb-type=some-lb-type",
			"--lb-domain=some-lb-domain",
			"--lb-cert=some-lb-cert",
			"--lb-key=some-lb-key",
			`--gcp-service-account-key="{\"some-json\": \"object\"}"`,
			"--gcp-region=some-region",
			"--iaas=some-iaas",
			"--state-dir=some-state-dir",
		))
	})

	Context("without optional args", func() {
		It("omits the corresponding flags", func() {
			outRequest = concourse.OutRequest{
				Source: concourse.Source{
					Name:                 "some-env-name",
					IAAS:                 "some-iaas",
					GCPServiceAccountKey: strconv.Quote(`{"some-json": "object"}`),
					GCPRegion:            "some-region",
				},
				Params: concourse.OutParams{
					Command: "up",
					Args:    structs.Map(concourse.UpArgs{}),
				},
			}
			err := outrunner.RunInjected(commandRunner, outRequest, "some-state-dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(commandRunner.RunCall.Receives.Command).To(Equal("up"))
			Expect(commandRunner.RunCall.Receives.Args).To(ConsistOf(
				"--name=some-env-name",
				`--gcp-service-account-key="{\"some-json\": \"object\"}"`,
				"--gcp-region=some-region",
				"--iaas=some-iaas",
				"--state-dir=some-state-dir",
			))
		})
	})
})

package outrunner_test

import (
	"github.com/cloudfoundry/bbl-state-resource/concourse"
	"github.com/cloudfoundry/bbl-state-resource/fakes"
	"github.com/cloudfoundry/bbl-state-resource/outrunner"
	"github.com/fatih/structs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Run", func() {
	var (
		stateDir      *fakes.StateDir
		commandRunner *fakes.CommandRunner
	)

	BeforeEach(func() {
		stateDir = &fakes.StateDir{}
		stateDir.PathCall.Returns.Path = "some-bbl-state-dir"
		commandRunner = &fakes.CommandRunner{}
	})

	Context("with optional flags", func() {
		var params concourse.OutParams
		BeforeEach(func() {
			params = concourse.OutParams{
				Command: "up",
				Args: structs.Map(concourse.UpArgs{
					LBCert: "some-lb-cert",
					LBKey:  "some-lb-key",
				}),
			}
		})

		It("runs bbl up with the appropriate inputs", func() {
			err := outrunner.RunInjected(commandRunner, "some-env-name", stateDir, params.Command, params.Args)
			Expect(err).NotTo(HaveOccurred())

			Expect(commandRunner.RunCall.Receives.Command).To(Equal("up"))
			Expect(commandRunner.RunCall.Receives.Args).To(ConsistOf(
				"--name=some-env-name",
				"--lb-cert=some-lb-cert",
				"--lb-key=some-lb-key",
				"--state-dir=some-bbl-state-dir",
			))

			Expect(stateDir.WriteNameCall.CallCount).To(Equal(1))
			Expect(stateDir.WriteNameCall.Receives.Name).To(Equal("some-env-name"))

			Expect(stateDir.WriteMetadataCall.CallCount).To(Equal(1))
			Expect(stateDir.WriteMetadataCall.Receives.Metadata).To(Equal("some-env-name"))
		})
	})

	Context("without optional args", func() {
		var params concourse.OutParams
		BeforeEach(func() {
			params = concourse.OutParams{
				Command: "up",
				Args:    structs.Map(concourse.UpArgs{}),
			}
		})

		It("omits the corresponding flags", func() {
			err := outrunner.RunInjected(commandRunner, "some-env-name", stateDir, params.Command, params.Args)
			Expect(err).NotTo(HaveOccurred())

			Expect(commandRunner.RunCall.Receives.Command).To(Equal("up"))
			Expect(commandRunner.RunCall.Receives.Args).To(ConsistOf(
				"--name=some-env-name",
				"--state-dir=some-bbl-state-dir",
			))
		})
	})
})

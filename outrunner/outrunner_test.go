package outrunner_test

import (
	"errors"
	"os"

	"github.com/cloudfoundry/bbl-state-resource/concourse"
	"github.com/cloudfoundry/bbl-state-resource/fakes"
	"github.com/cloudfoundry/bbl-state-resource/outrunner"
	. "github.com/onsi/ginkgo/v2"
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
		stateDir.ReadCall.Returns.BblState.Jumpbox.URL = "some-jumpbox"
		stateDir.ReadCall.Returns.BblState.Director.Address = "some-director"
		stateDir.JumpboxSSHKeyCall.Returns.Key = "some-ssh-key"
		commandRunner = &fakes.CommandRunner{}
	})

	Context("with optional flags", func() {
		var params concourse.OutParams
		BeforeEach(func() {
			params = concourse.OutParams{
				Command: "up",
				Args: map[string]interface{}{
					"lb-key":  "some-lb-key",
					"lb-cert": "some-lb-cert",
				},
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
		})

		It("writes out the correct metadata for interoperation with other concourse resources", func() {
			err := outrunner.RunInjected(commandRunner, "some-env-name", stateDir, params.Command, params.Args)
			Expect(err).NotTo(HaveOccurred())

			Expect(stateDir.WriteInteropFilesCall.CallCount).To(Equal(1))
			Expect(stateDir.WriteInteropFilesCall.Receives.Config).To(Equal(
				outrunner.BoshDeploymentResourceConfig{
					Target:          "some-director",
					JumpboxUrl:      "some-jumpbox",
					JumpboxSSHKey:   "some-ssh-key",
					JumpboxUsername: "jumpbox",
				}),
			)
		})
	})

	Context("without optional args", func() {
		var params concourse.OutParams
		BeforeEach(func() {
			params = concourse.OutParams{
				Command: "up",
				Args:    map[string]interface{}{},
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

		Context("when the call to bbl fails", func() {
			BeforeEach(func() {
				commandRunner.RunCall.Returns.Error = errors.New("some-error")
			})

			It("errors", func() {
				err := outrunner.RunInjected(commandRunner, "some-env-name", stateDir, params.Command, params.Args)
				Expect(err).To(MatchError("failed running bbl up --state-dir=some-bbl-state-dir <sensitive flags omitted>: some-error"))
			})
		})

		Context("when we fail to read the bbl state", func() {
			BeforeEach(func() {
				stateDir.ReadCall.Returns.Error = errors.New("some-error")
			})

			It("does not error", func() {
				err := outrunner.RunInjected(commandRunner, "some-env-name", stateDir, params.Command, params.Args)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when we fail to read the bbl state because it does not exist", func() {
			BeforeEach(func() {
				stateDir.ReadCall.Returns.Error = os.ErrNotExist
			})

			It("does not error, but does expunge interop files", func() {
				err := outrunner.RunInjected(commandRunner, "some-env-name", stateDir, params.Command, params.Args)
				Expect(err).NotTo(HaveOccurred())
				Expect(stateDir.ExpungeInteropFilesCall.CallCount).To(Equal(1))
			})

			Context("when we fail to expunge interop files", func() {
				BeforeEach(func() {
					stateDir.ExpungeInteropFilesCall.Returns.Error = errors.New("wat")
				})

				It("does not error, and does not try to continue", func() {
					err := outrunner.RunInjected(commandRunner, "some-env-name", stateDir, params.Command, params.Args)
					Expect(err).NotTo(HaveOccurred())
					Expect(stateDir.WriteInteropFilesCall.CallCount).To(Equal(0))
					Expect(stateDir.JumpboxSSHKeyCall.CallCount).To(Equal(0))
				})
			})
		})

		Context("when we fail to fetch the jumpbox ssh key", func() {
			BeforeEach(func() {
				stateDir.JumpboxSSHKeyCall.Returns.Error = errors.New("some-error")
			})

			It("does not error", func() {
				err := outrunner.RunInjected(commandRunner, "some-env-name", stateDir, params.Command, params.Args)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

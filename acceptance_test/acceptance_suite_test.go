package acceptance_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Test Suite")
}

var (
	checkBinaryPath string
	inBinaryPath    string
	outBinaryPath   string
)

var _ = BeforeSuite(func() {
	var err error
	checkBinaryPath, err = gexec.Build("github.com/cloudfoundry/bbl-state-resource/cmd/in")
	Expect(err).NotTo(HaveOccurred())
	inBinaryPath, err = gexec.Build("github.com/cloudfoundry/bbl-state-resource/cmd/in")
	Expect(err).NotTo(HaveOccurred())
	outBinaryPath, err = gexec.Build("github.com/cloudfoundry/bbl-state-resource/cmd/in")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

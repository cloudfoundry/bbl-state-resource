package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Out Binary Suite")
}

var (
	outBinaryPath string
)

var _ = BeforeSuite(func() {
	var err error
	outBinaryPath, err = gexec.Build("github.com/cloudfoundry/bbl-state-resource/cmd/out")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

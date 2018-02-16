package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Check Binary Suite")
}

var (
	checkBinaryPath string
)

var _ = BeforeSuite(func() {
	var err error
	checkBinaryPath, err = gexec.Build("github.com/cloudfoundry/bbl-state-resource/cmd/check")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

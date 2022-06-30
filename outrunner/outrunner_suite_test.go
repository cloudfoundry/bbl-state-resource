package outrunner_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOutrunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Outrunner Suite")
}

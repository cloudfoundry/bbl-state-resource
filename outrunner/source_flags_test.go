package outrunner_test

import (
	"github.com/cloudfoundry/bbl-state-resource/concourse"
	"github.com/cloudfoundry/bbl-state-resource/outrunner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppendSourceFlags", func() {
	var (
		yeOldeFlags map[string]interface{}
		source      concourse.Source
	)

	BeforeEach(func() {
		yeOldeFlags = map[string]interface{}{
			"lb-cert": "/path/to/lb/cert",
			"lb-key":  "/path/to/lb/key",
		}
	})

	It("gets ye flags", func() {
		source = concourse.Source{
			Bucket:               "dont-flaggify-this-bucket!",
			IAAS:                 "gcp",
			GCPServiceAccountKey: "some-service-account-key",
			LBType:               "cf",
			LBDomain:             "cf.example.com",
		}
		flags := outrunner.AppendSourceFlags(yeOldeFlags, source)

		Expect(flags).To(HaveKeyWithValue("iaas", "gcp"))
		Expect(flags).To(HaveKeyWithValue("gcp-service-account-key", "some-service-account-key"))
		Expect(flags).To(HaveKeyWithValue("lb-type", "cf"))
		Expect(flags).To(HaveKeyWithValue("lb-domain", "cf.example.com"))
		Expect(flags).To(HaveKeyWithValue("lb-cert", "/path/to/lb/cert"))
		Expect(flags).To(HaveKeyWithValue("lb-key", "/path/to/lb/key"))
		Expect(flags).NotTo(HaveKey("bucket"))
	})

	Context("when optional flags are omitted", func() {
		It("omits those flags", func() {
			source = concourse.Source{
				Bucket:               "dont-flaggify-this-bucket!",
				IAAS:                 "gcp",
				GCPServiceAccountKey: "some-service-account-key",
			}
			flags := outrunner.AppendSourceFlags(yeOldeFlags, source)

			Expect(flags).To(HaveKeyWithValue("iaas", "gcp"))
			Expect(flags).To(HaveKeyWithValue("gcp-service-account-key", "some-service-account-key"))
			Expect(flags).NotTo(HaveKey("lb-type"))
			Expect(flags).NotTo(HaveKey("lb-domain"))
			Expect(flags).To(HaveKeyWithValue("lb-cert", "/path/to/lb/cert"))
			Expect(flags).To(HaveKeyWithValue("lb-key", "/path/to/lb/key"))
			Expect(flags).NotTo(HaveKey("bucket"))
		})
	})
})

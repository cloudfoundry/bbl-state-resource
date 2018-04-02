package concourse

type Source struct {
	Bucket string `json:"bucket,omitempty" yaml:"bucket" structs:"bucket,omitempty"`
	IAAS   string `json:"iaas,omitempty" yaml:"iaas" structs:"iaas,omitempty"`

	LBType   string `json:"lb-type,omitempty" yaml:"lb-type" structs:"lb-type,omitempty"`
	LBDomain string `json:"lb-domain,omitempty" yaml:"lb-domain" structs:"lb-domain,omitempty"`

	GCPServiceAccountKey string `json:"gcp-service-account-key,omitempty" yaml:"gcp-service-account-key" structs:"gcp-service-account-key,omitempty"`
	GCPRegion            string `json:"gcp-region,omitempty" yaml:"gcp-region" structs:"gcp-region,omitempty"`
}

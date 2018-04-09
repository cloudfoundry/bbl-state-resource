package concourse

type Source struct {
	Bucket string `json:"bucket,omitempty" yaml:"bucket"`
	IAAS   string `json:"iaas,omitempty" yaml:"iaas"`

	LBType   string `json:"lb_type,omitempty" yaml:"lb_type"`
	LBDomain string `json:"lb_domain,omitempty" yaml:"lb_domain"`

	GCPServiceAccountKey string `json:"gcp_service_account_key,omitempty" yaml:"gcp_service_account_key"`
	GCPRegion            string `json:"gcp_region,omitempty" yaml:"gcp_region"`
}

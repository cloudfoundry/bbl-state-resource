package outrunner

import "github.com/cloudfoundry/bbl-state-resource/concourse"

func AppendSourceFlags(flags map[string]interface{}, source concourse.Source) map[string]interface{} {
	if flags == nil {
		flags = map[string]interface{}{}
	}
	flags["iaas"] = source.IAAS
	flags["gcp-service-account-key"] = source.GCPServiceAccountKey
	if source.LBType != "" {
		flags["lb-type"] = source.LBType
	}
	if source.LBDomain != "" {
		flags["lb-domain"] = source.LBDomain
	}
	return flags
}

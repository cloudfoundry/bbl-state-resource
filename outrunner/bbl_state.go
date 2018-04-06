package outrunner

type BblState struct {
	Jumpbox  Jumpbox  `json:"jumpbox"`
	Director Director `json:"bosh"`
	EnvID    string   `json:"envID"`
}

type Jumpbox struct {
	URL string `json:"url"`
}

type Director struct {
	ClientUsername string `json:"directorUsername"`
	ClientSecret   string `json:"directorPassword"`
	Address        string `json:"directorAddress"`
	CaCert         string `json:"directorSSLCA"`
}

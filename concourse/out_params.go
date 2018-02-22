package concourse

// the args that will be passed to bbl
// in addition to all the ones configured in source.
// for example:
// ```
// OutRequest {
//     Source: Source{
//         IAAS: "gcp",
//     },
//     Params: OutParams{
//     		Command: "up",
//          Args: map[string]string{"lb-cert": "----some cert----"},
//     },
// }
// ```
// will exec:
// bbl -n up --iaas=gcp --lb-cert=----some cert----

type OutParams struct {
	Command string                 `json:"command"`
	Args    map[string]interface{} `json:"args"`
}

type UpArgs struct {
	LBCert string `json:"lb-cert" structs:"lb-cert,omitempty"`
	LBKey  string `json:"lb-key" structs:"lb-key,omitempty"`
}

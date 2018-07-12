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
//          Name: "some-env-name",
//     		Command: "up",
//          Args: map[string]string{"lb-cert": "----some cert----"},
//     },
// }
// ```
// will exec:
// bbl -n up --iaas=gcp --name some-env-name --lb-cert=----some cert----

type OutParams struct {
	Name        string                 `json:"name"`
	NameFile    string                 `json:"name_file"`
	StateDir    string                 `json:"state_dir"`
	Command     string                 `json:"command"`
	Args        map[string]interface{} `json:"args"`
	PlanPatches []string               `json:"plan-patches"`
}

type UpArgs struct {
	LBCert string `json:"lb-cert"`
	LBKey  string `json:"lb-key"`
}

# the bbl state dir concourse resource
this is a concourse resource for provisioning bosh directors and using [bbl](https://github.com/cloudfoundry/bosh-bootloader) to manipulate cloud-stored bbl-states.

today, it automates provisioning bosh directors and their associated iaas resources. It also integrates well with concourse/pool-resource.
tomorrow, we'd like our resource-provisioned bbl-states to easily plug into [bosh-deployment-resource](https://github.com/cloudfoundry/bosh-deployment-resource).

to bring this in to your concourse pipeline:
```yaml
resource_types:
- name: bbl-state-resource
  type: docker-image
  source:
    repository: cfinfrastructure/bbl-state-resource
```

## Source Configuration
note that bbl arg keys use dashes, not underscores.
#### Examples:
```
resources:
- name: bbl-state
  type: bbl-state-resource
  source:
    bucket: bbl-state
    iaas: gcp
    gcp-region: us-east-1
    gcp-service-account-key: {{bbl_gcp_service_account_key}}
```
### Parameters:
`bucket`: **required**: the name of the bucket where you'd like your state-dir tarballs to be stored.

`iaas`: **required**: gcp, for now, but we'll take aws soon. This is the iaas where you want your new bosh directors.

`lb-type`: optional: `cf` or `concourse`, denotes the varietals of the load balancers you'd like to deploy with your director

`lb-domain`: optional: for cf, the system domain, for concourse, the web domain. NOTE: randomly named bosh directors will share a single domain at the moment and that will not go well. these features don't mix.

`gcp-service-account-key`: **required**: your gcp service account key, formatted as JSON.

`gcp-region`: **required**: the gcp region where you'd like your environments.

## Behaviour
### `put`: Deploy, upgrade, and destroy BOSH directors and its containing environment

There are two primary modes of operation for bbl-state puts:
1. By default, without a name configured, we'll generate random environment names for each `put: { command: up }`.
1. If you've configured a name, name_file, or a state_dir, the resource will manipulate that environent.

#### Examples:
```yaml
jobs:
- name: bbl-up-a-specifically-named-environment
  plan:
  - put: bbl-state
    params:
      command: up
	  name: my-lonely-bosh-director

- name: bbl-up-a-randomly-named-environment
  plan:
  - put: bbl-state
    params:
      command: up

- name: bbl-update-that-randomly-named-env
  plan:
  - get: bbl-state
  	passed: [bbl-up-a-randomly-named-environment]
  - put: bbl-state
  	params:
	  command: up
	  name_file: bbl-state/name

- name: bbl-delete-that-env-you-just-updated
  plan:
  - get: bbl-state-resource
  	passed: [bbl-update-that-env-from-before]
  - put: bbl-state
  	params:
	  command: down
	  state_dir: bbl-state
```
#### Parameters:

`command`: **required**: `up`, `down`, `destroy`, `rotate`, or `cleanup-leftovers`. Any top-level command available to bbl.

`args`: optional: a yaml hash containing additional flags as key-value pairs. these might be load balancer options or `filter: env-name` for leftovers.

`name`: optional: the name of the environment you'd like to manipulate. overrides name_file and state_dir.

`name_file`: optional: a file you'd like to load name from, useful if you're manipulating an env stored in a pool-resource. overrides state_dir.

`state_dir`: optional: an already-fetched bbl state directory containing the state for the environment you'd like to manipulate.

### `get`: Download information about a BOSH director and its associated iaas environment

`get`s download bbl-state dirs.

#### Examples:
```yaml
jobs:
- name: get-from-previous-put-and-add-it-to-a-pool
  plan:
  - get: bbl-state
  	trigger: true
  	passed: [bbl-up-a-randomly-named-environment]
  - put: concourse-pool-of-bbl-states
  	params:
	  add: bbl-state

- name: delete-a-random-unclaimed-env-nightly
  plan:
  - get: time-resource-nightly
  	trigger: true
  - put: lock
  	params:
	  resource: concourse-pool-of-bbl-states
	  acquire: true
  - put: bbl-state
  	params:
	  name_file: lock/name
	  command: down
	on_success:
		resource: concourse-pool-of-bbl-states
		remove: lock
```
### Parameters:
none! names, checksums, and timestamps are encoded in our concourse versions, so we've got to fetch those specific ones.
> note: `get`s don't have access to the file system for dynamic configuration, anyways, so name_files and state_dirs can't be reached.
If you want to get a specific state-dir, you'll have to use concourse primitives like `passed` to filter things down or do a put with a noop-ish bbl command like `env-id`.

Special outputs that you wouldn't find in a normal bbl-state include:
1. `bbl-state/name`, which contains the environment name
1. `bbl-state/metadata`, which is useful for plugging in to concourse/pool-resource

TODO: additionally format target-dir/bdr-source-configuration + metadata for use by [bosh-deployment-resource](https://github.com/cloudfoundry/bosh-deployment-resource), and document.

## Development:

things happen via the Makefile:

make help

# bbl state-dir resource
a wippity concourse resource for manipulating bbl-states

## Source Configuration

## Behaviour

### `in`: Download information about a BOSH director and its associated iaas environment

This will download the gcp-stored bbl-state dir to the target directory,
additionally formatting target-dir/bdr-source-configuration for use by [bosh-deployment-resource](https://github.com/cloudfoundry/bosh-deployment-resource)

### `out`: Deploy, upgrade, and destroy a BOSH director and its associated iaas environment

all your bbl plans, ups, and destroys!

### `check`: emits new versions when "out" runs?

#### Source Parameters:

```yaml
resource_types:
- name: bbl-state-resource
  type: docker-image
  source:
    repository: cfinfrastructure/bbl-state-resource

resources:
- name: your-environment-name-bbl-state
  type: bbl-state-resource
  source:
    name: your-environment-name
    iaas: gcp
    gcp_region: us-east-1
    gcp_service_account_key: {{bbl_gcp_service_account_key}}
```

## Development:

things happen via the Makefile:

make help

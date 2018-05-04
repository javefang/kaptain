# Kaptain - Production Grade Kubernetes Cluster Deployment
 
## TL;DR;

Dependencies
- kubectl 1.7.4

```
$ kaptain create --name=dev.my-project.aws
$ kaptain list

$ sailor provision --name=dev.my-project.aws --role=etcd
$ sailor provision --name=dev.my-project.aws --role=master
$ sailor provision --name=dev.my-project.aws --role=worker
$ ls /tmp/kaptain

$ kaptain export --name=dev.my-project.aws

$ # Deploy the cluster with Terraform

$ kaptain bootstrap --name=dev.my-project.aws 
```

## Storage backend

Currently Kaptain support `S3` and `Vault` as storage backend. Storage backend
can be specified via command line flag `--store=<uri>`. The URI has the following
format.

`<scheme>://<path>?<key1>=<value1>&<key2>=<value2>`

Where

- `<scheme>` is the name of the storage backend
- `<path>` is the resource path to the storage, this is storage specific
- `<key>` and `<value>` are configs for the storage backend, also storage specific

The current default value is `s3://aws.all.kaptain?region=eu-west-1` if not specified.


### S3

Although AWSCLI is not required. AWS credentials must be set. See http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html for details.

- Scheme: `s3`
- Path: name of the bucket (e.g. `aws.all.kaptain`)
- Keys:
  - `region`: S3 region (e.g. `eu-west-1`)

Example `s3://aws.all.kaptain?region=eu-west-1`

### Vault

Environment variable `VAULT_ADDR` must be set to the Vault server.
(e.g. `https://vault.service.consul:8200`)

- Scheme: `vault`
- Path: vault generic secret path without the `secret/` prefix (e.g. `project/kaptain`)
- Keys:
  - `role_id`: Vault role ID
  - `role_secret`: Vault role secret

Example `vault://project/kaptain?role_id=1234&secret_id=abcd`

## Usage

Kaptain is a commandline tool to streamline management of various config files and
x509 certificates needed by a Kubernetes Deployment. The tool is split into two components:

### Kaptain

An admin tool that is run by a cluster admin (Support Linux, OSX and Windows), which

- Allows the admin to define a cluster with given a unique `cluster name`
- Generates all x509 certificates and secrets needed for bootstrapping a new cluster
- Persists all state on remote storage backend
- Allows the admin to export kubeconfig to the current machine (default to `~/.kube/config`)

### Sailor

An agent that runs by the CloudInit script on provisioned cluster nodes 
(etcd, master or worker). Given `cluster name` and `role` (etcd, master or worker), it can

- Fetch cluster definition from remote storage backend
- Provision all necessary config files and x509 certificates depending on node type

## Development

### Pre-requisites

- make
- Golang 1.10
- awscli

### Build

The run the build to compile and install `kaptain` and `sailor` locally.

```
$ make
```

You should be able to test the binaries now

```
$ kaptain version
$ sailor version
```

### Release

Tag current commit on the master branch.

```
$ git tag <version>
$ git push --tags
```

Cross-compile for all platforms. This will generate compiled binaries at `./build`

```
$ make release
```

### Dockerised Build

You can build the project completely in Docker, without the need to have Go toolchain installed.

```
$ make docker-build
```

### Cleanup

This will delete all compiled binaries at `./build`

```
$ make clean
```

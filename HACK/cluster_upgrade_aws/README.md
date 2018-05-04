# Kaptain - AWS Cluster Upgrade Scripts

This folder contains the scripts used for upgrading a Kaptain cluster in AWS.
As any other good Kubernetes tooling project, the functionalities are first
implmented in BASH scripts, and will be implemented in Golang and merge into
Kaptain once matured.

## Overview

### Upgrade Kubernetes

Upgrading a Kubernetes cluster consists of two parts:

1. Upgrade the Kubernetes Control Plane (apiserver, controller-manager, scheduler)
2. Upgrade each Kubernetes worker nodes (kubelet, kube-proxy)

Both are easy with the help of Kaptain and Immutable Infrastructure.

Although Kubernetes itself is easy to upgrade. Things get complicated when we
need to consider all the Pods running on it. For example, some Pods might get 
upset when being restarted. This set of scripts allow us to perform cluster
upgrade in a more controlled manner. 

## Assumption

The scripts deployed here has the following assumptions on the deployed cluster:

1. Deployed on AWS
2. Deployed with Terraform module `aws.infra.kube-node` (The branch is not merged yet, see [bitbucket](https://stash/projects/TRF/repos/aws.infra.kube-node/browse?at=refs%2Fheads%2Fsimple_asg))

## Dependencies

The following tools must be installed:

- AWS CLI
- jq
- kubectl

## Perform Cluster Upgrade

For example, to upgrade Kubernetes from version 1.8.3 to 1.8.4 on node group `worker`. 
The upgrade will be performed as follow:

1. Bring up a new set of nodes with 1.8.4 (this doubles the node count in the ASG).
2. Cordon all 1.8.3 nodes, so new Pods will not be scheduled to them.
3. Drain each 1.8.3 node, which Kubernetes will reschedule them on 1.8.4 nodes.
4. Verify everything works normally, and destroy all 1.8.3 nodes.

### 0. Update LaunchConfiguration with Terraform

After the LaunchConfiguration in Terraform have been updated with the new AMI (1.8.4).

```
$ cd terraform/vpc/dev.xinghong.waws
$ make plan
$ make apply
```

This should create a new `LaunchConfiguration` that will launch 1.8.4 and 
update the `AutoScalingGroup` to use that. Until now, no instances have 
been terminated or started. But any new instances will be started with the
new LC.

### 1. Provision

Check ASG status, it should show all nodes are outdated.

```
export AWS_VPC_NAME=<vpc_name>
$ ./status.sh worker
```

Provision new nodes (1.8.4) by scaling up the ASG from 3 to 6. And cordon all
old nodes (1.8.3).

```
$ ./provision.sh worker 6
```

Now you have 3 nodes with 1.8.3 (cordoned) and 3 nodes with 1.8.4. Use `kubectl`
to check node status and wait for all of them to become `Ready`.

### 2. Migrate

Now we have both old and new nodes running. Now we need to migrate the Pods.

```
$ ./migrate worker
```

This should drain the old nodes one-by-one so that the Pods will be restarted
on new nodes. The script will pause between two nodes so you can check the status
before carry on.

### 3. Cleanup

All Pods should be running on new nodes now. If they are all happy. Clean up the
old nodes.

```
$ ./cleanup.sh worker 3
```

This should scale the ASG back to 3, since the ASG has termination policy 
`OldestLaunchConfiguration`, the 3 old nodes will be destroyed. 

Now we have an updated node group! Repeat this process with other node groups.

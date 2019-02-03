# aws-infra-controller

The aws-infra-controller manages the AWS infrastructure for your Kubernetes cluster for you.

## Status

This project is pre-alpha.  The instructions below will spin up a single-node Kubernetes cluster in AWS but no useful infra management functionality has been developed yet.

## Prerequisites

* have [packer](https://www.packer.io/) installed
* have an AWS account with credentials for API access

## Build Machine Image

    $ cd cmd/bootctl
    $ env GOOS=linux go build
    $ mv bootctl ../../machine_images/ubuntu
    $ cd ../../machine_images/ubuntu
    $ packer build bootstrap_master_template.json

## Edit Cluster Config

Set the following in `config/samples/infra_v1alpha1_cluster.yaml`:
    * `metadata.name`: arbitrary name for your cluster
    * `spec.region`: AWS region

## Edit Machine Config

Set the following in `config/samples/infra_v1alpha1_machine.yaml`:
    * `ami`: add the AMI from the image you built in the previous section
    * `keyName`: a key in AWS you have access to

## Build bootctl

This build is for your system, as opposed to the one you built for the machine image.

    $ cd cmd/bootctl
    $ go build

## Install a Cluster

Run the create command with cluster and machine flags, write output to an inventory config in your home directory:

    $ ./bootctl create -c ../../config/samples/infra_v1alpha1_cluster.yaml -m ../../config/samples/infra_v1alpha1_machine.yaml > ~/.aws-infra-controller-inventory.json

## Connect to Cluster

Get the IP address from AWS and connect to your AWS instance:

    $ ssh ubuntu@[ip address]

View the cluster's nodes:

    $ kubectl get nodes

## Destroy the Cluster

Run the destroy command and reference the inventory file generated by the create command:

    $ ./bootctl destroy -i ~/.aws-infra-controller-inventory.json


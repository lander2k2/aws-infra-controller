# aws-infra-controller

The aws-infra-controller manages the AWS infrastructure for your Kubernetes cluster for you.

## Prerequisites

* have [packer](https://www.packer.io/) installed
* have an AWS account and credentials for API access in your environment

## Build Machine Image

    $ cd cmd/bootctl
    $ env GOOS=linux go build
    $ mv bootctl ../../machine_images/ubuntu
    $ cd ../../machine_images/ubuntu
    $ packer build bootstrap_master_template.json

## Set Cluster Variables

Edit `examples.cluster.yaml` and set the variables for your purposes:
    * `name`: an arbitary name for your cluster
    * `region`: the AWS region you'd like to use
    * `ami`: add the AMI from the image you built in the previous section
    * `keyName`: a key in AWS you have access to

## Build bootctl

This build is for your system, as opposed to the one you built for the machine image.

    $ cd cmd/bootctl
    $ go build

## Install a Cluster

    $ ./bootctl create examples/cluster.yaml /tmp/aws-infra-controller-inventory.json

## Destroy a Cluster

    $ ./bootctl destroy /tmp/aws-infra-controller-inventory.json


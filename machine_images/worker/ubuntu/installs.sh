#!/bin/bash

K8S_VERSION=1.13.3

sudo apt-get update
DEBIAN_FRONTEND=noninteractive sudo -E apt-get -yq upgrade
sudo apt-get install -y docker.io apt-transport-https

curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee -a /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
sudo apt-get install -y kubelet=${K8S_VERSION}-00 kubeadm=${K8S_VERSION}-00 kubectl=${K8S_VERSION}-00

sudo mv /tmp/bootctl /usr/local/bin/
sudo chmod +x /usr/local/bin/bootctl


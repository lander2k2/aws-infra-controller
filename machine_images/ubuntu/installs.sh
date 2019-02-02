#!/bin/bash

ETCD_VERSION=v3.2.24
K8S_VERSION=1.13.3

sudo apt-get update
DEBIAN_FRONTEND=noninteractive sudo -E apt-get -yq upgrade
sudo apt-get install -y docker.io apt-transport-https

curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee -a /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
sudo apt-get install -y kubelet=${K8S_VERSION}-00 kubeadm=${K8S_VERSION}-00 kubectl=${K8S_VERSION}-00

sudo curl -o /usr/local/bin/cfssl https://pkg.cfssl.org/R1.2/cfssl_linux-amd64
sudo curl -o /usr/local/bin/cfssljson https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64
sudo chmod +x /usr/local/bin/cfssl*

curl -sSL https://github.com/coreos/etcd/releases/download/${ETCD_VERSION}/etcd-${ETCD_VERSION}-linux-amd64.tar.gz | sudo tar -xzv --strip-components=1 -C /usr/local/bin/
rm -rf etcd-$ETCD_VERSION-linux-amd64*


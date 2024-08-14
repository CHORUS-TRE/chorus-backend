#!/bin/bash


set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
cd "$DIR"

OS=darwin
if [[ $(uname -s) == Linux ]]
then
    OS=linux
fi

PATH="$PATH:$PWD/scripts/tools/$OS/bin"

export KUBECONFIG="$DIR/configs/dev/files/kubeconfig.yaml"

clusters=$(kind get clusters)
exists=0
for cluster in $clusters; do
  if [ "$cluster" == "chorus" ]; then
    exists=1
    break
  fi
done

if [ $exists -eq 1 ]; then
    echo "Cluster chorus already exist, skipping create..."
else
    kind create cluster --name chorus --config configs/dev/files/kind-config.yaml
fi

# create workbench CRD
kubectl apply -f internal/client/helm/chart/workbench-crd.yaml

### TODO clean

# pull local image controller 
# TODO get it locally first
kind load docker-image controller:latest --name chorus

# TODO launch operator
cd /home/localadmin/chuv/chorus/workbench-operator
kubectl apply -k config/manager

cd "$DIR"
kubectl apply -f internal/client/helm/chart/roles.yaml
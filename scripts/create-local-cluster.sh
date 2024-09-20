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
docker login registry.build.chorus-tre.ch 
docker pull registry.build.chorus-tre.ch/backend/workbench-controller:latest
docker tag registry.build.chorus-tre.ch/backend/workbench-controller:latest controller:latest
docker pull registry.build.chorus-tre.ch/xpra-server:latest
kind load docker-image controller:latest --name chorus
kind load docker-image registry.build.chorus-tre.ch/xpra-server:latest --name chorus

kubectl apply -k internal/client/helm/chart/manager

cd "$DIR"
kubectl apply -f internal/client/helm/chart/roles.yaml
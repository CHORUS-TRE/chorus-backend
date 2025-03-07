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

export KUBECONFIG="$PWD/configs/dev/files/kubeconfig.yaml"

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

rm -rf workbench-operator
git clone git@github.com:CHORUS-TRE/workbench-operator.git
cd workbench-operator
make installdry OUT=tmp-workbench-crd.yaml
cp tmp-workbench-crd.yaml ..
cd ..

# create workbench CRD
kubectl apply -f tmp-workbench-crd.yaml
rm tmp-workbench-crd.yaml

### TODO clean

# pull local image controller
# docker login registry.build.chorus-tre.ch 
# docker pull registry.build.chorus-tre.ch/backend/workbench-controller:latest
# docker tag registry.build.chorus-tre.ch/backend/workbench-controller:latest controller:latest
# docker pull registry.build.chorus-tre.ch/xpra-server:latest
# kind load docker-image controller:latest --name chorus
# kind load docker-image registry.build.chorus-tre.ch/xpra-server:latest --name chorus

rm -rf tmpoperator
mkdir tmpoperator && cd tmpoperator
rm -rf environments
git clone git@github.com:CHORUS-TRE/environments.git
#curl -L -O https://github.com/CHORUS-TRE/environments/raw/refs/heads/master/chorus-dev/workbench-operator/values.yaml
#curl -L -O https://github.com/CHORUS-TRE/environments/raw/refs/heads/master/chorus-dev/workbench-operator/config.json
cp environments/chorus-dev/workbench-operator/values.yaml .
cp environments/chorus-dev/workbench-operator/config.json .
rm -rf environments

latest=$(cat config.json | jq -r '.version')

helm template -s templates/deployment.yaml -f values.yaml ../workbench-operator/charts/workbench-operator/ > controller-deployment.yaml
kubectl create namespace system --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -n system -f controller-deployment.yaml
rm controller-deployment.yaml

cd ..
rm -rf tmpoperator
rm -rf workbench-operator

cd "$DIR"
kubectl apply -n system -f internal/client/helm/chart/roles.yaml
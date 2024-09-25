#!/bin/bash

set -e

docker build --pull -f dockerfiles/stage2.dockerfile -t registry.dip-dev.thehip.app/chorus-cicd-chorus:${IMAGE_TAG} ..
docker push registry.dip-dev.thehip.app/chorus-cicd-chorus:${IMAGE_TAG}
if [[ "$$BRANCH_NAME" == "master" ]]; then
    docker tag registry.dip-dev.thehip.app/chorus-cicd-chorus:${IMAGE_TAG} registry.dip-dev.thehip.app/chorus-cicd-chorus:master
    docker push registry.dip-dev.thehip.app/chorus-cicd-chorus:master
fi
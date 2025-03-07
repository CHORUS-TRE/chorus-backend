#!/bin/bash

set -e

docker build --pull -f dockerfiles/stage2.dockerfile -t harbor.build.chorus-tre.ch/chorus/backend:${IMAGE_TAG} ..
docker push harbor.build.chorus-tre.ch/chorus/backend:${IMAGE_TAG}
if [[ "$BRANCH_NAME" == "master" ]]; then
    docker tag harbor.build.chorus-tre.ch/chorus/backend:${IMAGE_TAG} harbor.build.chorus-tre.ch/chorus/backend:master
    docker push harbor.build.chorus-tre.ch/chorus/backend:master
fi
#!/bin/bash

set -e

if [ -n "$GIT_USERNAME" -a -n "$GIT_PASSWORD" ]; then
    docker build --pull -f dockerfiles/backend.dockerfile --secret id=GIT_USERNAME,env=GIT_USERNAME --secret id=GIT_PASSWORD,env=GIT_PASSWORD -t harbor.build.chorus-tre.ch/chorus/backend:${IMAGE_TAG} ..
else
    docker build --pull -f dockerfiles/backend.dockerfile -t harbor.build.chorus-tre.ch/chorus/backend:${IMAGE_TAG} ..
fi

docker push harbor.build.chorus-tre.ch/chorus/backend:${IMAGE_TAG}
if [[ "$BRANCH_NAME" == "master" ]]; then
    docker tag harbor.build.chorus-tre.ch/chorus/backend:${IMAGE_TAG} harbor.build.chorus-tre.ch/chorus/backend:master
    docker push harbor.build.chorus-tre.ch/chorus/backend:master
fi
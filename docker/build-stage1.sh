#!/bin/bash

set -e

docker build --pull -f dockerfiles/stage1.dockerfile -t harbor.build.chorus-tre.ch/chorus/backend-stage1 ..
docker push harbor.build.chorus-tre.ch/chorus/backend-stage1
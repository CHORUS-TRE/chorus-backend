#!/bin/bash

set -e

docker build --pull -f dockerfiles/stage1.dockerfile -t registry.dip-dev.thehip.app/chorus-stage1 ..
docker push registry.dip-dev.thehip.app/chorus-stage1
#!/bin/bash

set -e

docker build -f dockerfiles/dev.dockerfile -t local/chorus-cicd-chorus-stage1 ..
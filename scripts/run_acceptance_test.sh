#!/bin/bash

set -e

export TEST_CONFIG_FILE="./../../../configs/ci/main.yaml" 

go test -p 1 --tags acceptance ./tests/acceptance/...
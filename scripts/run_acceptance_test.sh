#!/bin/bash

set -e

export TEST_CONFIG_FILE="./../../../configs/ci/values.yaml"

go test -p 1 --tags acceptance ./tests/acceptance/...
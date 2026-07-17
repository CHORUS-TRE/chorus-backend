#!/bin/bash

set -e

export TEST_CONFIG_FILE="./../../../configs/ci_local/backend-config.yaml"

if [ -n "$1" ]; then
    go test -count=1 -p 1 --tags acceptance "./tests/acceptance/$1" -args --ginkgo.junit-report=junit.xml
else
    go test -count=1 -p 1 --tags acceptance ./tests/acceptance/... -args --ginkgo.junit-report=junit.xml
fi

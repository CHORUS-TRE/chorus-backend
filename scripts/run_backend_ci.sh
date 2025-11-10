#!/bin/bash

set -e

yq '.config' ./configs/ci/values.yaml > ./configs/ci/backend-config.yaml
go run ./cmd/chorus/main.go --config ./configs/ci/backend-config.yaml start | go run ./cmd/logger/main.go
rm ./configs/ci/backend-config.yaml

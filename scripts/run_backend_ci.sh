#!/bin/bash

set -e

go run ./cmd/chorus/main.go --config ./configs/ci/main.yaml start | go run ./cmd/logger/main.go
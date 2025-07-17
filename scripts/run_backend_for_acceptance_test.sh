#!/bin/bash

set -e

go run ./cmd/chorus/main.go --config ./configs/ci/main.yml start | go run ./cmd/logger/main.go
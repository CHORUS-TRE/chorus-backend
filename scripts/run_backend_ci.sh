#!/bin/bash

set -e

go run ./cmd/chorus/main.go --config ./configs/ci_local/backend-config.yaml start | go run ./cmd/logger/main.go

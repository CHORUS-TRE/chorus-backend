#!/bin/bash

set -e

cd "$(dirname "$0")/.."

echo "Running integration tests..."
go test -tags integration -p 1 ./...

#!/bin/bash

set -e

if [ -n "$1" ]; then
    go test -count=1 -p 1 --tags unit "./pkg/$1/service" -cover
else
    go test -count=1 -p 1 --tags unit ./... -cover
fi

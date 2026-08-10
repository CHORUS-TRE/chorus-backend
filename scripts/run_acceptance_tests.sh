#!/bin/bash

# Acceptance suites against a self-managed backend: starts it, waits for it to
# be ready, runs the suite, then stops it with SIGINT (a graceful exit is
# required to flush coverage counters when --coverage is used).
#
# Usage: ./scripts/run_acceptance_tests.sh [--coverage] [suite]
# Env:   COVERAGE_DIR (default tests/coverage),
#        CONFIG_FILE (default configs/config.yaml), ACCEPTANCE_CONFIG_SET (acceptance overrides)

set -euo pipefail
cd "$(dirname "$0")/.."

COVERAGE=false
if [ "${1:-}" = "--coverage" ]; then
    COVERAGE=true
    shift
fi

COVERAGE_DIR="${COVERAGE_DIR:-tests/coverage}"
CONFIG_FILE="${CONFIG_FILE:-configs/config.yaml}"
ACCEPTANCE_CONFIG_SET="${ACCEPTANCE_CONFIG_SET:-storage.datastores.chorus.database=chorus_ci,storage.file_stores.disk.disk_config.base_path=docker/.diskfilestoreci,clients.oci.enabled=false,clients.kubernetes.enabled=false}"

# go test runs each suite with its package directory as the working directory
# (tests/acceptance/<suite>), not the repo root, so the path handed to it must
# be absolute.
CONFIG_FILE="$(cd "$(dirname "$CONFIG_FILE")" && pwd)/$(basename "$CONFIG_FILE")"

overrides=()
IFS=',' read -ra kvs <<< "$ACCEPTANCE_CONFIG_SET"
for kv in "${kvs[@]}"; do overrides+=(--set "$kv"); done

# Read daemon.http.port from the resolved config (file + overrides) rather
# than assuming a fixed port, so this follows whatever CONFIG_FILE sets.
PORT=$(go run ./cmd/chorus/main.go export-current-config --config "$CONFIG_FILE" "${overrides[@]}" 2>/dev/null | awk '
    /^daemon:/ { d=1; next }
    d && !/^  / { d=0 }
    d && /^  http:/ { h=1; next }
    d && h && /^  [a-zA-Z]/ && !/^  http:/ { h=0 }
    d && h && /^    port:/ {
        val=$0
        sub(/^    port: */, "", val)
        gsub(/"/, "", val)
        print val
        exit
    }
')
PORT="${PORT:-5000}"

mkdir -p "$COVERAGE_DIR"

if $COVERAGE; then
    RAW_DIR="$COVERAGE_DIR/acceptance"
    rm -rf "$RAW_DIR"
    mkdir -p "$RAW_DIR"
    BIN=bin/chorus-cov
    go build -cover -coverpkg=./... -o "$BIN" ./cmd/chorus
    GOCOVERDIR="$RAW_DIR" "$BIN" --config "$CONFIG_FILE" "${overrides[@]}" start > "$COVERAGE_DIR/backend.log" 2>&1 &
else
    BIN=bin/chorus-acceptance
    go build -o "$BIN" ./cmd/chorus
    "$BIN" --config "$CONFIG_FILE" "${overrides[@]}" start > "$COVERAGE_DIR/backend.log" 2>&1 &
fi
BACKEND_PID=$!

# Check that the backend is listening on the expected port,
# and that it is the process we just started
for _ in $(seq 1 60); do
    [ "$(lsof -ti "tcp:$PORT" 2>/dev/null)" = "$BACKEND_PID" ] && break
    kill -0 "$BACKEND_PID" 2>/dev/null || break
    sleep 1
done
[ "$(lsof -ti "tcp:$PORT" 2>/dev/null)" = "$BACKEND_PID" ] || { echo "error: backend did not start, see $COVERAGE_DIR/backend.log" >&2; exit 1; }

TARGET="./tests/acceptance/..."
[ -n "${1:-}" ] && TARGET="./tests/acceptance/$1"

status=0
TEST_CONFIG_FILE="$CONFIG_FILE" TEST_CONFIG_SET="$ACCEPTANCE_CONFIG_SET" go test -count=1 -p 1 --tags acceptance "$TARGET" -args --ginkgo.junit-report=junit.xml || status=$?

# Stop the backend we started; SIGINT lets it flush coverage counters when instrumented.
kill -INT "$BACKEND_PID"
wait "$BACKEND_PID" || true

rm "$BIN"

if $COVERAGE; then
    go tool covdata textfmt -i="$RAW_DIR" -o "$COVERAGE_DIR/acceptance.out"
    go tool cover -func="$COVERAGE_DIR/acceptance.out" | tail -1
    echo "details: make coverage-html REPORT=acceptance"
fi

exit "$status"

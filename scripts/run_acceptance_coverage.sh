#!/bin/bash

# Acceptance suites against a coverage-instrumented backend.
#
# Fails if a backend is already running. Otherwise starts its own (ci_local
# config, k8s client disabled, DB host patched in) and stops it with SIGINT
# at the end -- a graceful exit is required to flush the coverage counters.
#
# Usage: ./scripts/run_acceptance_coverage.sh [suite]
# Env:   DB_HOST (default 127.0.0.1), COVERAGE_DIR (default tests/coverage)

set -euo pipefail
cd "$(dirname "$0")/.."

DB_HOST="${DB_HOST:-127.0.0.1}"
COVERAGE_DIR="${COVERAGE_DIR:-tests/coverage}"
RAW_DIR="$COVERAGE_DIR/acceptance"
CONFIG="$(mktemp -d)/backend-config.yaml"

if nc -z localhost 5000 2>/dev/null; then
    echo "error: port 5000 is already in use, stop the running backend first" >&2
    exit 1
fi

sed "s|^      host: \"\" #.*|      host: \"$DB_HOST\"|" configs/ci_local/backend-config.yaml > "$CONFIG"

rm -rf "$RAW_DIR"
mkdir -p "$RAW_DIR"

go build -cover -coverpkg=./... -o bin/chorus-cov ./cmd/chorus

GOCOVERDIR="$RAW_DIR" bin/chorus-cov --config "$CONFIG" start > "$COVERAGE_DIR/backend.log" 2>&1 &
BACKEND_PID=$!

for _ in $(seq 1 60); do
    nc -z localhost 5000 2>/dev/null && break
    sleep 1
done
nc -z localhost 5000 2>/dev/null || { echo "error: backend did not start, see $COVERAGE_DIR/backend.log" >&2; exit 1; }

TARGET="./tests/acceptance/..."
[ -n "${1:-}" ] && TARGET="./tests/acceptance/$1"

status=0
TEST_CONFIG_FILE="$CONFIG" go test -count=1 -p 1 --tags acceptance "$TARGET" -args --ginkgo.junit-report=junit.xml || status=$?

# Stop the backend we started so it flushes its coverage counters.
kill -INT "$BACKEND_PID"
wait "$BACKEND_PID" || true

# Delete binary
rm bin/chorus-cov

go tool covdata textfmt -i="$RAW_DIR" -o "$COVERAGE_DIR/acceptance.out"
go tool cover -func="$COVERAGE_DIR/acceptance.out" | tail -1
echo "details: make coverage-html REPORT=acceptance"

exit "$status"

#!/usr/bin/env bash
# Run isolated PostgreSQL feature tests and clean only this test session's containers.
# Cached-image example:
# GOSUMDB=sum.golang.org SUB2API_TEST_POSTGRES_IMAGE=postgres:18-alpine \
# SUB2API_TEST_REDIS_IMAGE=redis:7-alpine TESTCONTAINERS_RYUK_DISABLED=true \
# bash backend/scripts/test-usage-web-search-events.sh
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
if (( $# > 1 )); then
  echo "Usage: $0 [test-name-regexp]" >&2
  exit 2
fi
pattern=${1:-^TestUsageWebSearchEvents}

docker info >/dev/null 2>&1 || {
  echo "Docker must be running; no existing database will be used." >&2
  exit 1
}

log=$(mktemp "${TMPDIR:-/tmp}/sub2api-web-search-tests.XXXXXX")
cleanup() {
  local status=$? session id containers
  session=$(awk '/Test SessionID:/ { sub(/^.*Test SessionID: /, ""); gsub(/\r/, ""); print; exit }' "$log")
  if [[ "$session" =~ ^[0-9a-f]{64}$ ]]; then
    # The existing TestMain uses os.Exit; its deferred Terminate calls do not run.
    # This also safely supports opt-in runs without the external Ryuk image.
    if containers=$(docker ps -aq --filter "label=org.testcontainers.sessionId=$session"); then
      for id in $containers; do
        if ! docker rm -fv "$id" >/dev/null; then
          echo "Failed to clean test container $id (session $session)." >&2
          status=1
        fi
      done
    else
      echo "Cannot inspect containers for test session $session; cleanup is unverified." >&2
      status=1
    fi
  fi
  printf 'Integration log: %s\n' "$log"
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

# CI=1 prevents the harness from treating an unavailable Docker daemon as a pass.
CI=1 go test -race -tags integration ./internal/repository \
  -run "$pattern" -count=1 -timeout=120s -v 2>&1 | tee "$log"

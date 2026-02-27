#!/usr/bin/env bash
# Run MES (Manufacturing Execution System). Requires Fleet API (e.g. ./run-all.sh) on port 8080.
set -e
cd "$(dirname "$0")"

if [[ -x "$HOME/go-sdk-1.22.4/bin/go" ]]; then
  export PATH="$HOME/go-sdk-1.22.4/bin:$PATH"
fi

# Optional: ensure Fleet is reachable
if [[ -z "$FLEET_API_URL" ]]; then
  export FLEET_API_URL="${FLEET_API_URL:-http://localhost:8080}"
fi

if [[ ! -x ./bin/mes ]]; then
  echo "Building MES..."
  export GOPATH="${GOPATH:-$(pwd)/.gopath}"
  go build -o bin/mes ./cmd/mes
fi

echo "Starting MES on http://localhost:8081 (Fleet: $FLEET_API_URL)"
exec ./bin/mes

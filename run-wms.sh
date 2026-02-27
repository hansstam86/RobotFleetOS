#!/usr/bin/env bash
# Run WMS (Warehouse Management System). Requires Fleet API (e.g. ./run-all.sh) on port 8080.
set -e
cd "$(dirname "$0")"

if [[ -x "$HOME/go-sdk-1.22.4/bin/go" ]]; then
  export PATH="$HOME/go-sdk-1.22.4/bin:$PATH"
fi

if [[ -z "$FLEET_API_URL" ]]; then
  export FLEET_API_URL="${FLEET_API_URL:-http://localhost:8080}"
fi

if [[ ! -x ./bin/wms ]]; then
  echo "Building WMS..."
  export GOPATH="${GOPATH:-$(pwd)/.gopath}"
  go build -o bin/wms ./cmd/wms
fi

echo "Starting WMS on http://localhost:8082 (Fleet: $FLEET_API_URL)"
exec ./bin/wms

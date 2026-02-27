#!/usr/bin/env bash
# Run Traceability service (serial/lot genealogy, recall).
set -e
cd "$(dirname "$0")"

if [[ -x "$HOME/go-sdk-1.22.4/bin/go" ]]; then
  export PATH="$HOME/go-sdk-1.22.4/bin:$PATH"
fi

if [[ ! -x ./bin/traceability ]]; then
  echo "Building Traceability..."
  export GOPATH="${GOPATH:-$(pwd)/.gopath}"
  go build -o bin/traceability ./cmd/traceability
fi

echo "Starting Traceability on http://localhost:8083"
exec ./bin/traceability

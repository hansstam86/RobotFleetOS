#!/usr/bin/env bash
# Run CMMS (Computerized Maintenance Management System).
set -e
cd "$(dirname "$0")"

if [[ -x "$HOME/go-sdk-1.22.4/bin/go" ]]; then
  export PATH="$HOME/go-sdk-1.22.4/bin:$PATH"
fi

if [[ ! -x ./bin/cmms ]]; then
  echo "Building CMMS..."
  export GOPATH="${GOPATH:-$(pwd)/.gopath}"
  go build -o bin/cmms ./cmd/cmms
fi

echo "Starting CMMS on http://localhost:8085"
exec ./bin/cmms

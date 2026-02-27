#!/usr/bin/env bash
# Run the edge layer for one robot (needs shared bus with zone for commands; use ./run-all.sh for local dev).
set -e
cd "$(dirname "$0")"

if [[ -x "$HOME/go-sdk-1.22.4/bin/go" ]]; then
  export PATH="$HOME/go-sdk-1.22.4/bin:$PATH"
fi

if [[ ! -x ./bin/edge ]]; then
  echo "Building edge..."
  export GOPATH="${GOPATH:-$(pwd)/.gopath}"
  go build -o bin/edge ./cmd/edge
fi

echo "Starting edge gateway (use ./run-all.sh to run full stack with shared bus)"
exec ./bin/edge

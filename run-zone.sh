#!/usr/bin/env bash
# Run the zone layer (needs shared bus with area for zone tasks; use ./run-all.sh for local dev).
set -e
cd "$(dirname "$0")"

if [[ -x "$HOME/go-sdk-1.22.4/bin/go" ]]; then
  export PATH="$HOME/go-sdk-1.22.4/bin:$PATH"
fi

if [[ ! -x ./bin/zone ]]; then
  echo "Building zone..."
  export GOPATH="${GOPATH:-$(pwd)/.gopath}"
  go build -o bin/zone ./cmd/zone
fi

echo "Starting zone (use ./run-all.sh to run fleet+area+zone together with shared bus)"
exec ./bin/zone

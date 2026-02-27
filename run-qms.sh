#!/usr/bin/env bash
# Run QMS (Quality Management System).
set -e
cd "$(dirname "$0")"

if [[ -x "$HOME/go-sdk-1.22.4/bin/go" ]]; then
  export PATH="$HOME/go-sdk-1.22.4/bin:$PATH"
fi

if [[ ! -x ./bin/qms ]]; then
  echo "Building QMS..."
  export GOPATH="${GOPATH:-$(pwd)/.gopath}"
  go build -o bin/qms ./cmd/qms
fi

echo "Starting QMS on http://localhost:8084"
exec ./bin/qms

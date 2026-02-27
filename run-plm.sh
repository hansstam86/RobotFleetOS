#!/usr/bin/env bash
# Run PLM (Product Lifecycle Management).
set -e
cd "$(dirname "$0")"

if [[ -x "$HOME/go-sdk-1.22.4/bin/go" ]]; then
  export PATH="$HOME/go-sdk-1.22.4/bin:$PATH"
fi

if [[ ! -x ./bin/plm ]]; then
  echo "Building PLM..."
  export GOPATH="${GOPATH:-$(pwd)/.gopath}"
  go build -o bin/plm ./cmd/plm
fi

echo "Starting PLM on http://localhost:8086"
exec ./bin/plm

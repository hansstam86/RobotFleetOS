#!/usr/bin/env bash
# Run ERP (order source, submits to MES).
set -e
cd "$(dirname "$0")"

if [[ -x "$HOME/go-sdk-1.22.4/bin/go" ]]; then
  export PATH="$HOME/go-sdk-1.22.4/bin:$PATH"
fi

if [[ ! -x ./bin/erp ]]; then
  echo "Building ERP..."
  export GOPATH="${GOPATH:-$(pwd)/.gopath}"
  go build -o bin/erp ./cmd/erp
fi

echo "Starting ERP on http://localhost:8087"
exec ./bin/erp

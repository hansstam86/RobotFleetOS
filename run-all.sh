#!/usr/bin/env bash
# Run fleet + area + zone + edge in one process (shared in-memory bus). Use for local dev.
set -e
cd "$(dirname "$0")"

if [[ -x "$HOME/go-sdk-1.22.4/bin/go" ]]; then
  export PATH="$HOME/go-sdk-1.22.4/bin:$PATH"
fi

# Check if port 8080 is already in use (e.g. old fleet/run-all still running)
if command -v lsof >/dev/null 2>&1; then
  PID=$(lsof -ti :8080 2>/dev/null || true)
  if [[ -n "$PID" ]]; then
    echo "Port 8080 is already in use (PID: $PID). Stop the old server first:"
    echo "  kill $PID"
    echo "Or to force-kill: kill -9 $PID"
    echo ""
    echo "Then run ./run-all.sh again to start the new one (with latest dashboard)."
    exit 1
  fi
fi

if [[ ! -x ./bin/all ]]; then
  echo "Building all-in-one..."
  export GOPATH="${GOPATH:-$(pwd)/.gopath}"
  go build -o bin/all ./cmd/all
fi

echo "Starting fleet + area + zone + edge (API + dashboard on http://localhost:8080)"
echo "Submit work orders and use 'Simulate firmware update' from the UI."
exec ./bin/all

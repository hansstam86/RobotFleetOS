#!/usr/bin/env bash
# Run the fleet layer (builds if needed).
set -e
cd "$(dirname "$0")"

# Use project Go if present
if [[ -x "$HOME/go-sdk-1.22.4/bin/go" ]]; then
  export PATH="$HOME/go-sdk-1.22.4/bin:$PATH"
fi

# Check if port 8080 is already in use
if command -v lsof >/dev/null 2>&1; then
  PID=$(lsof -ti :8080 2>/dev/null || true)
  if [[ -n "$PID" ]]; then
    echo "Port 8080 is already in use (PID: $PID). Stop the old server first:"
    echo "  kill $PID"
    exit 1
  fi
fi

# Build if binary missing
if [[ ! -x ./bin/fleet ]]; then
  echo "Building fleet..."
  export GOPATH="${GOPATH:-$(pwd)/.gopath}"
  go build -o bin/fleet ./cmd/fleet
fi

echo "Starting fleet layer (API on http://localhost:8080)..."
echo "  Web UI:  http://localhost:8080/  or  http://localhost:8080/ui"
echo "  GET  /health      - liveness"
echo "  POST /work_orders - submit work order"
echo "  GET  /state      - global state"
echo ""
exec ./bin/fleet

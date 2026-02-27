#!/usr/bin/env bash
# Launch all systems: Fleet (all-in-one), MES, WMS, Traceability, QMS, CMMS.
# Press Ctrl+C to stop all.
set -e
cd "$(dirname "$0")"

if [[ -x "$HOME/go-sdk-1.22.4/bin/go" ]]; then
  export PATH="$HOME/go-sdk-1.22.4/bin:$PATH"
fi

export GOPATH="${GOPATH:-$(pwd)/.gopath}"
export FLEET_API_URL="${FLEET_API_URL:-http://localhost:8080}"

# Check ports
for port in 8080 8081 8082 8083 8084; do
  if command -v lsof >/dev/null 2>&1; then
    PID=$(lsof -ti :$port 2>/dev/null || true)
    if [[ -n "$PID" ]]; then
      echo "Port $port is already in use (PID: $PID). Stop it first: kill $PID"
      exit 1
    fi
  fi
done

# Build all binaries
echo "Building binaries (if needed)..."
[[ ! -x ./bin/all ]]          && go build -o bin/all ./cmd/all
[[ ! -x ./bin/mes ]]          && go build -o bin/mes ./cmd/mes
[[ ! -x ./bin/wms ]]          && go build -o bin/wms ./cmd/wms
[[ ! -x ./bin/traceability ]] && go build -o bin/traceability ./cmd/traceability
[[ ! -x ./bin/qms ]]          && go build -o bin/qms ./cmd/qms
[[ ! -x ./bin/cmms ]]         && go build -o bin/cmms ./cmd/cmms

echo "Starting all systems..."
./bin/all &
PID_FLEET=$!
sleep 2
./bin/mes &
PID_MES=$!
./bin/wms &
PID_WMS=$!
./bin/traceability &
PID_TRACE=$!
./bin/qms &
PID_QMS=$!
./bin/cmms &
PID_CMMS=$!

cleanup() {
  echo ""
  echo "Stopping all systems..."
  kill $PID_FLEET $PID_MES $PID_WMS $PID_TRACE $PID_QMS $PID_CMMS 2>/dev/null || true
  wait $PID_FLEET $PID_MES $PID_WMS $PID_TRACE $PID_QMS $PID_CMMS 2>/dev/null || true
  exit 0
}
trap cleanup INT TERM

echo ""
echo "All systems running. Press Ctrl+C to stop all."
echo ""
echo "  Fleet:             http://localhost:8080  (dashboard, work orders)"
echo "  Fleet maintenance: http://localhost:8080/maintenance  (firmware, maintenance)"
echo "  MES:               http://localhost:8081  (production orders)"
echo "  WMS:               http://localhost:8082  (warehouse, inventory, tasks)"
echo "  Traceability:      http://localhost:8083  (serial/lot genealogy, recall)"
echo "  QMS:               http://localhost:8084  (inspections, NCRs, holds)"
echo "  CMMS:              http://localhost:8085  (equipment, maintenance work orders)"
echo ""

wait

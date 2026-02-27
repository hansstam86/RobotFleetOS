#!/usr/bin/env bash
# Test the fleet layer API (run this while ./run-fleet.sh or ./bin/fleet is running).
set -e
BASE="${1:-http://localhost:8080}"

echo "=== Testing Fleet API at $BASE ==="
echo ""

echo "1. GET /health"
curl -s "$BASE/health" | head -c 200
echo -e "\n"

echo "2. POST /work_orders (submit a work order)"
RESP=$(curl -s -X POST "$BASE/work_orders" \
  -H "Content-Type: application/json" \
  -d '{"area_id":"area-1","priority":1,"payload":[]}')
echo "$RESP" | head -c 300
echo -e "\n"

echo "3. POST /work_orders (another with deadline)"
curl -s -X POST "$BASE/work_orders" \
  -H "Content-Type: application/json" \
  -d '{"area_id":"area-2","priority":2,"payload":[1,2,3],"deadline":"2025-12-31T23:59:59Z"}' | head -c 300
echo -e "\n"

echo "4. GET /state"
curl -s "$BASE/state" | head -c 500
echo -e "\n"

echo "5. GET /state/areas"
curl -s "$BASE/state/areas" | head -c 300
echo -e "\n"

echo "6. POST /firmware/simulate (sends firmware campaign; use run-all.sh to see zone/edge handle it)"
curl -s -X POST "$BASE/firmware/simulate" | head -c 400
echo -e "\n"

echo "=== Done ==="

#!/usr/bin/env bash
# Quick test of MES API. Start Fleet (./run-all.sh) and MES (./run-mes.sh) first.
set -e
MES="${MES_URL:-http://localhost:8081}"

echo "=== MES health ==="
curl -s "$MES/health" | head -1

echo ""
echo "=== Create order ==="
curl -s -X POST "$MES/orders" \
  -H "Content-Type: application/json" \
  -d '{"sku":"WIDGET-001","quantity":50,"area_id":"area-1","priority":1}'

echo ""
echo "=== List orders ==="
curl -s "$MES/orders"

echo ""
echo "=== Release first order (replace ord-1 with actual id from create response) ==="
ID=$(curl -s "$MES/orders" | grep -o '"id":"ord-[^"]*"' | head -1 | cut -d'"' -f4)
if [[ -n "$ID" ]]; then
  curl -s -X POST "$MES/orders/$ID/release"
else
  echo "No order id found; run create first"
fi

echo ""
echo "=== List orders (after release) ==="
curl -s "$MES/orders"

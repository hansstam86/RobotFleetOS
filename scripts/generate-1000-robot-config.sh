#!/usr/bin/env bash
# Generates config files for deploying 1000 robots: 1 area, 5 zones (200 robots each), 1000 edge configs.
# Usage: ./scripts/generate-1000-robot-config.sh [NATS_URL]
# Default NATS_URL: nats://localhost:4222

set -e
NATS_URL="${1:-nats://localhost:4222}"
ROBOTS=1000
ZONES=5
ROBOTS_PER_ZONE=$((ROBOTS / ZONES))  # 200

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$REPO_ROOT/deploy/1000"
mkdir -p "$OUT/edge"

echo "Generating configs for $ROBOTS robots ($ZONES zones, $ROBOTS_PER_ZONE per zone)"
echo "NATS URL: $NATS_URL"
echo "Output: $OUT"

# Fleet
cat > "$OUT/fleet.yaml" << EOF
fleet:
  scheduler_workers: 10
  api_listen: ":8080"
messaging:
  broker: "$NATS_URL"
state:
  endpoints: ["memory"]
EOF
echo "  fleet.yaml"

# Area
echo "area:" > "$OUT/area.yaml"
echo "  area_id: area-1" >> "$OUT/area.yaml"
echo "  zones:" >> "$OUT/area.yaml"
for z in $(seq 1 $ZONES); do
  echo "    - zone-$z" >> "$OUT/area.yaml"
done
echo "messaging:" >> "$OUT/area.yaml"
echo "  broker: \"$NATS_URL\"" >> "$OUT/area.yaml"
echo "  area.yaml"

# Zones
for z in $(seq 1 $ZONES); do
  first=$(( (z - 1) * ROBOTS_PER_ZONE + 1 ))
  last=$(( z * ROBOTS_PER_ZONE ))
  {
    echo "zone:"
    echo "  zone_id: zone-$z"
    echo "  area_id: area-1"
    echo "  robots:"
    for r in $(seq $first $last); do
      echo "    - robot-$r"
    done
    echo "messaging:"
    echo "  broker: \"$NATS_URL\""
  } > "$OUT/zone-$z.yaml"
  echo "  zone-$z.yaml"
done

# Edge configs (one per robot)
for r in $(seq 1 $ROBOTS); do
  zone_num=$(( (r - 1) / ROBOTS_PER_ZONE + 1 ))
  zone_id="zone-$zone_num"
  cat > "$OUT/edge/robot-$r.yaml" << EOF
edge:
  robot_id: robot-$r
  zone_id: $zone_id
  robot_protocol: stub
messaging:
  broker: "$NATS_URL"
EOF
done
echo "  edge/robot-1.yaml ... edge/robot-$ROBOTS.yaml"

# CSV for K8s or scripting (robot_id,zone_id)
echo "robot_id,zone_id" > "$OUT/robots.csv"
for r in $(seq 1 $ROBOTS); do
  zone_num=$(( (r - 1) / ROBOTS_PER_ZONE + 1 ))
  echo "robot-$r,zone-$zone_num" >> "$OUT/robots.csv"
done
echo "  robots.csv"

echo "Done. Start NATS, then run fleet, area, 5 zones, and 1000 edges with these configs."

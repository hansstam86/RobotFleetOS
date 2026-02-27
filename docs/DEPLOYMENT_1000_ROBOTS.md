# Deploying 1000 Robots

This guide describes how to deploy RobotFleetOS for **1000 robots** using a shared **NATS** message broker so fleet, area, zones, and edge processes run separately and scale.

## Topology for 1000 Robots

| Layer   | Count | Role |
|---------|--------|------|
| **Fleet** | 1   | API, scheduler, global state |
| **Area**  | 1   | area-1, dispatches to 5 zones |
| **Zone**  | 5   | zone-1 … zone-5, 200 robots each |
| **Edge**  | 1000| One process per robot |
| **NATS**  | 1   | Message broker (all components connect to it) |

- **1 area** owns **5 zones** (zone-1 … zone-5).
- Each **zone** owns **200 robots** (e.g. zone-1: robot-1 … robot-200, zone-2: robot-201 … robot-400, etc.).
- Each **edge** process is configured with one `robot_id` and one `zone_id` (and optionally a config file).

## Prerequisites

- **NATS** server running and reachable (e.g. `nats://nats:4222` or `nats://localhost:4222`).
- Binaries built: `fleet`, `area`, `zone`, `edge` (or Docker images).

## 1. Start NATS

**Docker:**
```bash
docker run -d --name nats -p 4222:4222 nats:latest
```

**Local binary:**  
Download from [nats.io](https://nats.io) and run:
```bash
nats-server
```

## 2. Generate Configs for 1000 Robots

From the repo root:

```bash
./scripts/generate-1000-robot-config.sh
```

This creates:

- `deploy/1000/area.yaml` – area-1 with zones zone-1 … zone-5.
- `deploy/1000/zone-1.yaml` … `deploy/1000/zone-5.yaml` – each zone with 200 robots.
- `deploy/1000/edge/robot-1.yaml` … `deploy/1000/edge/robot-1000.yaml` – each edge with `robot_id` and `zone_id`.
- `deploy/1000/robots.csv` – `robot_id,zone_id` for scripting/Kubernetes.

All configs set `messaging.broker` to `nats://localhost:4222` (override with `MESSAGING_URL` if needed).

## 3. Run the Stack

Set the broker URL (if not in config):

```bash
export MESSAGING_URL=nats://localhost:4222
```

**Fleet (1 process):**
```bash
FLEET_CONFIG=deploy/1000/fleet.yaml ./bin/fleet
# or override broker only:
MESSAGING_URL=nats://localhost:4222 ./bin/fleet
```

**Area (1 process):**
```bash
AREA_CONFIG=deploy/1000/area.yaml MESSAGING_URL=nats://localhost:4222 ./bin/area
```

**Zones (5 processes):**
```bash
ZONE_CONFIG=deploy/1000/zone-1.yaml MESSAGING_URL=nats://localhost:4222 ./bin/zone &
ZONE_CONFIG=deploy/1000/zone-2.yaml MESSAGING_URL=nats://localhost:4222 ./bin/zone &
# ... zone-3, zone-4, zone-5
```

**Edges (1000 processes):**  
Each edge needs its own `robot_id` and `zone_id`. Use generated configs or env vars.

Option A – config file per robot:
```bash
for i in $(seq 1 1000); do
  EDGE_CONFIG=deploy/1000/edge/robot-$i.yaml MESSAGING_URL=nats://localhost:4222 ./bin/edge &
done
```

Option B – env vars (e.g. from `robots.csv` or Kubernetes):
```bash
EDGE_ROBOT_ID=robot-1 EDGE_ZONE_ID=zone-1 MESSAGING_URL=nats://localhost:4222 ./bin/edge
```

## 4. Kubernetes (optional)

- Run **NATS** as a Deployment + Service (or use a managed NATS).
- Run **1** Fleet Deployment, **1** Area Deployment, **5** Zone Deployments.
- Run **1000** Edge Deployments (or one Deployment with 1000 replicas and different `EDGE_ROBOT_ID` / `EDGE_ZONE_ID` per pod via a generated ConfigMap or downward API).

Example edge Deployment snippet (one pod per robot):

```yaml
env:
  - name: MESSAGING_URL
    value: "nats://nats:4222"
  - name: EDGE_ROBOT_ID
    valueFrom:
      fieldRef:
        fieldPath: metadata.name
  - name: EDGE_ZONE_ID
    valueFrom:
      configMapKeyRef:
        name: robot-zone-map
        key: $(EDGE_ROBOT_ID)
```

You’d populate `robot-zone-map` from `deploy/1000/robots.csv` (robot_id -> zone_id).

## 5. Verify

1. Open the fleet dashboard: `http://<fleet-host>:8080/`.
2. Submit a work order with `area_id: "area-1"`.
3. Check logs: area should dispatch to a zone, zone to a robot, edge should log the command and publish status.
4. Dashboard **Fleet state** / **Areas** should show area-1 with 5 zones and 1000 robots (once zone/edge reports are flowing).

## Scaling Notes

- **More robots:** Add more zones (and zone configs) and/or more robots per zone; add edge processes and configs accordingly.
- **More areas:** Add area processes and configs; point fleet work orders at the right `area_id`.
- **NATS:** Single NATS server is fine for 1k robots; for 10k+ consider a NATS cluster.

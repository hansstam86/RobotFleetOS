# RobotFleetOS

Operating system for controlling **up to 1,000,000 robots** in a single factory.

## Architecture

Control is **hierarchical** to achieve scale and resilience:

| Layer | Role | Scale (example) |
|-------|------|------------------|
| **Fleet** | Global scheduling, API, analytics | 3–10 nodes (HA) |
| **Area** | Per factory region | 10–50 areas, 20k–100k robots each |
| **Zone** | Per physical zone (aisle, cell) | 500–5k zones, 200–2k robots each |
| **Edge** | Per robot or small cell | 10k–1M nodes |

- **Commands**: Fleet → Area → Zone → Edge → Robot (no single broadcast to 1M).
- **Telemetry**: Robot → Edge → Zone → Area → Fleet (aggregated at each step).
- **Safety**: Real-time and e-stop stay at the edge; fleet is not in the critical path.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) and [docs/COMPONENTS.md](docs/COMPONENTS.md) for full design.

## Repository layout

```
docs/         Architecture and component docs
pkg/          Shared libraries (api, messaging, state, telemetry)
cmd/          Executables: fleet, area, zone, edge
configs/      Default and example configs
```

## Build and run

```bash
go mod tidy
go build -o bin/fleet ./cmd/fleet
go build -o bin/area  ./cmd/area
go build -o bin/zone  ./cmd/zone
go build -o bin/edge  ./cmd/edge
```

### Fleet layer

The fleet layer runs the global scheduler and HTTP API.

```bash
./run-fleet.sh
# or: ./bin/fleet
```

- **Web UI**: **http://localhost:8080/** — dashboard (health, state, submit work orders).
- **API**: `GET /health`, `POST /work_orders`, `GET /state`, `GET /state/areas`.

### Area layer

The area layer subscribes to work orders (for its `area_id`), creates zone tasks, and reports area summaries back to the fleet. It must share a message bus with the fleet to receive work orders.

**Recommended for local dev:** run fleet, area, zone, and an edge stub together (shared in-memory bus):

```bash
./run-all.sh
```

This starts **fleet + area + zone + edge stub** in one process. Open http://localhost:8080/, submit a work order with `area_id: "area-1"`. The area dispatches to the zone, the zone dispatches robot commands, and the edge stub publishes robot status so the zone summary (and thus area/fleet state) updates. The dashboard **Fleet state** and **Areas** will show area-1 with zone and robot counts within a few seconds.

To run layers as separate processes you need a shared broker (e.g. NATS/Kafka); the in-memory bus is per-process only.

```bash
# Optional: AREA_CONFIG=configs/default.yaml ./bin/area
./bin/area
```

Config: `area_id`, `zones` (list of zone IDs to dispatch to). See `configs/default.yaml`.

### Zone layer

The zone layer subscribes to zone tasks (for its `zone_id`), publishes robot commands to the edge, and aggregates robot status into a zone summary for the area.

```bash
./run-zone.sh
# or: ./bin/zone
```

For local dev with the in-memory bus, use `./run-all.sh` so fleet, area, zone, and edge share one bus. Config: `zone_id`, `area_id`, `robots` (list of robot IDs in this zone). See `configs/default.yaml`.

### Edge layer

The edge layer runs one process per robot (or small cell): it subscribes to robot commands for that robot, executes them (stub or real protocol), and publishes robot status to the zone.

```bash
./run-edge.sh
# or: ./bin/edge
```

For local dev use `./run-all.sh` so the full stack (fleet + area + zone + one edge per robot) shares one bus. Config: `robot_id`, `zone_id`, `robot_protocol` (`stub` | `opcua` | etc.). Stub protocol: on each TASK command the robot goes BUSY for 2s then back to IDLE. See `configs/default.yaml`.

### Build all binaries

```bash
go build -o bin/fleet ./cmd/fleet
go build -o bin/area  ./cmd/area
go build -o bin/zone  ./cmd/zone
go build -o bin/edge  ./cmd/edge
go build -o bin/all   ./cmd/all   # full stack in one process
```

Each binary can be configured via config file and/or env.

### Deploying 1000 robots

To run 1000 robots with **separate processes** (fleet, area, zones, edges) you need a **shared message broker**. Use **NATS** and the generated configs:

1. **Start NATS:** `docker run -d -p 4222:4222 nats:latest`
2. **Generate configs:** `./scripts/generate-1000-robot-config.sh` → creates `deploy/1000/` (1 area, 5 zones of 200 robots, 1000 edge configs).
3. **Run the stack:** 1 fleet, 1 area, 5 zone processes, 1000 edge processes, all with `MESSAGING_URL=nats://localhost:4222` (or use the generated YAML which sets `messaging.broker`).

See **[docs/DEPLOYMENT_1000_ROBOTS.md](docs/DEPLOYMENT_1000_ROBOTS.md)** for the full guide and Kubernetes notes.

### Docker: one container per layer and per node

Each layer and each node runs as its own container. Lower layers (zone, edge) keep operating if upper layers (fleet, area) are down; they only depend on NATS.

```bash
docker compose up -d
```

Then open **http://localhost:8080**. To test resilience: stop fleet and area (`docker compose stop fleet area`); zones and edges keep running and will resume feeding the dashboard when fleet/area are started again.

- **Compose stack**: NATS + 1 fleet + 1 area + 2 zones + 3 edges (see `docker-compose.yml`).
- **Config**: `deploy/docker/` is mounted into containers; add more zone/edge services and configs to scale.
- See **[docs/RESILIENCE_AND_LAYER_INDEPENDENCE.md](docs/RESILIENCE_AND_LAYER_INDEPENDENCE.md)** for the independence and resilience model.

### Local simulation and testing

You can **simulate and test** the full stack (work orders + firmware flow) on your local machine:

1. **Run the stack:** `./run-all.sh` (fleet + area + zone + edges in one process, in-memory bus).
2. **Open the dashboard:** http://localhost:8080 — submit work orders and click **Simulate firmware update** to see the flow end-to-end (area → zone → edge; edge simulates download/apply and reports firmware version).
3. **See [docs/LOCAL_SIMULATION.md](docs/LOCAL_SIMULATION.md)** for step-by-step testing and curl examples.

For multi-container simulation with Docker: `docker compose up -d`, then use the same dashboard and API.

### Firmware updates (up to 1M robots, heterogeneous)

Design for **resilient, staged firmware updates** over the platform, and a **firmware model** for 1M robots that are **not all the same** (multiple models, versions, compatibility):

- **[docs/FIRMWARE_UPDATE_DESIGN.md](docs/FIRMWARE_UPDATE_DESIGN.md)** — Update flow (fleet → area → zone → edge), staged rollout, health gates, rollback; firmware catalog and targeting.
- **[docs/FIRMWARE_CATALOG_EXAMPLE.md](docs/FIRMWARE_CATALOG_EXAMPLE.md)** — Example model mix and catalog for 1M heterogeneous robots.
- **`pkg/api/firmware.go`** — API types: `FirmwareUpdatePayload`, `FirmwareCampaign`, `RobotStatus.Extra` keys for model_id / firmware_version / firmware_update_status.

## License

Proprietary / TBD.

# Layer Independence and Resilience

RobotFleetOS is designed so that **each layer and each node runs in its own process (or container)** and **lower layers continue to operate when upper layers are down**. The only shared dependency is the **message bus (NATS)**.

## One container per node

- **Fleet**: one or more fleet containers (e.g. 1 for small deployments, 3+ for HA).
- **Area**: one container per area (e.g. area-1, area-2).
- **Zone**: one container per zone (e.g. zone-1, zone-2, …).
- **Edge**: one container per robot (e.g. edge for robot-1, robot-2, …).

There is no HTTP or RPC between layers; all communication is via the message bus. So:

- Fleet does not need area/zone/edge to be up to run.
- Area does not need fleet to be up to run (it just won’t receive new work orders).
- Zone does not need area or fleet to be up to run (it just won’t receive new zone tasks).
- Edge does not need zone, area, or fleet to be up to run (it just won’t receive new commands).

## What happens when an upper layer goes down

| Upper layer down | Lower layers keep doing |
|------------------|---------------------------|
| **Fleet**        | Area, zone, edge keep running. Area stops receiving new work orders. Zones still get tasks already dispatched; edges still get commands and publish status. |
| **Area**         | Zone, edge keep running. Zones stop receiving new zone tasks. Zones keep aggregating robot status and publishing zone summaries (NATS buffers or drops; when area is back, it sees new summaries). Edges keep publishing status. |
| **Zone**         | Edge keeps running. Edges stop receiving new robot commands. Edges keep publishing status so when the zone is back it can aggregate again. |
| **NATS**         | No process can talk to others until NATS is back. All components retry connecting to NATS (see below). |

So:

- **Zones and edges do not fail or block** when fleet or area are down; they only stop receiving new work from above.
- **Edges do not fail or block** when zone is down; they only stop receiving new commands and keep publishing status.
- **Safety and local operation** stay at the edge; no upper layer is in the critical path for robot control.

## Dependency rule

- **Every component depends only on NATS** (and its own config).
- **No component** calls fleet/area/zone/edge over HTTP or expects another layer to be up at startup.
- **Startup**: NATS is the only “required” service; components retry connecting to NATS so they can start before or after NATS is ready (e.g. in Docker Compose).

## Deployment

- Run **one server/container per node** (one per fleet instance, per area, per zone, per edge).
- Use **Docker Compose** or **Kubernetes** with one service/pod per node; see `docker-compose.yml` and `docs/DEPLOYMENT_1000_ROBOTS.md`.
- For **redundancy**: run multiple fleet (or area) replicas if you need HA at that layer; lower layers are unchanged and do not care which fleet/area instance is up.

## Summary

- **Modular**: One process/container per layer and per node.
- **Resilient**: Lower layers (zone, edge) keep operating when upper layers (fleet, area) or the zone are down; they only depend on NATS.
- **No hard dependency** of lower layers on upper layers; communication is message-based only.

# RobotFleetOS — Software Architecture

## Overview

RobotFleetOS is an operating system for controlling **up to 1,000,000 robots** in a single factory. The architecture is designed for:

- **Scale**: Hierarchical control to avoid single points of failure and bandwidth bottlenecks
- **Latency**: Safety-critical control at the edge; coordination and analytics in the cloud/on-prem core
- **Resilience**: Zone isolation, graceful degradation, and no single point of failure
- **Operability**: Observability, health checks, and clear boundaries between layers

---

## Design Principles

1. **No single brain** — No one process or node commands all 1M robots; control is partitioned by zone and tier.
2. **Edge-first safety** — Real-time safety and motion control run on or near the robot; the fleet OS does not sit in the critical path of emergency stop.
3. **Event-driven** — Commands and telemetry flow as events/messages; we avoid polling 1M endpoints.
4. **Bounded consistency** — Strong consistency only where required (e.g., safety); eventual consistency for fleet-wide state and analytics.
5. **Horizontal scaling** — Every tier can scale out by adding nodes (zones, area controllers, message brokers).
6. **Layer independence** — Each layer and each node runs in its own process/container. Lower layers (zone, edge) do not depend on upper layers (fleet, area) being up; they only require the message bus (NATS). See [RESILIENCE_AND_LAYER_INDEPENDENCE.md](RESILIENCE_AND_LAYER_INDEPENDENCE.md).

For **factory areas** (PCBA, molding, warehouse, assembly, etc.) and **enterprise systems** (WMS, MES, PLM, ERP, QMS, CMMS, and more), see [FACTORY_STACK_AND_SYSTEMS.md](FACTORY_STACK_AND_SYSTEMS.md).

---

## Layered Architecture

```
                    ┌─────────────────────────────────────────────────────────┐
                    │                    FLEET LAYER                           │
                    │  Scheduling • Global Optimization • Analytics • UI/API   │
                    └───────────────────────────┬─────────────────────────────┘
                                                │
                    ┌───────────────────────────▼─────────────────────────────┐
                    │                    AREA LAYER                           │
                    │  Area Controllers (per 10k–100k robots)                 │
                    │  Work orders • Zone coordination • Aggregated telemetry  │
                    └───────────────────────────┬─────────────────────────────┘
                                                │
                    ┌───────────────────────────▼─────────────────────────────┐
                    │                    ZONE LAYER                           │
                    │  Zone Controllers (per 100–2k robots per zone)           │
                    │  Task dispatch • Local state • Health & heartbeat       │
                    └───────────────────────────┬─────────────────────────────┘
                                                │
                    ┌───────────────────────────▼─────────────────────────────┐
                    │                    EDGE LAYER                           │
                    │  Robot Controllers / Gateways (per robot or small cell) │
                    │  Real-time control • Safety • Local sensor fusion       │
                    └───────────────────────────┬─────────────────────────────┘
                                                │
                    ┌───────────────────────────▼─────────────────────────────┐
                    │                    ROBOTS (1M)                          │
                    └─────────────────────────────────────────────────────────┘
```

---

## Layer Responsibilities

### 1. Fleet Layer (Central)

- **Role**: Global scheduling, optimization, analytics, and operator interfaces.
- **Scale**: One logical fleet brain, implemented as a **distributed system** (multiple replicas for HA).
- **Does NOT**: Send real-time motion commands to individual robots; that stays in Zone/Edge.
- **Key components**:
  - **Fleet Scheduler**: Assigns work orders to areas/zones; handles priorities and global constraints.
  - **Global State Store**: Eventually consistent view of fleet (which robots/zones are busy, failed, etc.).
  - **Analytics & Telemetry Pipeline**: Ingest aggregated metrics and events; dashboards, ML, reporting.
  - **API / UI**: External and internal clients (MES, WMS, operators).

### 2. Area Layer

- **Role**: One area controller per **factory region** (e.g., 10k–100k robots). Bridges Fleet and Zones.
- **Responsibilities**:
  - Receive work orders from Fleet; decompose into zone-level tasks.
  - Coordinate multiple zones (e.g., handoffs, shared resources).
  - Aggregate zone telemetry and forward to Fleet.
  - Maintain area-level state (which zones are healthy, overloaded, etc.).

### 3. Zone Layer

- **Role**: One zone controller per **physical zone** (e.g., warehouse aisle, assembly cell). Typically 100–2,000 robots per zone.
- **Responsibilities**:
  - Task dispatch to specific robots (or edge gateways).
  - Local heartbeat and health; report status to Area.
  - Bounded local state (robot IDs, current task, battery/status).
  - Optional: local conflict resolution (e.g., traffic, charging queues).

### 4. Edge Layer

- **Role**: Directly connected to robots (one process per robot, or one gateway per small cell of robots).
- **Responsibilities**:
  - **Real-time control**: Motion, safety interlocks, emergency stop (low-latency, local).
  - **Protocol translation**: Fleet OS ↔ robot-native protocols (e.g., OPC-UA, vendor APIs).
  - **Local sensor fusion**: Combine onboard sensors for local decisions.
  - **Telemetry**: Stream status/events to Zone (batched or on change).

### 5. Robots

- Physical assets. Communication is always via Edge (and optionally direct safety wiring that bypasses software where required).

---

## Communication & Data Flow

### Command Flow (Fleet → Robot)

1. **Fleet** publishes **work orders** or **high-level goals** to a message bus (e.g., Kafka, NATS, or Pulsar).
2. **Area** subscribes to orders for its region; **Zone** subscribes to tasks for its zone.
3. **Zone** sends **tasks** to **Edge** (RPC or durable queue); **Edge** translates to robot commands.
4. No single broadcast to 1M robots; fan-out is tree-shaped (Fleet → Areas → Zones → Edges → Robots).

### Telemetry Flow (Robot → Fleet)

1. **Edge** collects robot state and events; batches and forwards to **Zone**.
2. **Zone** aggregates (e.g., health, utilization) and forwards to **Area**.
3. **Area** aggregates and forwards to **Fleet** (metrics, events, alerts).
4. High-volume raw telemetry can be sampled or sent to a separate pipeline (e.g., time-series DB) to avoid overloading the control path.

### Message Bus

- **Recommended**: Distributed log / pub-sub (e.g., **Kafka** or **NATS JetStream**) for:
  - Work orders and task streams (durable, replayable).
  - Telemetry and events (scalable consumers).
- **Alternative**: Event mesh (e.g., **Redis Streams**, **RabbitMQ**) if operational constraints favor them.
- **Critical**: Topic partitioning by **area** and **zone** so that each Area/Zone only subscribes to its partition.

---

## State Management

| Scope       | Where it lives              | Consistency   | Example                          |
|------------|-----------------------------|--------------|-----------------------------------|
| Robot state| Edge + optional Zone cache  | Strong local | Position, task, safety status     |
| Zone state | Zone controller             | Strong       | Robot list, task assignments      |
| Area state | Area controller             | Strong       | Zone health, work queue           |
| Fleet state| Distributed store (e.g. CRDTs, or DB) | Eventually consistent | Fleet map, schedules, analytics |

- **Fleet-level state**: Use a distributed store (e.g., **etcd**, **Consul**, or **distributed DB**) with eventual consistency for global view; or event-sourced state from the message bus.
- **No single global lock** across 1M robots; all coordination is via messages and partitioned state.

---

## Scalability Numbers (Target)

| Layer   | Count (example)     | Robots per node | Notes                    |
|---------|---------------------|------------------|--------------------------|
| Fleet   | 3–10 (HA cluster)   | N/A              | Stateless + shared store |
| Area    | 10–50               | 20k–100k         | One per factory region   |
| Zone    | 500–5,000           | 200–2,000        | One per physical zone    |
| Edge    | 10k–1M              | 1–10             | One per robot or cell    |

- **Message throughput**: Zone → Edge and Edge → Zone dominate. Design for **millions of messages per second** in aggregate; partitioning and batching are essential.
- **Telemetry**: Pre-aggregate at Edge and Zone to avoid 1M individual streams to the center.

---

## Fault Tolerance & Safety

1. **Robot/Edge failure**: Zone marks robot unavailable; reassigns tasks; no global impact.
2. **Zone failure**: Area reassigns zone’s tasks to adjacent zones or holds; robots in that zone can pause (safety) until Zone recovers or is replaced.
3. **Area failure**: Fleet reassigns area’s work to other areas if possible; or queues work until Area recovers.
4. **Fleet failure**: Areas and Zones continue with last-known work; no new global schedules until Fleet is back.
5. **Safety**: Emergency stop and safety-critical logic must be **local** (robot or Edge). Fleet/Area/Zone must never be in the critical path for e-stop.

---

## Technology Suggestions

- **Runtime**: Go or Rust for Zone/Area/Edge (throughput, low latency); Fleet can use same or JVM/.NET for rich ecosystem.
- **Messaging**: Kafka or NATS JetStream (durable, partitioned).
- **State**: etcd/Consul for config and leader election; PostgreSQL or Cassandra for persistent fleet/area state; time-series DB for telemetry.
- **Observability**: OpenTelemetry; metrics (Prometheus); structured logs; tracing across Fleet → Area → Zone → Edge.

---

## Security & Identity

- **Authentication**: Every component (Area, Zone, Edge) has an identity (e.g., mTLS or JWT).
- **Authorization**: Role-based; Fleet can issue work; Zone can command only its Edges; Edge can command only its robots.
- **Network**: Segment by layer (Fleet/Area in a control DMZ; Zone/Edge in factory network; robots on isolated segments).

---

## Summary

RobotFleetOS scales to **1M robots** by:

1. **Hierarchical partitioning** (Fleet → Area → Zone → Edge → Robot) so no node handles 1M entities.
2. **Event-driven, message-based** command and telemetry flow with a partitioned bus.
3. **Local real-time control** at the Edge; Fleet and Area handle scheduling and optimization, not low-latency safety.
4. **Bounded state** per layer and **eventual consistency** at fleet level.
5. **Fault isolation** per zone and area, with clear failure modes and recovery.

This document is the single source of truth for the high-level architecture; ADRs and module-level docs will reference it.

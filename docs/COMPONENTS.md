# RobotFleetOS — Component Structure & Interfaces

## Repository Layout

```
RobotFleetOS/
├── docs/                    # Architecture, ADRs, runbooks
├── pkg/                     # Shared libraries (no main)
│   ├── api/                 # API types, validation, versioning
│   ├── messaging/           # Message bus abstraction (Kafka/NATS)
│   ├── state/               # State interfaces, CRDTs, stores
│   └── telemetry/           # Metrics, tracing, logging
├── cmd/                     # Executables
│   ├── fleet/               # Fleet layer (scheduler, API, global state)
│   ├── area/                # Area controller
│   ├── zone/                # Zone controller
│   └── edge/                # Edge gateway / robot controller
├── configs/                 # Default configs, schemas
├── deploy/                  # K8s, Docker, systemd (optional)
└── internal/                # Private shared code (optional)
```

---

## Core Abstractions

### 1. Identity & Addressing

- **RobotID**: globally unique (e.g. `area-zone-robot` or UUID).
- **ZoneID**, **AreaID**: hierarchical; used for routing and partitioning.
- **WorkOrderID**, **TaskID**: for tracing and idempotency.

### 2. Message Types (examples)

| Type        | Direction       | Purpose                          |
|------------|------------------|-----------------------------------|
| WorkOrder  | Fleet → Area     | High-level job (e.g. "pick 10k items from A to B") |
| ZoneTask   | Area → Zone      | Decomposed task for a zone        |
| RobotCommand | Zone → Edge    | Task or control command for robot |
| RobotStatus | Edge → Zone     | Heartbeat, state, events          |
| ZoneSummary | Zone → Area    | Aggregated health, utilization    |
| AreaSummary | Area → Fleet   | Same for area                    |
| FleetConfig | Fleet → *      | Config push (zones, limits, etc.) |

### 3. Interfaces (contracts)

- **FleetScheduler**: `SubmitWorkOrder(WorkOrder)`, `CancelWorkOrder(ID)`, `GetGlobalState()`
- **AreaController**: `AcceptWorkOrder(WorkOrder)`, `DispatchToZones(ZoneTask[])`, `ReportToFleet(AreaSummary)`
- **ZoneController**: `AcceptTask(ZoneTask)`, `AssignToRobot(RobotID, Task)`, `ReportToArea(ZoneSummary)`
- **EdgeGateway**: `ExecuteCommand(RobotCommand)`, `StreamStatus()`, `GetRobotState()`
- **StateStore**: `Get(key)`, `Put(key, value)`, `Watch(prefix)` (per layer: fleet/area/zone scope)

### 4. Configuration

- **Partitioning**: Mapping of RobotID/ZoneID to AreaID and ZoneID (config or discovery).
- **Limits**: Max robots per zone, max tasks per zone, timeouts, retry policies.
- **Feature flags**: Per area/zone for rollout (e.g. new scheduler, new protocol).

---

## Factory Areas & Enterprise Systems

Factory areas (PCBA, molding, CNC, warehouse, assembly, packaging, shipping, quality lab, rework) and enterprise systems (WMS, MES, PLM, ERP, QMS, CMMS, APS, traceability, andon) are described in [FACTORY_STACK_AND_SYSTEMS.md](FACTORY_STACK_AND_SYSTEMS.md). **MES** and **WMS** are the primary work-order sources for the Fleet layer; Fleet is the execution layer for robots and automation. **Traceability** records serial/lot events and supports genealogy and recall. **QMS** handles inspections, NCRs, and holds. **CMMS** manages equipment and maintenance work orders and can submit them to Fleet. See [MES.md](MES.md), [WMS.md](WMS.md), [TRACEABILITY.md](TRACEABILITY.md), [QMS.md](QMS.md), and [CMMS.md](CMMS.md) for runbooks and APIs.

---

## Deployment Units

| Component   | Deployment        | Scaling                         |
|------------|-------------------|----------------------------------|
| Fleet      | K8s Deployment    | Replicas for API/scheduler HA   |
| Area       | One per area      | Add nodes when adding areas     |
| Zone       | One per zone      | Add nodes when adding zones     |
| Edge       | One per robot/cell| Add nodes with robot count      |

Message bus and state stores are shared infrastructure; their sizing is derived from total robots and message rate.

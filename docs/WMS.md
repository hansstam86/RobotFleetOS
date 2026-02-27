# WMS (Warehouse Management System)

The WMS manages **storage locations**, **inventory** (by location and SKU), and **warehouse tasks** (pick, putaway, move). Tasks can be **released to Fleet** so the robot/AGV layer executes them.

---

## Overview

- **Locations** — Receiving, storage, staging, shipping. Seeded with RECV-01, STAGE-01, A-01-01, A-01-02 (zone area-1).
- **Inventory** — Receive at a location (location_id, sku, quantity, optional lot). List by location and/or SKU.
- **Tasks** — Pick (from), Putaway (from → to), Move (from → to). Create → Release (sends work order to Fleet) → Complete (updates inventory).

---

## Run WMS

1. Start the Fleet stack: `./run-all.sh` (http://localhost:8080)
2. Start WMS: `./run-wms.sh` (http://localhost:8082)

**Web UI**: http://localhost:8082 or http://localhost:8082/ui

Env: `FLEET_API_URL`, `WMS_LISTEN`, `WMS_CONFIG`

---

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| GET | /locations | List locations |
| POST | /locations | Create location (id, zone_id, type, name) |
| GET | /inventory | List inventory. Query: `?location_id=`, `?sku=` |
| POST | /inventory | Receive: body `location_id`, `sku`, `quantity`, `lot?` |
| GET | /tasks | List tasks. Query: `?status=`, `?type=` |
| POST | /tasks | Create task: `type` (pick/putaway/move), `sku`, `quantity`, `from_location_id`, `to_location_id` |
| GET | /tasks/:id | Get task |
| POST | /tasks/:id/release | Release to Fleet |
| POST | /tasks/:id/complete | Complete (updates inventory) |
| POST | /tasks/:id/cancel | Cancel |

---

## Fleet integration

When you **Release** a task, WMS POSTs a work order to Fleet with:

- **area_id** — From config `warehouse_area_id` (default area-1)
- **payload** — `wms_task_id`, `type` (pick/putaway/move), `from_location_id`, `to_location_id`, `sku`, `quantity`, `lot`

Fleet’s **Recent work orders** shows these (e.g. `pick WIDGET-001 × 10 (RECV-01 → A-01-01)`). Area/Zone/Edge consume the work order like any other; warehouse zones/AGVs would interpret the payload to run the task.

---

## Example flow

1. **Receive** — In WMS UI: Location RECV-01, SKU WIDGET-001, Quantity 100 → Receive.
2. **Putaway task** — Create task: Type Putaway, From RECV-01, To A-01-01, SKU WIDGET-001, Qty 50 → Create task.
3. **Release** — Click Release on the task. It appears in Fleet dashboard “Recent work orders”.
4. **Complete** — After execution (or manually): Click Complete. Inventory at RECV-01 decreases by 50, A-01-01 increases by 50.

---

## Config (configs/wms.yaml)

```yaml
wms:
  listen: ":8082"
  warehouse_area_id: "area-1"

fleet:
  api_url: "http://localhost:8080"
```

See [FACTORY_STACK_AND_SYSTEMS.md](FACTORY_STACK_AND_SYSTEMS.md) for how WMS fits with MES, Fleet, and ERP.

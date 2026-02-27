# CMMS (Computerized Maintenance Management System)

The CMMS manages **equipment** (maintainable assets: robots, zones, machines) and **maintenance work orders** (preventive, corrective, inspection). Work orders can be submitted to the Fleet so maintenance appears in the Fleet dashboard and can be executed by the robot fleet.

---

## Overview

- **Equipment** — Register assets with name, type (robot/zone/machine/other), area_id, zone_id, and status (operational, under_maintenance, out_of_service).
- **Maintenance work orders (MWOs)** — Create work orders for equipment. Types: preventive, corrective, inspection, **firmware_upgrade**. For firmware upgrades, set **target_firmware_version** (e.g. 2.0.0). Status flow: open → in_progress → completed (or cancelled). Optional due date and priority (1–5).
- **Fleet integration** — "Submit to Fleet" sends the MWO to the Fleet layer (payload includes target_firmware_version for firmware upgrades). **Firmware campaign** — From CMMS you can trigger a fleet-wide firmware update via **Firmware campaign** (POST /firmware/trigger); optional seed_busy to simulate N busy robots that defer the update.

---

## Run CMMS

```bash
./run-cmms.sh
```

**Web UI**: http://localhost:8085 or http://localhost:8085/ui

Env: `CMMS_LISTEN` (default :8085), `CMMS_CONFIG`, `FLEET_API_URL` (default http://localhost:8080).

---

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| GET | /equipment | List. Query: ?status=, ?type= |
| POST | /equipment | Create. Body: name, type?, area_id?, zone_id? |
| GET | /equipment/:id | Get equipment |
| PUT | /equipment/:id | Update equipment |
| GET | /mwo | List MWOs. Query: ?status=, ?equipment_id= |
| POST | /mwo | Create MWO. Body: equipment_id, type?, target_firmware_version? (for type firmware_upgrade), priority?, due_date?, description |
| GET | /mwo/:id | Get MWO |
| POST | /firmware/trigger | Trigger firmware campaign on Fleet. Body: seed_busy? (optional) |
| POST | /mwo/:id/start | Set status to in_progress (equipment → under_maintenance) |
| POST | /mwo/:id/complete | Set status to completed (equipment → operational) |
| POST | /mwo/:id/cancel | Set status to cancelled |
| POST | /mwo/:id/submit_to_fleet | Submit to Fleet. Query: ?priority= (default 3) |

---

## Integration

- **Fleet** — CMMS submits maintenance work orders to Fleet's `/work_orders`. For firmware_upgrade MWOs, payload includes target_firmware_version. CMMS can also trigger a fleet-wide firmware campaign via POST /firmware/trigger (calls Fleet's /firmware/simulate). Fleet dashboard shows recent work orders (e.g. "firmware 2.0.0" for upgrade MWOs).
- **Fleet maintenance dashboard** — Use CMMS to create firmware upgrade MWOs, submit to Fleet, and/or use the **Firmware campaign** card in CMMS to trigger the update.

See [FACTORY_STACK_AND_SYSTEMS.md](FACTORY_STACK_AND_SYSTEMS.md) for the full system map.

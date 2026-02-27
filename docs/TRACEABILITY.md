# Traceability — Serial & Lot Genealogy

The Traceability service stores **serial and lot events** (produced, received, shipped, component_linked, inspection, rework, scrap) and supports **genealogy lookup** and **recall** reporting.

---

## Overview

- **Record events** — When a unit is produced (MES), received/shipped (WMS), or inspected/reworked, post a trace record with serial and/or lot, SKU, event type, and optional MES order ID, station, parent serial.
- **Genealogy** — Look up all events for a given **serial** or **lot** (full history).
- **Recall** — Query records by lot, SKU, or date range to identify affected units for recalls.

---

## Run Traceability

No dependency on Fleet or MES; it is a standalone service that **receives** events from other systems (or the UI/API).

```bash
./run-traceability.sh
```

**Web UI**: http://localhost:8083 or http://localhost:8083/ui

Env: `TRACEABILITY_LISTEN` (default :8083), `TRACEABILITY_CONFIG`

---

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| GET | /stats | Total record count |
| POST | /records | Record an event. Body: `serial?`, `lot?`, `event_type`, `sku`, `quantity?`, `mes_order_id?`, `wms_task_id?`, `station_id?`, `zone_id?`, `parent_serial?`, `extra?` |
| GET | /genealogy?serial= | All records for this serial |
| GET | /genealogy?lot= | All records for this lot |
| GET | /recall?lot=&sku=&from=&to= | Records matching filters (RFC3339 for from/to) |

**Event types**: `produced`, `received`, `shipped`, `component_linked`, `inspection`, `rework`, `scrap`

At least one of `serial` or `lot` is required when recording.

---

## Integration with MES / WMS

- **MES**: When a production order is completed (or per unit), call `POST /records` with `event_type: produced`, `serial` or `lot`, `sku`, `mes_order_id`, `station_id` (e.g. zone or line).
- **WMS**: When receiving inventory, record `event_type: received` with `lot` and `sku`; when shipping, `event_type: shipped`.

Future: MES and WMS could call the Traceability API automatically on completion; for now you can record from the Traceability UI or any client.

---

## Example

**Record production:**
```bash
curl -s -X POST http://localhost:8083/records \
  -H "Content-Type: application/json" \
  -d '{"serial":"SN-001","event_type":"produced","sku":"SCOOTER-001","quantity":1,"mes_order_id":"ord-1","station_id":"assembly-line-1"}'
```

**Genealogy by serial:**
```bash
curl -s "http://localhost:8083/genealogy?serial=SN-001"
```

**Recall by lot:**
```bash
curl -s "http://localhost:8083/recall?lot=LOT-2024-001"
```

See [FACTORY_STACK_AND_SYSTEMS.md](FACTORY_STACK_AND_SYSTEMS.md) for how Traceability fits with MES, WMS, and QMS.

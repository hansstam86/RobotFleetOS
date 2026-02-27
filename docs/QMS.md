# QMS (Quality Management System)

The QMS manages **inspections** (pass/fail by serial or lot), **NCRs** (non-conformance reports), and **holds** (place hold on serial/lot, release when resolved).

---

## Overview

- **Inspections** — Record an inspection: serial or lot, SKU, station, result (pass/fail), notes. Optional MES order ID.
- **NCRs** — Create a non-conformance report (serial/lot, SKU, description). Status: open → close.
- **Holds** — Place a hold on a serial or lot (reason). Release when resolved.

---

## Run QMS

```bash
./run-qms.sh
```

**Web UI**: http://localhost:8084 or http://localhost:8084/ui

Env: `QMS_LISTEN` (default :8084), `QMS_CONFIG`

---

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| POST | /inspections | Record inspection. Body: serial?, lot?, sku, station_id?, result (pass/fail), notes?, mes_order_id? |
| GET | /inspections | List. Query: ?serial=, ?lot= |
| POST | /ncr | Create NCR. Body: serial?, lot?, sku, description |
| GET | /ncr | List. Query: ?status=open\|closed |
| GET | /ncr/:id | Get NCR |
| POST | /ncr/:id/close | Close NCR |
| POST | /holds | Place hold. Body: serial?, lot?, reason |
| GET | /holds | List. Query: ?active=true (active only) |
| GET | /holds/:id | Get hold |
| POST | /holds/:id/release | Release hold |

---

## Integration

- **MES**: At quality gates, call POST /inspections when a unit is inspected; link via mes_order_id. NCR can reference the same serial/lot.
- **Traceability**: Record inspections and NCRs there too (or have QMS call Traceability when recording) for full genealogy.
- **Hold/release**: MES or WMS can check QMS for active holds before releasing or shipping.

See [FACTORY_STACK_AND_SYSTEMS.md](FACTORY_STACK_AND_SYSTEMS.md) for the full system map.

# ERP (Order source)

The ERP module is a lightweight **demand/order source**: you create orders (order ref, SKU, quantity, due date) and **submit them to MES**. MES creates a production order and you can release it to Fleet from the MES UI.

---

## Overview

- **Orders** — Order ref (e.g. PO-001), optional customer ref, SKU, quantity, optional due date. Status: draft, submitted (to MES), cancelled.
- **Submit to MES** — Sends the order to MES as a production order (with `erp_order_ref`). The MES order ID is stored so you can track it in MES.
- **Cancel** — Only draft orders can be cancelled.

---

## Run ERP

```bash
./run-erp.sh
```

**Web UI**: http://localhost:8087 or http://localhost:8087/ui

Env: `ERP_LISTEN` (default :8087), `MES_API_URL` (default http://localhost:8081), `ERP_DEFAULT_AREA_ID` (default area-1), `ERP_CONFIG`.

---

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| GET | /orders | List. Query: ?status=draft\|submitted\|cancelled |
| POST | /orders | Create. Body: order_ref, customer_ref?, sku, quantity, due_date? |
| GET | /orders/:id | Get order |
| POST | /orders/:id/submit_to_mes | Submit to MES. Body: zone_id?, priority? |
| POST | /orders/:id/cancel | Cancel (draft only) |

---

## Integration

- **MES** — ERP calls MES `POST /orders` with `erp_order_ref`, `sku`, `quantity`, `area_id` (from config), optional `zone_id` and `priority`. The created MES order can then be released to Fleet from the MES dashboard.

See [FACTORY_STACK_AND_SYSTEMS.md](FACTORY_STACK_AND_SYSTEMS.md) for the full system map.

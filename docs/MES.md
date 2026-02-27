# MES (Manufacturing Execution System)

The MES layer manages **production orders**, releases work to the **Fleet** (robot/automation execution), and tracks status, scrap, and completion.

---

## Overview

- **Create** production orders (draft) with SKU, quantity, area, priority.
- **Release** an order → MES submits a work order to the Fleet API; Fleet/Area/Zone/Edge execute it.
- **Pause**, **Complete**, **Cancel** orders in MES.
- **Report scrap** (quantity + reason) per order.
- **List/Get** orders, optionally filtered by status.

---

## Run MES

1. Start the Fleet stack (so work orders can be executed):
   ```bash
   ./run-all.sh
   ```

2. In another terminal, start MES:
   ```bash
   ./run-mes.sh
   ```

   MES listens on **http://localhost:8081** and sends work orders to **http://localhost:8080** (Fleet).

   **Web UI**: Open **http://localhost:8081** or **http://localhost:8081/ui** for the MES dashboard (create orders, list with status filter, release, pause, complete, cancel, report scrap).

   Override with env:
   - `FLEET_API_URL` — Fleet API base URL (default: http://localhost:8080)
   - `MES_LISTEN` — MES HTTP listen address (default: :8081)
   - `MES_CONFIG` — Path to YAML config file

---

## Example: 1000 scooters + firmware during production

1. **Start the stack**
   - Terminal 1: `./run-all.sh` (Fleet dashboard: http://localhost:8080)
   - Terminal 2: `./run-mes.sh` (MES dashboard: http://localhost:8081)

2. **Create and release the production order (MES)**
   - Open http://localhost:8081
   - Create production order: SKU `SCOOTER-001`, Quantity `1000`, Area ID `area-1`, Priority `1`
   - Click **Create order**, then click **Release** on the new order

3. **See it in production (Fleet dashboard)**
   - Open http://localhost:8080
   - In **Recent work orders** you should see the new work order with summary `SCOOTER-001 × 1000`

4. **Trigger firmware update during the order (MES)**
   - In MES (http://localhost:8081), scroll to **Firmware update (Fleet)**
   - Optionally set **Seed busy** (e.g. 200) to simulate robots busy with work before the campaign
   - Click **Trigger firmware update**
   - Fleet receives the firmware campaign; the zone broadcasts to all robots. Busy robots defer the update until their task completes.

5. **See how it works together**
   - Fleet dashboard: **Recent work orders** shows both the scooter order and the firmware work order (e.g. `firmware 2.0.0`)
   - MES: Your production order stays **in_progress**; you can report scrap, pause, or complete when done

---

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| POST | /orders | Create order (draft). Body: `erp_order_ref`, `sku`, `product_id`, `quantity`, `area_id`, `zone_id`, `priority`, `bom_revision`, `routing_revision` |
| GET | /orders | List orders. Query: `?status=draft` (or `released`, `in_progress`, `completed`, `paused`, `cancelled`) |
| GET | /orders/:id | Get one order |
| POST | /orders/:id/release | Release to Fleet (submits work order) |
| POST | /orders/:id/pause | Pause order |
| POST | /orders/:id/complete | Mark completed |
| POST | /orders/:id/cancel | Cancel order |
| POST | /orders/:id/scrap | Report scrap. Body: `{"quantity": 1, "reason": "defect"}` |
| POST | /firmware/trigger | Trigger firmware campaign on Fleet. Body: `{"seed_busy": 0}` (optional) |

---

## Order status flow

```
draft → released / in_progress → completed
                    ↘ paused → (can release again)
                    ↘ cancelled
```

On **release**, MES builds a JSON payload (mes_order_id, sku, quantity, area_id, zone_id, etc.) and POSTs to Fleet `/work_orders`. Fleet assigns a work order ID; MES stores it in `fleet_work_order_id` and sets status to `in_progress`.

---

## Example (curl)

**Create order:**
```bash
curl -s -X POST http://localhost:8081/orders \
  -H "Content-Type: application/json" \
  -d '{"sku":"WIDGET-001","quantity":100,"area_id":"area-1","priority":1}'
```

**List orders:**
```bash
curl -s http://localhost:8081/orders
curl -s "http://localhost:8081/orders?status=draft"
```

**Release (submit to Fleet):**
```bash
curl -s -X POST http://localhost:8081/orders/ord-1/release
```

**Report scrap:**
```bash
curl -s -X POST http://localhost:8081/orders/ord-1/scrap \
  -H "Content-Type: application/json" \
  -d '{"quantity":2,"reason":"solder defect"}'
```

**Complete:**
```bash
curl -s -X POST http://localhost:8081/orders/ord-1/complete
```

---

## Config (configs/mes.yaml)

```yaml
mes:
  listen: ":8081"

fleet:
  api_url: "http://localhost:8080"
```

---

## Integration

- **Fleet**: MES is a work-order source. Release → POST /work_orders on Fleet; Fleet publishes to the bus and Area/Zone/Edge execute.
- **ERP**: (Future) ERP pushes production orders to MES (POST /orders); MES reports completion and scrap back.
- **Traceability**: (Future) Payload sent to Fleet can include serial/lot; completion events can feed traceability.

See [FACTORY_STACK_AND_SYSTEMS.md](FACTORY_STACK_AND_SYSTEMS.md) for the full system map.

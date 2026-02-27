# Local Simulation and Testing

You can **simulate the full RobotFleetOS stack** (fleet, area, zone, edge) and **firmware updates** on your local machine with a single process and no Docker.

---

## 0. Launch all systems at once

To run **Fleet, MES, WMS, and Traceability** together:

```bash
./run-all-systems.sh
```

This builds all binaries (if needed), starts each system in the background, and prints the URLs. Press **Ctrl+C** to stop all.

| System        | URL                      |
|---------------|--------------------------|
| Fleet         | http://localhost:8080    |
| MES           | http://localhost:8081    |
| WMS           | http://localhost:8082    |
| Traceability  | http://localhost:8083    |
| QMS               | http://localhost:8084    |
| CMMS              | http://localhost:8085    |
| Fleet maintenance | http://localhost:8080/maintenance (firmware, maintenance) |

---

## 1. Run the full stack (one process)

From the repo root:

```bash
./run-all.sh
```

This starts **fleet + area + zone + one edge per robot** (from zone config) with a **shared in-memory message bus**. No NATS or Docker required.

- **Dashboard**: http://localhost:8080
- **API**: http://localhost:8080/health, /work_orders, /state, /firmware/simulate

---

## 2. Test work orders

1. Open http://localhost:8080
2. In **Submit work order**, set **Area ID** to `area-1`, **Priority** to `1`, leave **Payload** empty or use `[]`
3. Click **Submit work order**

You should see in the terminal:

- `area: work order wo-... -> zone zone-1 task ...`
- `zone: task ... -> robot robot-1` (or robot-2)
- `edge robot-1: received command ... type=TASK`
- `edge robot-1: task ... completed (stub)`

**Fleet state** and **Areas** on the dashboard will show area-1 with zones and robots (after a few seconds).

---

## 3. Test firmware simulation

1. With `./run-all.sh` still running, open http://localhost:8080
2. Optionally set **Seed busy** to a number (e.g. `200`) so that many robots receive work orders first and will **defer** the firmware update until they finish.
3. Click **Simulate firmware update**

This sends a **firmware campaign** work order. The zone **broadcasts** the update to **all robots** in the zone. Robots that are **IDLE** start the update immediately; robots that are **BUSY** (e.g. from seed work orders) **defer** the update and apply it when their task completes. In the terminal you should see:

- `area: work order ... -> zone zone-1 task ...`
- `zone: firmware task ... -> N robots (broadcast)`
- For each robot: either immediate firmware flow or `firmware update deferred until work order complete` then later `applying deferred firmware update`.

The edge **simulates** download (2s) and apply (2s), then reports **firmware_version** and **firmware_update_status** in `RobotStatus.Extra`. No real download or flash.

---

## 3b. Simulate 1000+ robots and deferred firmware

To simulate **more than a thousand robots** with some busy (deferring firmware until work complete):

1. Start the stack with a robot count (e.g. 1200):
   ```bash
   SIMULATE_ROBOTS=1200 ./run-all.sh
   ```
   This creates `robot-1` … `robot-1200` in one zone and runs a single **edge simulator** process that handles all of them (no 1200 separate processes).

2. In the dashboard, set **Seed busy** to e.g. `200` and click **Simulate firmware update**.
   - 200 work orders are submitted first → 200 robots become BUSY.
   - Then the firmware campaign is sent; the zone **broadcasts** to all 1200 robots.
   - ~1000 idle robots start the update immediately; the 200 busy ones **defer** and apply when their 2s stub task completes.

3. Or use curl with `seed_busy`:
   ```bash
   curl -s -X POST http://localhost:8080/firmware/simulate \
     -H "Content-Type: application/json" \
     -d '{"seed_busy": 200}'
   ```

---

## 4. Test with curl

**Health:**
```bash
curl -s http://localhost:8080/health
```

**Submit work order:**
```bash
curl -s -X POST http://localhost:8080/work_orders \
  -H "Content-Type: application/json" \
  -d '{"area_id":"area-1","priority":1,"payload":""}'
```

**Simulate firmware update (all robots in zone):**
```bash
curl -s -X POST http://localhost:8080/firmware/simulate
```

**Simulate with some robots busy (they defer until work complete):**
```bash
curl -s -X POST http://localhost:8080/firmware/simulate \
  -H "Content-Type: application/json" \
  -d '{"seed_busy": 200}'
```

**Fleet state:**
```bash
curl -s http://localhost:8080/state
```

---

## 5. Run the test script

From the repo root:

```bash
./scripts/test-fleet.sh
```

This hits health, work orders, and state. Start the stack with `./run-all.sh` in another terminal first.

---

## 6. Docker (multi-container simulation)

To run **each layer in its own container** with NATS:

```bash
docker compose up -d
```

Then open http://localhost:8080 and use **Submit work order** and **Simulate firmware update** as above. Logs: `docker compose logs -f zone-1 edge-1` etc.

---

## Summary

| Goal                    | Command        | URL                    |
|-------------------------|----------------|------------------------|
| Full stack (no Docker)  | `./run-all.sh` | http://localhost:8080 |
| Work order flow         | Dashboard or curl | POST /work_orders   |
| Firmware simulation     | Dashboard or curl | POST /firmware/simulate |
| API test script         | `./scripts/test-fleet.sh` | —                |

All of this runs on your local machine with no external services except the browser.

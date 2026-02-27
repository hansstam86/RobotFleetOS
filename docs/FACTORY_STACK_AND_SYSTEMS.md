# Factory Stack & Enterprise Systems

This document extends the RobotFleetOS architecture for a **consumer hardware factory**: production areas, warehouse, assembly, and the enterprise systems (WMS, MES, PLM, ERP, QMS, etc.) that integrate with the robot fleet and with each other.

---

## 1. Factory Areas (Production & Logistics)

These map to **Areas** and **Zones** in the existing Fleet/Area/Zone/Edge model. Each area contains zones (lines, cells, aisles) with robots and automation.

| Area ID | Name | Description | Typical zones / lines |
|--------|------|-------------|------------------------|
| **pcba** | PCBA / SMT Line | Printed circuit board assembly: solder paste, pick-and-place, reflow, AOI, selective solder, conformal coat | paste-1, pickplace-1..N, reflow-1, aoi-1, wash, test-ict, test-fct |
| **molding** | Injection Molding | Plastic parts: injection presses, mold changes, cooling, degating | press-1..N, mold-storage, cooling, quality-gate |
| **cnc-tooling** | CNC & Tooling | Mold production, machining, tool grinding, EDM | cnc-1..N, edm-1..N, tool-crib, inspection |
| **warehouse-raw** | Raw Materials Warehouse | Inbound raw materials, components, consumables | receiving, staging, rack-aisles-1..N, kitting, quarantine |
| **warehouse-wip** | WIP Warehouse | Work-in-progress between lines (e.g. PCBA, molded parts) | buffer-zones, AGV drop-off, handoff-to-assembly |
| **assembly** | Assembly Area | Fully automated final assembly (robots, conveyors, screw-driving, bonding) | line-1..N, kitting, fastening, bonding, test-station |
| **warehouse-fg** | Finished Goods Warehouse | Finished products ready to ship | putaway, pick-zones, packing-staging, returns |
| **packaging** | Packaging & Kitting | Boxing, labeling, palletizing, shipping prep | pack-lines-1..N, label, palletize, stretch-wrap |
| **shipping** | Shipping & Logistics | Outbound docks, loading, carrier handoff | dock-1..N, staging, loading, yard |
| **quality-lab** | Quality Lab | Incoming QC, in-process audit, failure analysis, calibration | incoming-inspection, lab-test, fa-lab, calibration |
| **rework-repair** | Rework & Repair | Rework stations, repair, refurb, scrap handling | rework-1..N, repair-cell, scrap-disposition |

Additional areas you may add:

- **coating-plating** — Surface treatment, plating, coating lines
- **staging-kitting** — Dedicated kitting for assembly (if not inside warehouse or assembly)
- **utilities** — Central cooling, compressed air, power monitoring (often integrated with CMMS/MES)

---

## 2. Enterprise Systems Overview

| System | Acronym | Purpose | Integrates with |
|--------|---------|---------|------------------|
| **Warehouse Management System** | WMS | Inventory locations, putaway, picking, cycle counts, lot/serial in warehouse | MES, ERP, Fleet (AGVs, conveyors) |
| **Manufacturing Execution System** | MES | Work orders, routing, traceability, real-time production, downtime, OEE | ERP, PLM, Fleet, WMS, QMS |
| **Product Lifecycle Management** | PLM | BOM, ECO, variants, revisions, compliance | ERP, MES, QMS |
| **Enterprise Resource Planning** | ERP | Finance, orders, demand, procurement, capacity planning | MES, WMS, PLM, QMS, suppliers |
| **Quality Management / QC Center** | QMS | Inspections, SPC, NCR, CAPA, calibration, audits | MES, PLM, lab equipment, Fleet (vision/measure) |
| **Maintenance Management** | CMMS | Preventive & corrective maintenance, spares, downtime tracking | MES, Fleet (robot/line health), sensors |
| **Advanced Planning & Scheduling** | APS | Finite capacity scheduling, what-if, prioritization | ERP, MES, Fleet scheduler |
| **Traceability & Serialization** | — | Unit/lot/serial genealogy, recall, compliance | MES, WMS, QMS, assembly/PCBA lines |
| **Supply Chain / Procurement** | — | PO, ASN, supplier quality, inbound logistics | ERP, WMS, QMS |
| **Andon / Production Monitoring** | — | Line status, alarms, escalation, dashboards | MES, Fleet, CMMS |

---

## 3. System Responsibilities & Integration Points

### 3.1 Warehouse Management System (WMS)

- **Owns**: Storage locations, bin/serial/lot inventory in warehouse, putaway rules, pick waves, cycle counts, replenishment.
- **Integrates with**:
  - **Fleet**: Publish “pick” / “putaway” / “move” work orders to the robot fleet (AGVs, conveyors, shuttle systems). Consume completion and location updates.
  - **MES**: Receive material demand for production; report material issues; confirm kitting for assembly.
  - **ERP**: Sync inventory levels, reservations, and movements for finance and procurement.

**Suggested APIs / events**: `CreatePickTask`, `CreatePutawayTask`, `InventoryAdjustment`, `LocationUpdate`; consume `WorkOrderCompleted`, `RobotStatus` (for warehouse zones).

---

### 3.2 Manufacturing Execution System (MES)

- **Owns**: Work orders (from ERP or internal), routing by area/zone, release to production, real-time execution, downtime/OEE, traceability (serial/lot at station).
- **Integrates with**:
  - **Fleet**: MES is a **primary work-order source**. MES creates high-level jobs (e.g. “assemble 1000 units of SKU X in area assembly, zone line-1”); Fleet/Area/Zone/Edge execute and report back (start, complete, fail).
  - **PLM**: Consume BOM, routing, and ECO; report as-built and deviations.
  - **WMS**: Request kitting and material consumption; WMS confirms picks and locations.
  - **QMS**: Trigger inspections, NCR, and hold at quality gates; receive results and release/hold decisions.
  - **ERP**: Report production completion, scrap, labor; consume demand and capacity.

**Suggested APIs / events**: `ReleaseWorkOrder`, `PauseWorkOrder`, `CompleteWorkOrder`, `ReportScrap`; consume `ZoneSummary`, `AreaSummary`, `WorkOrderCompleted`, `RobotStatus` (aggregated).

---

### 3.3 Product Lifecycle Management (PLM)

- **Owns**: Product structure (BOM), revisions, ECO/ECN, variants, compliance (e.g. RoHS, REACH), document control.
- **Integrates with**:
  - **ERP**: Master data (item, BOM) for planning and costing.
  - **MES**: Released BOM and routing for production; ECO effective dates and cut-in points.
  - **QMS**: Approved specs, inspection plans, and change control.

**Suggested APIs / events**: `GetBOM`, `GetRouting`, `GetEffectiveECO`, `NotifyECOEffective`; feed MES/ERP with item and BOM versions.

---

### 3.4 Enterprise Resource Planning (ERP)

- **Owns**: Demand, orders, capacity planning, procurement, finance, costing.
- **Integrates with**:
  - **MES**: Send production orders and due dates; receive completion, scrap, and labor.
  - **WMS**: Inventory and reservation sync; outbound shipment triggers.
  - **PLM**: Item and BOM master data.
  - **QMS**: Quality costs, claims, and audit data.
  - **Fleet**: Optional high-level “demand” or “campaign” signals that MES or APS translates into work orders.

**Suggested APIs / events**: `CreateProductionOrder`, `UpdateDemand`; consume `ProductionCompleted`, `ScrapReport`, `InventorySnapshot`.

---

### 3.5 Quality Management / Quality Control Center (QMS)

- **Owns**: Inspection plans, SPC, NCR, CAPA, calibration, audits, hold/release.
- **Integrates with**:
  - **MES**: Quality gates in routing; hold/release at station; NCR linked to work order/serial.
  - **PLM**: Specs and approved changes.
  - **Fleet**: In quality-lab or in-line zones: trigger “inspect” or “measure” work; receive results from vision/measure robots or lab equipment.
  - **CMMS**: Calibration and equipment qualification.

**Suggested APIs / events**: `CreateInspection`, `RecordNCR`, `HoldUnit`, `ReleaseUnit`, `SPCDataPoint`; consume `WorkOrderCompleted` (inspection type), instrument results.

---

### 3.6 Maintenance Management (CMMS)

- **Owns**: PM/CM work orders, asset registry, spares, downtime reasons.
- **Integrates with**:
  - **MES**: Downtime and reason codes; maintenance windows.
  - **Fleet**: Consume robot/line health and utilization; create maintenance work orders; optional “maintenance mode” or reduced load for a zone.

**Suggested APIs / events**: `CreateMaintenanceOrder`, `RecordDowntime`; consume `ZoneSummary`, `RobotStatus` (health, errors), alarms.

---

### 3.7 Advanced Planning & Scheduling (APS)

- **Owns**: Finite capacity scheduling, prioritization, what-if, material and resource constraints.
- **Integrates with**:
  - **ERP**: Demand and capacity.
  - **MES**: Released work orders and actual progress.
  - **Fleet**: Optional: suggested sequence or priority for work orders (Fleet scheduler consumes or uses as input).

**Suggested APIs / events**: `GetCapacity`, `ProposeSchedule`, `PublishSchedule`; consume `AreaSummary`, `ZoneSummary` for capacity and load.

---

### 3.8 Traceability & Serialization

- **Owns**: Serial/lot genealogy, component-to-assembly linkage, recall and compliance reporting.
- **Integrates with**:
  - **MES**: Record serial/lot at each step; link to work order and station.
  - **WMS**: Lot/serial at receive and ship.
  - **QMS**: Link NCR and inspections to serial/lot.
  - **Fleet**: Assembly and test zones report serial/lot with completion events.

**Suggested APIs / events**: `RegisterSerial`, `LinkComponentToAssembly`, `QueryGenealogy`; consume completion and scan events from MES/Fleet.

---

## 4. How the Fleet Layer Fits

- **Fleet** does not replace MES, WMS, or ERP. It is the **execution layer for robots and automation** in the factory.
- **MES** (and optionally APS) **create work** (e.g. “produce 500 units on assembly line-1”); **Fleet** receives work orders, decomposes them into Area → Zone → Edge → Robot, and reports status back.
- **WMS** creates warehouse **tasks** (pick, putaway, move); Fleet executes them in warehouse zones (AGVs, conveyors, shuttles).
- **QMS** can trigger **inspection work** in quality-lab or in-line zones; Fleet dispatches to the right robots or stations.
- **CMMS** consumes **health and utilization** from Fleet (Area/Zone summaries, robot status) to drive maintenance orders.

Data flow (conceptual):

```
ERP → MES (production orders)
PLM → MES (BOM, routing, ECO)
MES → Fleet (work orders for assembly, PCBA, etc.)
WMS → Fleet (work orders for warehouse)
QMS ↔ MES (hold/release, inspections); QMS → Fleet (inspection work in lab zones)
Fleet → MES / WMS / APS (completion, status, utilization)
Fleet → CMMS (health, downtime signals)
```

---

## 5. Additional Systems (Short List)

| System | Purpose |
|--------|--------|
| **Andon / Production Monitoring** | Real-time line status, alarms, escalation; consumes MES and Fleet events. |
| **Energy / Sustainability** | Metering, energy per unit, carbon; can consume zone/line data from MES/Fleet. |
| **Training & Compliance** | Operator certifications, SOPs, audit trail; can link to MES and QMS. |
| **Supplier Portal** | ASN, quality documents, forecasts; feeds WMS and QMS. |
| **Label & Print** | Serial labels, shipping labels, work order labels; integrates with MES, WMS, Traceability. |

---

## 6. Implementation Roadmap (Suggested Order)

1. **MES ↔ Fleet** — Work order flow from MES to Fleet; completion and status back to MES (foundation for all production areas).
2. **WMS ↔ Fleet** — Warehouse work orders (pick/putaway/move) and inventory-related events.
3. **Traceability** — Serial/lot at key points; genealogy and recall (build on MES and Fleet completion events).
4. **QMS** — Inspection plans, NCR, hold/release; quality gates in MES; link to Fleet for lab/in-line inspection work.
5. **PLM ↔ MES/ERP** — BOM, routing, ECO; then ERP ↔ MES for orders and completion.
6. **CMMS** — Asset and maintenance orders; consume Fleet/MES health and downtime.
7. **APS** — Scheduling and prioritization; optional integration with Fleet scheduler.

---

## 7. Summary

- **Factory areas**: PCBA, molding, CNC/tooling, warehouse (raw/WIP/finished), assembly, packaging, shipping, quality lab, rework/repair (plus optional coating, staging, utilities).
- **Core systems**: WMS, MES, PLM, ERP, QMS, CMMS, APS, Traceability, Supply Chain, Andon.
- **Fleet** is the robot/automation execution layer; **MES** and **WMS** are the main sources of work orders; **QMS**, **PLM**, **ERP**, **CMMS**, and **APS** integrate with MES, WMS, and Fleet as above.

This document is the single reference for factory areas and enterprise systems; ADRs and component-level docs should reference it for integration contracts and data flows.

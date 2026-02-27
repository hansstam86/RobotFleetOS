# Firmware Update Design: Resilient Rollout for Up to 1M Heterogeneous Robots

This document describes how to perform **firmware updates** for up to **1,000,000 robots** in a **resilient, staged manner** using the RobotFleetOS platform, and how **firmware** is modeled when robots are **not all the same** (multiple models, variants, and compatibility).

---

## 1. Goals

- **Scale**: Update up to 1M robots without overwhelming the network, update servers, or control plane.
- **Resilience**: Rollout continues safely if fleet/area/zone go down mid-update; edge/robot can complete or rollback without dependency on upper layers.
- **Heterogeneity**: Support many robot **models** and **firmware variants**; only apply compatible firmware to each robot.
- **Safety**: Staged rollout with **health gates**; pause or rollback if failure rate exceeds threshold.
- **Observability**: Fleet knows which robots are on which version; campaign progress and per-zone/area metrics.

---

## 2. Firmware Model: 1M Robots, Not All the Same

### 2.1 Robot identity and compatibility

- Each robot has a **robot_id** (already in the platform).
- Each robot has a **model_id** (e.g. `picker-v2`, `agv-x1`, `conveyor-b`) that determines which firmware images are compatible.
- Each robot reports **current_firmware_version** (e.g. `2.1.0`) and **model_id** in telemetry (e.g. in `RobotStatus.Extra` or a dedicated firmware-status message).

So: **1M robots** = many **models** × many **zones** × many **current versions**. A single “firmware update” is really: “for each (model, target_version), update the set of robots that match.”

### 2.2 Firmware catalog (registry)

- **Firmware image**: identified by **(version, model_id)**. Example: `(2.2.0, picker-v2)`.
- **Metadata per image**:  
  - **download_url**: where the edge fetches the image (CDN, artifact store). Do **not** push binary over the message bus.  
  - **checksum** (e.g. SHA-256): integrity and idempotent retries.  
  - **rollback_version**: version to revert to on failure (optional).  
  - **min_hardware_rev** / **compatibility** (optional): for sub-variants within a model.
- **Catalog** is stored at the fleet (or a dedicated firmware service). Areas/zones/edges receive only **references** (version, model_id, url, checksum) in commands.

### 2.3 Example model mix for 1M robots

| model_id    | Role      | Approx. count | Example firmware versions |
|------------|-----------|----------------|----------------------------|
| picker-v2  | Picking   | 400,000        | 2.1.0 → 2.2.0             |
| agv-x1     | Transport | 350,000        | 1.5.0 → 1.6.0             |
| conveyor-b | Conveyor  | 200,000        | 3.0.0 → 3.0.1             |
| welder-c   | Welding   | 50,000         | 1.0.0 (no update)         |

Firmware updates are **per (model_id, target_version)**; the fleet targets only robots that match that model and (optionally) current version.

---

## 3. Update Flow Using the Platform

Firmware updates reuse the **existing hierarchy** (fleet → area → zone → edge) so that:

- No single node orchestrates 1M robots; work is **partitioned by area and zone**.
- **Staged rollout** is enforced at fleet/area (e.g. by zone, by area, by cohort).
- **Rate limiting** is applied at zone or area (e.g. max N concurrent updates per zone).
- The **edge** is the only layer that talks to the robot and performs download/apply/rollback; it does **not** depend on fleet/area/zone being up during the actual flash.

### 3.1 High-level flow

1. **Fleet** creates a **firmware campaign** (target model_id, target version, rollout policy: stages, rate limits, health gates).
2. **Fleet** (or a firmware service) resolves the **firmware catalog** for (model_id, target_version) → url, checksum, rollback_version.
3. **Fleet** pushes **campaign metadata** and **target list** (or targeting rules) to areas; areas **decompose by zone** and push **firmware update tasks** (zone tasks) to zones.
4. **Zone** dispatches **firmware commands** to edges (one command per robot, or batched by zone). Each command includes: **campaign_id**, **version**, **model_id**, **url**, **checksum**, **rollback_version** (and optionally **deadline**).
5. **Edge** receives the command, **validates** (model_id match, compatibility), **downloads** from url (with checksum check), **applies** (install/reboot), and **reports** result (success / failure / rollback). Edge does **not** need fleet/area/zone to be up during download/apply.
6. **Zone** aggregates **update results** (success/fail/rollback) and reports to area; **area** aggregates and reports to **fleet**.
7. **Fleet** evaluates **health gates** (e.g. success rate &gt; 99% in current stage). If pass → proceed to next stage; if fail → **pause** campaign and optionally **trigger rollback** for failed or all-in-stage robots.

### 3.2 Message and command types

- **Firmware campaign** (fleet-internal or fleet → area):  
  `campaign_id`, `model_id`, `target_version`, `url`, `checksum`, `rollback_version`, `stages` (e.g. list of zone_ids or area_ids to update in order), `max_concurrent_per_zone`, `health_gate_success_rate`, `health_gate_min_count`.
- **Firmware zone task** (area → zone):  
  Same as above plus `zone_id` and list of **robot_ids** (or “all in zone matching model_id”). Zone rate-limits how many it sends to edges at once.
- **Robot command** (zone → edge):  
  Existing `RobotCommand` with **type** `FIRMWARE_UPDATE` and **payload** = JSON of `{ campaign_id, version, model_id, url, checksum, rollback_version }`. Edge only needs this payload to run the update; no further calls to fleet/area/zone during apply.
- **Robot status / firmware result** (edge → zone):  
  In **RobotStatus.Extra** (or a dedicated topic): `firmware_version`, `model_id`, `firmware_update_status` (idle | downloading | applying | success | failed | rollback). Zone aggregates and reports to area → fleet.

So: **firmware for 1M robots** is a set of **many (model_id, version) images** in a **catalog**; **updates** are **campaigns** that flow down as **zone tasks** and **robot commands**, with **results** flowing back up for **health gates** and **progress**.

---

## 4. Resilient Behavior

### 4.1 No upper-layer dependency during apply

- Once the **edge** has received a **FIRMWARE_UPDATE** command, it has everything needed (url, checksum, rollback_version) to **download**, **verify**, **apply**, and optionally **rollback** without talking to fleet/area/zone again.
- If **fleet/area/zone** go down mid-campaign:  
  - Edges that **already got** the command continue and report result when zone is back (or buffer and send later if we add a result buffer).  
  - Edges that **did not** get the command simply don’t start the update; no half-state.

### 4.2 Idempotent and resumable at the edge

- **Download**: Use **checksum**; if the same version is requested again, skip download or verify and continue.  
- **Apply**: Prefer **atomic or two-phase** (e.g. install to secondary partition, then switch on reboot) so that a crash during apply can be recovered on next boot (retry or rollback using stored rollback_version).
- **Reporting**: Report **success** or **failure** once; zone/area/fleet can deduplicate by (robot_id, campaign_id).

### 4.3 Staged rollout and rate limiting

- **Stages**: e.g. update **zone-1** first, then **zone-2**, … or **area-1** then **area-2**. Fleet (or area) only sends firmware zone tasks for the **current stage**.
- **Rate limit**: Zone (or area) allows at most **max_concurrent_per_zone** (e.g. 50) robots in “downloading” or “applying” at once. Prevents thundering herd and overloads the CDN and NATS.
- **Health gate**: After each stage (or window), fleet checks **success_rate** and **min_count**. If below threshold → **pause** campaign, **alert**, optionally **issue rollback commands** for the failed cohort.

### 4.4 Rollback

- **Per-robot rollback**: Edge stores **rollback_version** and **url/checksum** for it (from the update command). On **failure** or on **explicit rollback command** (type `FIRMWARE_ROLLBACK`), edge applies rollback without needing fleet.
- **Campaign-level rollback**: Fleet can emit a **rollback campaign** (same flow as update but with `target_version = rollback_version`) for the same **robot set** that received the failed stage.

---

## 5. Firmware Catalog and Targeting (Summary)

- **Catalog**: Stored at fleet (DB or config); maps (model_id, version) → url, checksum, rollback_version. Edges never need to “discover” the catalog; they receive everything in the command.
- **Targeting**: Fleet decides **which robots** get which update (e.g. “all robots with model_id=picker-v2 and current_firmware_version=2.1.0”). This uses **inventory** built from **RobotStatus.Extra** (firmware_version, model_id) reported by edges and aggregated up.
- **1M heterogeneous robots**: Many **model_id**s; each has its own **firmware versions**. Campaigns are **per (model_id, target_version)**; multiple campaigns can run in parallel for different models if desired, with separate stages and health gates.

---

## 6. Implementation Plan (Phased)

### Phase 1 – Foundation

- **RobotStatus.Extra**: Add **model_id**, **firmware_version**, **firmware_update_status** so zone/area/fleet can see current state and build inventory.
- **Firmware catalog** (fleet): API or config to register (model_id, version) → url, checksum, rollback_version.
- **RobotCommand** type **FIRMWARE_UPDATE** with payload schema (campaign_id, version, model_id, url, checksum, rollback_version).

### Phase 2 – Edge update protocol

- **Edge**: Handle **FIRMWARE_UPDATE** (validate model_id, download with checksum, apply, report). Handle **FIRMWARE_ROLLBACK** (apply stored rollback image). Optional: resume after reboot if apply was interrupted.

### Phase 3 – Campaign and rollout

- **Fleet**: **Firmware campaign** creation (target model, version, stages, rate limits, health gates). Campaign produces **firmware zone tasks** (or work orders tagged as firmware) to areas.
- **Area/Zone**: Consume firmware zone tasks; **rate limit** concurrent updates per zone; dispatch **FIRMWARE_UPDATE** commands to edges; aggregate **firmware_update_status** and report to fleet.
- **Fleet**: **Health gate** evaluation after each stage; **pause** or **rollback** decision; optional **dashboard** for campaign progress.

### Phase 4 – Scale and hardening

- **CDN / artifact store** for firmware images (scalable download for 1M robots).
- **Observability**: Metrics and logs per campaign, per zone, per model; alerts on failure rate and rollback.
- **Testing**: Canary (one zone, one model) before full rollout.

---

## 7. Summary

- **Firmware for 1M robots**: Modeled as **many (model_id, version)** images in a **catalog**; robots report **model_id** and **firmware_version** so updates are **targeted** and **compatible**.
- **Resilient updates**: Use **existing platform** (fleet → area → zone → edge); **staged rollout** and **rate limiting**; **health gates**; **rollback** at edge and campaign level.
- **Independence**: **Edge** has everything in the **command** (url, checksum, rollback); **no dependency on upper layers** during download/apply; **idempotent/resumable** where possible.

---

## 8. API types (reference)

The following types are defined in `pkg/api/firmware.go` for use across layers:

- **FirmwareUpdatePayload**: payload for `RobotCommand` type `FIRMWARE_UPDATE` (campaign_id, version, model_id, download_url, checksum_sha256, rollback_version, rollback_url). Edge uses this alone to perform the update.
- **FirmwareRollbackPayload**: payload for `RobotCommand` type `FIRMWARE_ROLLBACK`.
- **FirmwareImage**: catalog entry (model_id, version, download_url, checksum_sha256, rollback_version, rollback_url).
- **FirmwareCampaign**, **FirmwareCampaignTarget**, **FirmwareCampaignStage**: fleet-side campaign definition (targeting, stages, rate limits, health gates).

**RobotStatus.Extra** keys for firmware reporting (edge → zone → area → fleet):

- `model_id`: robot hardware model (e.g. picker-v2).
- `firmware_version`: current installed version.
- `firmware_update_status`: `idle` | `downloading` | `applying` | `success` | `failed` | `rollback`.

See **[FIRMWARE_CATALOG_EXAMPLE.md](FIRMWARE_CATALOG_EXAMPLE.md)** for an example model mix and catalog for 1M heterogeneous robots.

Next step is **Phase 1** (inventory from status, catalog API, command types) and **Phase 2** (edge handling FIRMWARE_UPDATE and FIRMWARE_ROLLBACK).

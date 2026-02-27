# Example Firmware Catalog for 1M Heterogeneous Robots

This document gives a **concrete example** of how firmware is modeled when **not all 1M robots are the same**: multiple **models**, each with their own **firmware versions** and **compatibility**.

---

## 1. Robot model mix (example)

Assume a factory with **1,000,000 robots** across several **models** and **roles**:

| model_id   | Role           | Approx. count | Hardware family |
|-----------|----------------|---------------|------------------|
| picker-v2 | Item picking   | 400,000       | Arm + gripper    |
| picker-v1 | Item picking   | 50,000        | Legacy arm       |
| agv-x1    | Transport      | 350,000       | Autonomous cart  |
| agv-x2    | Transport      | 100,000       | Newer cart       |
| conveyor-b| Conveyor       | 80,000        | Belt segments    |
| conveyor-c| Conveyor       | 15,000        | Next-gen         |
| welder-c  | Welding        | 5,000         | Fixed cell       |

**Total**: 1,000,000. Each **model_id** has its own **firmware lineage**; you never apply picker firmware to an AGV.

---

## 2. Firmware catalog (example entries)

Catalog maps **(model_id, version)** → download URL, checksum, optional rollback. Example:

| model_id   | version | download_url (example)              | checksum_sha256 | rollback_version |
|-----------|---------|--------------------------------------|-----------------|------------------|
| picker-v2 | 2.2.0   | https://cdn.example/fw/picker-v2/2.2.0.bin | a1b2c3...       | 2.1.0            |
| picker-v2 | 2.1.0   | https://cdn.example/fw/picker-v2/2.1.0.bin | d4e5f6...       | 2.0.0            |
| picker-v1 | 1.4.0   | https://cdn.example/fw/picker-v1/1.4.0.bin | ...             | 1.3.0            |
| agv-x1    | 1.6.0   | https://cdn.example/fw/agv-x1/1.6.0.bin    | ...             | 1.5.0            |
| agv-x2    | 2.0.0   | https://cdn.example/fw/agv-x2/2.0.0.bin    | ...             | 1.9.0            |
| conveyor-b| 3.0.1   | https://cdn.example/fw/conveyor-b/3.0.1.bin| ...             | 3.0.0            |
| conveyor-c| 1.0.0   | https://cdn.example/fw/conveyor-c/1.0.0.bin | ...             | —                |
| welder-c  | 1.0.0   | https://cdn.example/fw/welder-c/1.0.0.bin  | ...             | —                |

- **picker-v2** and **agv-x1** have multiple versions (upgrades and rollbacks).
- **conveyor-c** and **welder-c** might have a single version initially; catalog still stores (model_id, version) for consistency.

---

## 3. Targeting an update (example)

**Campaign**: “Update all **picker-v2** robots from **2.1.0** to **2.2.0**.”

- **Target**: model_id = `picker-v2`, target_version = `2.2.0`, optional current_version = `2.1.0`.
- **Fleet** uses **inventory** (from RobotStatus.Extra: model_id, firmware_version) to know which robots match.
- **Catalog** gives url + checksum + rollback for (picker-v2, 2.2.0).
- Rollout is **staged** (e.g. by zone); **rate limit** e.g. 50 concurrent per zone; **health gate** e.g. 99% success before advancing.

So “firmware for 1M robots” is **many (model_id, version) images** in the catalog; **each update campaign** targets **one (model_id, target_version)** and only compatible robots.

---

## 4. Scale and CDN

- **1M robots** do **not** pull firmware through the control plane; they pull from **CDN / artifact store** using the **download_url** in the command.
- Each **(model_id, version)** is one binary (or a small set); e.g. 10 models × 3 versions = 30 objects. Many robots share the same URL; the CDN serves at scale.
- **Checksum** in the command ensures integrity and allows **idempotent retries** at the edge.

This example can be turned into **config or seed data** for the fleet firmware catalog (Phase 1 implementation).

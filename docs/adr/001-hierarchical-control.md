# ADR 001: Hierarchical control topology for 1M robots

## Status

Accepted.

## Context

A single factory may contain up to 1,000,000 robots. We need a control architecture that:

- Does not require one node to command or poll 1M endpoints.
- Keeps safety-critical and real-time control at low latency (near the robot).
- Allows horizontal scaling and fault isolation by region.

## Decision

We adopt a **four-layer hierarchy**: Fleet → Area → Zone → Edge → Robot.

- **Fleet**: Global scheduling and API; stateless replicas + shared store; does not send real-time commands to robots.
- **Area**: One logical controller per factory region (e.g. 10k–100k robots); receives work orders, decomposes to zone tasks.
- **Zone**: One controller per physical zone (e.g. 100–2k robots); dispatches tasks to edge/robots; reports summaries to area.
- **Edge**: One process per robot or small cell; real-time control and protocol translation; reports status to zone.

Commands flow top-down; telemetry flows bottom-up with aggregation at each layer. Partitioning is by AreaID and ZoneID so that no single node handles more than a bounded subset of robots.

## Consequences

- **Pros**: Bounded fan-out per node; clear fault boundaries; safety stays at edge; can scale by adding areas/zones/edges.
- **Cons**: Added latency for fleet-wide decisions; eventual consistency at fleet level; operational complexity of multiple tiers.

## Alternatives considered

- **Flat mesh**: Every robot talks to a central bus. Rejected due to message volume and single-point congestion.
- **Two tiers (Fleet + Edge)**: Rejected because a single fleet layer would still need to partition work; introducing Area/Zone makes partitioning explicit and matches physical factory layout.

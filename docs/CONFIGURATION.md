# Functional configuration

`config/gati.yaml` is the complete production-default document. `internal/config`
is its authoritative typed schema and validation policy (schema version 1).
Every field is required, including explicit zero/false values. Unknown or duplicate
keys, aliases, nulls, coercion from strings to numbers, extra documents and files
above 64 KiB are rejected. No field accepts credentials or external URLs.

`config/simulation.yaml` contains low-count local test defaults. Normal builds
reject it. Only explicitly compiled `-tags simulation` binaries can load it;
HTTP serving with a simulation build is restricted to literal loopback addresses.
There are no clock-control or test routes in either scaffold build.

Configuration values for matching, arrivals, aggregates and notifications are
validated **future functional inputs**, not evidence that those features exist.
At this milestone the server exposes configuration and health only. New features
must use these values and add behavioral tests as they land.

The public representation uses deterministic JSON field order and a SHA-256 hash
of those bytes. YAML comments/formatting do not change the hash. All consumers can
inspect the effective settings and schema version. Changes require restart; live
gathering version pinning will land with gathering storage/matching.

Production floors: activation/arrival counts ≥10, public minimum ≥20, publication
interval ≥300 seconds and delay ≥1 epoch. Availability is 30..120 minutes. Retention
caps follow the plan; deadlines must leave enough time for matching stability and
delayed publication. Public alerts accept only released bucket boundaries. The
schema also bounds geography, lists and numeric values to reject overflow or
unbounded configuration. These are explicit review guardrails, not anonymity proofs.

Simulation can lower count/publication floors; all lifetime and shape validation
still applies. Changing a threshold does not change other independent thresholds.
Changing schema guardrails requires a visible code/doc review rather than an env
variable bypass. Proposed geographic settings still require validation against
actual crossing fixtures in milestone 03.

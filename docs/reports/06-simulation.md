# Milestone 06 — reproducible willingness simulation

Date: 2026-09-13. Ubuntu 24.04.3 local Docker, pinned toolchain and real imported
Tirana geography. Scope: willingness stage only, not full gathering simulation.

| Scenario, seed 42 | Synthetic people | Credentials | Accepted | Rejected | Cancelled | Retry verified | Remaining expired |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| tirana-evening | 40 | 40 | 40 | 0 | 6 | 5 | 34 |
| tirana-sybil | 40 | 238 | 59 | 179 | 0 | 1 | 59 |

The evening report was generated twice using separate fresh stores; `cmp` verified
byte-identical JSON. Random capabilities never enter the report. Rate windows use
real time, so larger/slow scenarios may produce different admission totals when
crossing windows; this is explicitly not a deterministic throughput benchmark.
The extra duplicate POST consumes one of the burst's 60 creation requests.

Validation: `make test`, `make test-store`, `make verify-local`, both simulation
scenarios, and simulation-tag real-store clock/namespace checks passed. Existing
HTTP smoke verifies absent production clock route; the built production binary
contains no `/api/simulation/clock` route string. Tests cover hostile target refusal,
redirect refusal, strict scenario settings, seeded input, invalid probabilities,
clock expiry/retry and denial of production namespace access by simulation roles.

During regression testing, two Go packages concurrently used the same synthetic
cell with different TTLs, making a store TTL assertion order dependent. The
isolated-store script now runs packages sequentially (`-p 1`); within-package
concurrency/race tests remain enabled.

No human-independence, physical-presence, aggregate-privacy or 100k-capacity claim
follows from these results. Future milestones extend these fixtures as actual
matching/arrival/publication code lands.

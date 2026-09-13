# Development progress and handoff

Updated: 2026-09-13. Status: milestones 02–07 implemented; next is arrival claims and confirmation.

## Current task and authorization

The user resumed implementation after documentation preparation. Installing needed
tools, implementing/testing milestones and making incremental local commits are
authorized. Docker access is now available; continue through the plan. No push or deployment
is authorized, and GitHub authentication/write access remains unverified.

## Actual repository state

- Product requirements, plan, agent instructions and initial threat/decision records.
- Checksum-pinned user-local bootstrap for Go 1.27.1, Node 24.21.0/npm 11.19.0 and
  Compose 5.5.1. Source `scripts/env.sh` or use Make to select installed tools.
- Typed YAML configuration, production/simulation validation, canonical JSON/hash,
  Go config/health API, Albanian client preparation screen and native HTTP smoke.
- AGPL-3.0-or-later LICENSE and contributor/security/third-party guidance.
- Pinned Dockerfiles/development Compose and CI definition; container runtime checks passed; remote CI has not run. The API connects to private Valkey with a restricted role.
- Fixed coarse Tirana grid, public map import/provenance and static crossing reachability index.
- Expiring willingness APIs, replay tombstones, bounded indexes, rate limits and restricted ephemeral Valkey roles.
- Albanian willingness UI with first-party roads/coarse selection, optional location, expiring client capability and cancellation.
- No activation engine, arrival, collective map or notification code yet.
  The configuration fields for those features describe future functional inputs.

## Milestone tracker

| Plan milestone | State | Evidence / next work |
| --- | --- | --- |
| 01 — Product/threat boundaries | Documented | Requirements, plan, decision 0001, THREAT_MODEL |
| 02 — Toolchain, scaffold, license, typed config | Complete | Native and container checks passed; reports/02-scaffold.md |
| 03 — Coarse geography and crossing fixtures | Complete | 2,571 imported crossings, 28,124 road segments; grid/selection tests and byte-identical offline rebuild |
| 04 — Expiring willingness | Complete | Restricted real-store lifecycle/TTL/replay/race/ACL tests and running Compose create/status/cancel passed |
| 05 — Albanian willingness UI | Complete | Five Chromium browser checks: actual map load, manual/GPS coarse-only payloads, reload, retry/cancel failures, expiry restoration and mobile/keyboard layout |
| 06 — Seeded Tirana simulation | Complete for willingness stage | Dedicated store/credentials/build, frozen clock, seeded API scenarios, standalone synthetic map and repeatable report |
| 07 — Continuous activation and late admission | Implemented | Worker lease/deadlines, atomic reservation/activation, existing/direct late joins, decline/cutoff tests and real-browser gathering flow |
| 08 — Arrival claims | Not started | Founding/late participants, nonces/freshness/retraction |
| 09 — Aggregate map/statistics | Not started | Canonical releases and inference review |
| 10 — Notifications and follows | Not started | Foreground, fake sink, expiring optional push and joins |
| 11 — Security and lifecycle hardening | Not started | Store/host/runtime controls and hostile-use scenarios |
| 12 — 100k benchmark | Not started | Reference hardware, real-time workload and capacity report |
| 13 — Deployment | Not started | Production stack, backup/restore/rollback and clean VM validation |
| 14 — Verifiable releases | Not started | SBOM, signed manifests and independent artifact comparison |
| 15 — Independent audit/pilot evidence | Not started | Resolve findings; no production-ready claim yet |

## Working commands

```sh
make deps
make config-check
make config-show
make config-check-simulation
make verify-local
make dev-native
```

`make verify-local` runs native tests/builds/config checks, Compose **parsing** and
an HTTP smoke. It does not start containers. `make dev-native` serves a read-only UI/API preview at http://127.0.0.1:5173; Ctrl-C stops its API/client processes. It has no Valkey connection; use Compose for willingness. Read [DEVELOPMENT.md](DEVELOPMENT.md) for setup.

`make dev` / `make down` are implemented for Compose but require Docker Engine.
`make simulate` runs the isolated willingness population simulator and controllable
clock. `make load` does not exist yet. Matching/arrival scenarios follow those features.

## Validation and commits

[Milestone 02 report](reports/02-scaffold.md) records exact validation and limitations.
Use `git log` for authoritative commit IDs. The preparation commit is `a082b7a`;
toolchain installation was committed as `b9b362e`. Subsequent scaffold work is a
separate coherent local commit. Do not embed a commit's own hash in its files.

## Next action

Implement milestone 08: shared founding/late arrival nonce, fresh coarse location
claims, retraction/expiry, stable JEMI KËTU and arrival UI. Keep reviewing cross-surface
privacy and scaling; neither is certified. The public-card client connection to
atomic direct join belongs with maps/notifications (09–10).
The user-created untracked `commands.txt` is unrelated: leave it untouched/uncommitted.

The real map extract is archived with a 2026-09-13 base timestamp/checksum. Offline
rebuild needs no network. `make map-check` and geographic race tests passed.
Keep updating this tracker with tests, failures and the next concrete action.

## Milestone 04 evidence

`make test-store` passed using a disposable Valkey (TTL, cleanup, concurrent retries,
cancel/replay, rate counters, denied privileged commands and real HTTP validation).
`make verify-local` passed race/unit/config/build/native smoke checks. Updated pinned
Compose images built and the running stack passed `make check-containers` plus a
synthetic create/status/cancel/410 check through Vite. Source secrets live only in
ignored `.runtime/`; no values were printed. API schema and retention are documented
in API.md/STORAGE.md. Configuration schema 2 adds public abuse/cleanup limits.

Setup tests found/fixed a Redis client logger-interface mismatch and repeated ACL
file write permissions. No current credential blocker remains.

## Milestone 05 evidence

Pinned MapLibre 6.9.0 and Playwright 1.63.0; installed Playwright Chromium in the
user cache. Five real-browser scenarios passed against the built production client and local Compose API.
They verify no external HTTP requests, cookies, localStorage, exact coordinate
payloads or location history; sessionStorage holds only the bounded capability,
coarse request and deadline. Actual basemap worker loading is asserted.
Inspected the generated mobile screenshot in ignored reports/local. Initial checks
found/fixed MapLibre ESM import and Vite worker bundling issues. Production browser tests
exposed the missing worker sibling import; the documented ?worker&url pipeline
now bundles the worker and its dependencies. make verify-local also passed. Browser tests
exercise synthetic sessions and cancel them; no real location was requested.

The map is currently a first-party road backdrop and the user's own selected coarse
area, not a public activity map. PWA installation/offline shell, notifications and
activation are still pending. Browser session restoration/backups are outside
forensic erasure guarantees. Server deadlines remain authoritative.

## Milestone 06 evidence

See reports/06-simulation.md and SIMULATION.md for exact results and limitations.
Two evening runs produced byte-identical JSON; a single-network Sybil burst accepted
59/238 credentials for 40 synthetic people, rejecting 179. Logical expiry and
restricted simulation namespace were checked against real disposable Valkey.
Production store integration and native checks passed after fixing cross-package
test interference. No simulation-control credential or synthetic participant token
is written to a report or committed. Normal Compose remains separate.

## Milestone 07, planner slice

`internal/matching` selects oldest compatible founders, ranks mapped crossings,
computes the nearest common crossing, enforces unaligned remaining availability,
prefers compatible open gatherings, respects declines/cutoffs and freezes the
proposed deadline. Focused race tests passed. Decision 0003 explains the bounded
cohort and conservative reservation stability. This planner is not wired into the
API yet and does not establish end-to-end activation. Next: atomic Valkey
reservation/activation/admission and worker recovery/timer integration.

## Milestone 07, transaction slice

Atomic temporary reservation and activation now exist in the store (not yet wired
to the running API worker). Real-store race tests cover overlapping reservations,
stale coarse-area/radius/deadline validation, early activation refusal, cancelled
founder replacement refusal, replay-stable destination/deadline and removal of
private bookkeeping from own-session JSON. `make test-store` passed.
Founding transactions currently support activation thresholds up to 500; validation
rejects larger thresholds and thresholds exceeding configured signal capacity.

## Milestone 07 integration evidence

The running code now schedules the transactions and serves Albanian invitations,
going and decline. The engine uses a short Valkey lease, coalesced dirty flag,
2-second timer checks and 10-second reconciliation defaults. Snapshot records are
read in bounded pages; only destination records are cached between runs. Atomic
transitions protect against stale/racing plans. Throughput/failover timing remains
unbenchmarked; pending/open-index cleanup and worst-case scheduling need hardening.

Schema 3 adds invitation cooldown and maximum temporary decline links. Config, unit,
race, native build/smoke, real-store and Compose checks passed. Six Chromium checks
passed against built assets and the real API, including collective invitation,
going, decline and both map renders. Inspected synthetic destination screenshots.
Simulation tests exercise the real Tirana index/API at 9,999ms and 10,000ms, late
admission with fixed destination/end, decline retention, cutoff rejection, atomic
new-recipient join without side effects on failure and expiry. The normal binary
still omits simulation controls. The Sybil fixture remains 59/238 accepted credentials.

Outstanding release work includes map-update closure, worker failure observability,
load/fairness/missed-match measurement, comprehensive lifecycle/privacy review and
the later arrival/aggregate/notification/deployment features. These are tracked
gates, not completed protections.

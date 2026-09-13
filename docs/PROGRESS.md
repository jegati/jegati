# Development progress and handoff

Updated: 2026-09-13. Status: milestones 02–08 implemented; nearest-crossroad/device-only correction implemented and validated. Milestone 09 can resume under the disclosed inference limitation.

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
- Fixed coarse Tirana grid, public map import/provenance and static intersection reachability index (6,491 road junctions).
- Expiring willingness APIs, replay tombstones, bounded indexes, rate limits and restricted ephemeral Valkey roles.
- Albanian willingness UI with first-party roads, required one-shot device location, expiring client capability and cancellation.
- Continuous activation, late admission, going/decline and temporary arrivals are implemented.
  Collective public maps/statistics and notification subscriptions remain future work.

## Milestone tracker

| Plan milestone | State | Evidence / next work |
| --- | --- | --- |
| 01 — Product/threat boundaries | Documented | Requirements, plan, decision 0001, THREAT_MODEL |
| 02 — Toolchain, scaffold, license, typed config | Complete | Native and container checks passed; reports/02-scaffold.md |
| 03 — Coarse geography and intersection fixtures | Complete | 6,491 imported road intersections, 28,124 road segments; grid/selection tests and byte-identical offline rebuild |
| 04 — Expiring willingness | Complete | Restricted real-store lifecycle/TTL/replay/race/ACL tests and running Compose create/status/cancel passed |
| 05 — Albanian willingness UI | Complete | 13 Chromium checks pass after device-only correction; previous five-check milestone evidence below is historical |
| 06 — Seeded Tirana simulation | Complete for willingness stage | Dedicated store/credentials/build, frozen clock, seeded API scenarios, standalone synthetic map and repeatable report |
| 07 — Continuous activation and late admission | Implemented | Worker lease/deadlines, atomic reservation/activation, existing/direct late joins, decline/cutoff tests and real-browser gathering flow |
| 08 — Arrival claims | Implemented | Shared fresh coarse claims, one-use nonce/replay tests, stable JEMI KËTU, retraction/expiry and browser arrival flow |
| 09 — Aggregate map/statistics | Known failure accepted for local work | Production-threshold collusion probe reconstructs one synthetic target cell; see reports/09-inference-gate.md |
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

The latest user request prioritizes planning Tirana population-response simulations
before implementing/running the new suite. Present docs/SIMULATION_PLAN.md and the
16–28-hour effort estimate; docs/MATCHING_PARAMETERS.md explains every current
matching field, related settings, validation choices and inactive controls.
The next implementation step is the scenario schema/event scheduler, followed by
responses/journeys and replay/reporting. Milestone 09 remains pending behind this
requested simulation work. No new response-population run has been claimed.
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

`internal/matching` selects oldest compatible founders, ranks mapped intersections,
computes the nearest common intersection, enforces unaligned remaining availability,
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

## Milestone 08 evidence

Arrival nonce, claim and retraction APIs are implemented for both founding and late
participants. Nonces are random client-generated authorization header credentials;
only hashes are stored, with at most 120-second lifetime. Claims transmit only a
fresh coarse cell and retain no arrival coordinate/cell history. Replays return
the same confirmation without extending freshness or increasing credential count.
JEMI KËTU requires a stable cohort; expiry/retraction/cancel/switch remove current
contributions, and a drop below threshold resets confirmation immediately.

Real Valkey race/replay tests, fixed-clock API tests, native/config/build checks and
Compose runtime checks passed. Seven Chromium checks passed, including wrong/lost
response recovery for arrival and cancellation during an in-flight creation.
Simulation tests cover founding plus late arrival, wrong-area rejection, 9,999ms/
10,000ms presence boundaries, retraction/replay, joining after JEMI KËTU and stale
confirmation removal on read before the scheduled cleanup.

Schema 4 adds the per-signal arrival request budget. Development Valkey was recreated
to mount the updated restrictive ACL; this deliberately resets ephemeral local
sessions. GPS was mocked for all tests. Distinct credentials and claimed coarse
location do not establish independent humans or proof of physical presence.

## Milestone 09 original privacy gate — historical decision request

`make privacy-probe` reproduced a counterexample against the actual API with the
production activation threshold of 20 and normal rate limits. Nineteen controlled
credentials from one network plus one unknown synthetic signal produced a intersection
that narrowed 127 compatible coarse cells to one. No database or public aggregate
endpoint was needed. The command deliberately exited 2: this is a FAILED privacy
gate, not a passing assurance test. Synthetic evidence and assumptions are committed
in reports/09-inference-gate.md and 09-inference-counterexample.json.

No public map/statistics/follow surfaces were shipped after finding the issue.
Changing the deterministic closest-intersection rule or accepting weaker protection
against colluding inputs is a material product decision. The default unit/race/
store/browser checks remain separate from this deliberately failing counterexample.
No credentials, remote push, public deployment or real location were used.

## Crossroad and required device-location correction

Decision 0004 records the user's correction: derive road intersections, require
fresh device-service location with no manual fallback, and retain nearest selection
with the known collusion limitation. Schema 5 changes map/config/destination names
explicitly and adds local fix-quality settings (100m maximum reported error, 60s
maximum age, configurable). Old manual-input client sessions are discarded.

The original archived OSM response rebuilds 6,491 street junctions and the unchanged
28,124 road segments offline. Shared-node connectivity, three distinct road legs,
access filters and no geometry-only crossing inference are tested. Road map tags
and browser coordinates cannot prove pedestrian suitability or physical presence.

Validation so far: make map-check, make verify-local and make test-store passed.
The isolated 40-person scenario and continuous activation/arrival tests passed.
make privacy-probe deliberately exited 2: with the new intersection dataset, 19
controlled credentials again narrow 127 possible target cells to one. New evidence
is in reports/09-crossroad-inference-counterexample.json; original evidence retained.
Rebuilt/recreated local Compose and make check-containers passed. All 13 Chromium
checks passed against the built client and real API, including permission denial,
missing location service, stale/poor/out-of-area fixes, selection expiry, no manual
map input, keyboard/mobile, retry/cancellation and the complete arrival flow.
Inspected the synthetic mobile screenshot. No real device location was requested.
The disposable local store was recreated for the incompatible destination schema.

## Population-response simulation planning

Inspected the current generator, API driver, worker/planner, validation, configs and
existing reports. Added a six-scenario/three-seed proposal with synthetic behavior,
production-equivalent thresholds in an isolated profile, real-window rate-limit
handling, mocked-device browser journeys and measurements. Estimate: 16–28 focused
developer hours; first response-enabled run after roughly 8–13 hours of that work.
No new simulation or app tests were run for this documentation-only change.

Documented all matching fields and related choices. Notable implementation limits:
intersection_index_batch_size is unused; the current runner hardcodes the small
simulation profile (3 willing/2 arrivals), generates all traffic from one peer and
has narrower availability/radius choices than the application schema. Previous
population reports remain willingness-only; fixed integration tests are separate.
Corrected obsolete SIMULATION.md wording that implied activation/arrival app code
had not landed. Validation: source/config cross-check and git diff --check.

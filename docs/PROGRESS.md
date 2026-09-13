# Development progress and handoff

Updated: 2026-09-13. Status: milestones 02–08 implemented; nearest-crossroad/device-only correction implemented and validated. The 3,000-person response simulation suite is complete. Milestone 09 can resume under the disclosed inference limitation.

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
| 06 — Seeded Tirana simulation | Complete; response suite added | Dedicated store/build, controllable clock, 18 checked city cases, 3,000-person behavior/arrival replay; reports/population-3000.md |
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

Implement commit 2 of ACTIVITY_IMPLEMENTATION_PLAN.md: bounded store capture,
immutable delayed publication and the public activity endpoint, then connect cards
and the cell-only city map. Decision 0007 accepts documented inference limitations;
no new inference-approval gate is required. Push remains explicitly opt-in/off by
default. Preserve private 100 m cells / 50 m device-error / 30/20 thresholds.
The schema-7 public policy is implemented and tested; no aggregate API/UI exists yet.
No credential or toolchain blocker remains for local work.
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

## 3,000-person simulation implementation — historical starting point

User increased the main population to 3,000 and authorized current settings:
activation 30, arrivals 20, radius choices 0.1/0.5/1/3 km, plus support for 100 m
cells. Current grid remains the user's 1,000 m setting; short radii cannot reach
any destination under the conservative whole-cell rule. Decision 0005 records
this distinction, finer-grid exposure, and a possible explicit comparison.
Schema 6 adds fractional radii/100 m grid support. Regression config fixtures are
now separate from the user's editable operating config. At that starting point, simulation actor/event work was ongoing and no completed
3,000-person result existed; final evidence is recorded below.


## Population-response implementation and initial validation

Added bounded event scheduling, configurable weighted inputs, real TCP peer groups,
write pacing, actual API going/decline/late admission/nonce/arrival/retraction, and
normalization of report-only gathering aliases. One-shot browser location remains
mandatory in the product. Simulation worker timing is driven only by the authenticated
virtual clock; normal builds retain real background timing.

Initial exploratory 3,000-person baseline accepted all 3,000, observed 23 gatherings
and 15 JEMI KËTU events, with 639 arrival claims in 203.5 real seconds. This was
before the final journey-position/unique-actor reporting refinement; retain it as
exploration, not the final experiment result. Radius 0.1/0.5 km actors received no
invitations with 1,000 m cells. The 80-person guaranteed-success scenario passed
all funnel/config/expiry checks with all 80 arriving and a confirmed gathering.

Unit/race tests, a bounded exhaustive planner oracle, real-store tests, legacy
simulation integration, and all 13 current-config browser checks passed. The
standalone replay browser test passed on the exploratory 3,000-person artifact.
The final 18-run suite and 100 m comparison are next. No real GPS or deployment.

## Population suite findings and limiter correction

The first 18-run batch completed. Four checks initially failed: three sparse runs
correctly had no gathering but exposed a null-slice report-checker bug; one dense
run stopped on HTTP 503 exactly at the 16:20:00 UTC network-secret rollover.
The limiter previously used separate SET NX and GET calls, allowing expiration
between them. Secret selection/read now occurs atomically using Valkey's real
clock, retaining the existing ten-minute bounded lifecycle. Real-store race tests
passed, including a one-millisecond-before-rollover fixture, old-secret expiration
and concurrent replicas sharing the next epoch's secret. No raw address logging.
The failed run remains preserved; rerun and final suite evidence are pending.

The baseline seed-42 repeat has identical complete functional JSON/timelines after
excluding only wall runtime, sampled driver heap and source build metadata.
Canonical functional SHA-256:
`f9141abd66bc40039ba48131ddab20ac98404b28da460182312a69295c69c13a`.

Report tooling now accepts empty gathering results, compares complete functional
reports, respects a scenario's seed unless explicitly overridden, and archives
failed runs on suite resume. Added a shared-network diagnostic and CI success/replay
checks. Updated the parameter reference and original plan to schema 6/current
settings. `make verify-local`, the real-store rollover checks and the 80-person
success control passed. Final resumed suite and post-fix browser validation follow.


## Final population simulation evidence

All 18 scenario checks now pass (six cases × seeds 42/43/44). Initial failures remain
preserved: three empty-result checker failures, and the dense seed-42 rollover 503.
The latter was rerun after the atomic limiter fix; completed cases were rechecked.
Baseline seed 42: 3,000 accepted, 1,107 invited, 489 unique actors confirming arrival,
746 confirmation events, 27 observed gatherings, 15 ever JEMI KËTU. All accepted
sessions cancelled/expired, zero baseline rate rejections. Fine-grid comparison:
1,891 invited, 1,067 unique arrival actors, 41 gatherings, 30 ever JEMI KËTU.
The 1,000 m baseline excludes all 0.1/0.5 km-radius matching; it remains unchanged.

The initial batch took 929.22 seconds on this 16-thread Ryzen host. Before/after-fix
baseline repeats have identical complete functional reports. The 80-person control
has all 80 arriving; a YAML seed-43 control confirms seed selection. One-network
120-person control accepted 55 with expected rate rejection, all 55 arriving.

Final validation: make verify-local (race/unit/simulation-tag/vet/TS/build/config/
normal-route smoke), make test-store including rollover/ACL/expiry, legacy real-
store simulation integration, rebuilt Compose and all 13 app browser tests passed.
Standalone replay checks passed for the control and final 3,000-person artifact;
make check-containers passed after the final Compose rebuild.
The report checker rejects corrupted expiry evidence; duplicate suite seeds are
rejected before touching output. CI includes the control/replay checks but remote
CI has not run. No downloaded tools or real participant records were committed.

Reviewed evidence: reports/population-3000.md and population-3000-summary.json;
interactive artifacts remain in ignored reports/local. Remaining limitations:
synthetic decisions/journeys, injected late discovery, no human usability or 100k
capacity claim, finer-cell exposure, known collusion/location-spoofing risks, and
unfinished public maps/notifications/deployment verification. Work is committed
locally; unrelated commands.txt remains untouched. Next action is milestone 09.

## Default 100 m grid and functionality audit

The user selected 100 m (0.1 km) cells as the default after the population comparison.
Updated gati.yaml and the matching population example to 100 m / 50 m maximum
reported device error (decision 0006); thresholds/radii remain unchanged. Restored
supported schema 6 after validation found a stale schema-5 declaration. Fixed the
browser founding fixture to derive its cell from the running grid rather than
hardcoding a 1 km cell. Historical experiment snapshots and regression grids remain
explicitly unchanged. The development stack was restarted with a fresh ephemeral
store to avoid interpreting old grid state.

Runtime production-profile SHA-256:
`32e4166672c16df72d5a34538396c37937ddc5fd9f95241fafecfb9357f0c43b`.
Equivalent isolated population profile SHA-256:
`8ec7ec11fead062be0566d5d3243acfc32977d2c716982f4e15b5f3083b57335`.

Validation passed: config checks, make verify-local, make test-store, legacy
simulation/fixed-clock integration, and all 13 browser checks on the new grid.
The existing 100 m population comparison already exercises these functional
settings; the 18-case suite was not rerun. No missing tool/credential blocks work.

The user clarified four journeys: nearby counts after JAM GATI, threshold
notifications/decisions, location-checked arrival with going/arrival statistics,
and a citywide aggregate activity map. FUNCTIONALITY_STATUS.md audits each journey
and maps the broader MVP to present/tested, partial and missing evidence. Core private flow is present; public
aggregates/notification subscriptions, 100k validation, independent audit and
production/deployment verification remain incomplete. Next: milestone 09 under the
disclosed inference limitation. No privacy or deployment certification is claimed.


## Plan for the four missing activity journeys

Prepared ACTIVITY_IMPLEMENTATION_PLAN.md with ten commit targets and acceptance
checks. It defines shared willingness/going/fresh-arrival semantics, a proposed
separate 1 km public grid, existing 5-minute/one-epoch delayed releases, joint
suppression, bounded snapshot capture, cached APIs and Albanian cards/map/join flows.
It also covers server-side invitation discovery (some current offers depend on
polling), an atomic bounded outbox, optional push with an expiring opt-in resume
record, area follows, and a 3,000-person simulation extension using real released
activity instead of injected public discovery.

The plan preserves continuous private matching and the current operating config.
It does not equate thresholding with anonymity or treat the accepted private
counterexample as approval of additional public leaks. The precise publication policy,
new adversarial fixtures, capture consistency and real push interoperability remain
implementation gates. Estimated effort: 75–125 focused hours, with privacy redesign
and provider/device troubleshooting outside a guaranteed schedule.

Validation for this planning-only change: inspected current code/API/config/tests,
reviewed W3C/Apple push lifecycle and platform guidance, checked Markdown links and
git diff formatting. No application tests were rerun and no missing feature was
implemented. Local implementation needs no new credentials through fake delivery;
real device/secure-origin interoperability is a separately recorded final check.

## Activity implementation — policy and user correction

Decision 0007 records accepted inferred information, public gathering geography in
the same 1 km cells as willingness, and optional user-chosen push. Updated the
requirements, activity/development plans, agent instructions and threat model.
Schema 7 adds typed public area size, capture duration and snapshot-byte limits.
The pure activity builder implements willingness/going/fresh-arrival semantics,
minimum suppression, configured buckets, delayed release/retention timestamps,
deterministic bytes, and cell-only gathering entries. It rejects duplicate inputs,
invalid cells, oversized output and excessive capture duration.

Validation: `go test -race ./internal/activity ./internal/config` and
`npm --prefix web run check` passed. Initial tests caught an overbroad field-name
assertion (map dataset version is public) and the fast simulation's capture deadline
exceeding its release interval; both corrected. Existing installed pinned tools
were found through scripts/env.sh; no additional installation was necessary.
The threshold-transition fixture records accepted inference, not a privacy proof.
Public publisher/API/UI and push remain incomplete at this commit.

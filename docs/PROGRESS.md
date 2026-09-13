# Development progress and handoff

Updated: 2026-09-13. Status: milestones 02–08 implemented; nearest-crossroad/device-only correction implemented and validated. The 3,000-person response simulation suite is complete. Milestone 09 now has the current activity publisher/API/cards/cell map; daily history remains. Decision 0007 accepts documented inference for implementation.

## Current task and authorization

The user resumed implementation after documentation preparation. Installing needed
tools, implementing/testing milestones and making incremental local commits are
authorized. Docker access is now available; continue through the plan. No remote Git push or public deployment
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
  Delayed public cell maps and willingness/going/here statistics are implemented.
  Optional temporary background notifications are implemented with local fake-provider
  and browser evidence; real provider/device verification and daily history remain.

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
| 09 — Aggregate map/statistics | Current activity implemented; daily history remains | Cell-only delayed releases/API/cards/map, real-store and browser evidence below; inference accepted in decision 0007 |
| 10 — Notifications and follows | In progress | Worker-side offers, foreground display and temporary optional push/outbox/browser resume implemented; live provider/device verification and follows remain |
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

USABILITY_IMPLEMENTATION_PLAN.md selections 1/2/3/5 are implemented locally:
optional push, one-action willingness, private destination preview and clearer
waiting/recovery. Final integration evidence is recorded below as it completes.
Sharing links/QR (4) and new purpose copy (6) remain excluded. Default push stays
off until the operator supplies a real public project/security-contact URL and
provider/device delivery is tested. Local service keys already exist in ignored
.runtime; no credentials should be posted in chat. No native location verification
is claimed. Broader daily summaries, follows, deployment hardening, 100k benchmark
and independently verified releases remain separate incomplete milestones.
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

## Activity implementation — publisher and API

Added separate publisher loop/lease, bounded ZSCAN activity capture with deduplicated
credential reads, end-of-capture expiry checks, and cell-only aggregate staging.
Observation spans an interval: concurrent transitions may be sampled on either
side, not a fictitious atomic instant. Stable index members are visited; churn is
sampled once per credential. Cancellation after a sampled read can remain in that
interval's aggregate, just as cancellation after publication cannot recall it.
Capture deadline/capacity/work errors discard the entire attempt.

`GET /api/activity/latest` reads only bounded precomputed aggregate keys, enforces
release/expiry with store time, rejects query filters, and supplies ETags/cache
lifetimes bounded by expiry. Authenticated requests remain no-store. Empty/unready
publication returns 204, not a zero count. Epoch keys are config-scoped, immutable
SET NX with owner fencing, native TTL and logical deadlines. No participant capture
is persisted. Added restricted ZSCAN permission for the activity index; secrets
remain unchanged. Simulation clock steps also drive the isolated publisher.

Validation: `make verify-local` passed (Go/race/simulation/vet, TypeScript/build,
configuration, Compose parsing, HTTP smoke). `make test-store` passed against actual
Valkey: paging, duplicate creates, cancellation/churn, cancelled capture, lease
fencing, immutable release, delayed visibility, native/logical expiry, cached HTTP,
100 concurrent identical public reads, auth no-store and query rejection. These are
functional checks, not a 100k benchmark. Logs: /tmp/gati-activity-verify.log and
/tmp/gati-activity-store.log. Next: statistics/map browser implementation and checks.

## Activity implementation — browser cards, cell map and joining

The same public snapshot now supplies nearby willingness and per-gathering going/
here cards before and after arrival, with neutral suppression and observation time.
The city map stays visible through all states. Its two layers use only public cell
polygons; cell selection displays the published bucket and gathering list without
changing participant location or querying individual state. Multiple events share
a cell; private actionable destination maps remain separate.

Public-map joining requires an explicit choice. New participants supply a fresh
device fix and confirm PO, PO SHKOJ through atomic /api/join. A lost response retains
the same temporary admission capability and request; no second enrollment is made.
An existing participant confirms through a dialog and the existing /api/going path.
Map browsing/area selection never enrolls. Expired snapshots disable map joining;
server checks remain authoritative. Added a temporary pending-join ID to existing
tab session storage, removed on acknowledgment/cancellation/expiry; no push provider
or durable browser resume storage was added.

Validation: full `make test-browser` passed 17 checks; four public presentation/join
fixtures are distinguished from real API participation tests. After final selected-
cell text and Albanian 24-hour formatting, the four focused activity browser tests
passed again. Inspected reports/local/activity-here.png after both maps were ready.
`make verify-local` passed; focused activity race tests passed after the nested-count
inference fixture. `make simulate SCENARIO=tirana-evening` passed with a real
publisher/API extension: private confirmed presence can remain publicly suppressed.
`make check-containers` passed after the rebuilt web server became ready (an initial
immediate post-restart socket-close check was retried only after HTTP readiness).
The running Compose app uses schema 7, private 100 m/public 1000 m grids. The
publisher-enabled 3,000-person seed-42 run and report checker passed; evidence is in
reports/activity-3000.md. All count metrics match the prior 100 m seed-42 baseline.
Logs: /tmp/gati-activity-browser.log, /tmp/gati-activity-browser-focused.log,
/tmp/gati-activity-simulation.log, /tmp/gati-activity-container-check.log.

Final activity review added a regression guard: public JEMI KËTU also requires the
captured fresh-arrival count to meet the configured private confirmation threshold,
so stale stored state cannot retain the label after expiry. Focused activity race
tests passed. The 3,000-person run accepted 3000 credentials, invited 1891, recorded
1067 unique arriving credentials / 1498 arrival confirmations, and observed 41
meetings (30 ever confirmed), in 221.33 wall seconds. It ran concurrently with other
checks; no load/capacity claim is made. See reports/activity-3000.md for source-dirty
provenance and the distinction between real publication and injected discovery.

## Selected usability task — plan

Recorded selected suggestions and implementation/test sequence in
USABILITY_IMPLEMENTATION_PLAN.md and decision 0008. Inspected current code, requirements,
recent commits and toolchain. This planning commit does not claim new behavior or
test passes. Next is one-action willingness with waiting/recovery, then eligible
private preview and optional background delivery. Existing commands.txt is untouched.

## Selected usability slice B/C — 2026-09-13

JAM GATI now obtains one fresh device fix and submits, with an abort action and no
on-load location access. Existing coarse-only, uncertain-response/retry and expiry
bounds remain. Waiting, offline/recovery, server restart/expiry and arrival freshness
have explicit Albanian guidance. Public-map joining previews the private crossroad
and deadline before PO, PO SHKOJ. Preview-only capabilities remain in page memory;
no signal, reservation, enrollment or intent is created. Existing capabilities must
use their immutable claim. Preview shares network create/write budgets, accepts
only current released IDs, checks live reachability/cutoffs, uses no-store and expires
within the configured location-fix age (and release/gathering/session bounds).

Validation: `make verify-local`, `make test-store`, `make simulate` passed. The
40-person synthetic run accepted 40, cancelled 6 and verified 34 expiries; its
fixed-clock real-store HTTP checks now also cover preview eligibility/no enrollment,
unchanged existing participation/public snapshot, malformed/duplicate/exact-location
input, unknown IDs and reachability. `make test-browser` passed 19 Chromium tests,
including new/active preview cancellation, offline recovery and stale confirmation;
public UI fixtures are distinguished from real API integration. No push provider
was contacted and no production readiness or 100k throughput is claimed.

## Background invitation discovery — 2026-09-13

The matching worker now offers open gatherings to waiting sessions without a
status request, independently of push preference. It reuses `candidate_batch_size`
for bounded attempts and a rotating private cursor across reconciliation snapshots;
atomic intent checks enforce live deadlines, immutable claims and decline cooldown.
Already assigned/reserved sessions are skipped. Pending destinations still prevent
competing founding cohorts. This changes when background sessions become invited,
so older population totals are historical until rerun. `make simulate` passed its
40-person run and race-enabled fixed-clock real-store test, including direct store
creation followed by worker-only invitation (no HTTP status poll). Delivery is not
implemented yet. Next: bounded optional subscription/outbox and transport.

## Usability validation and handoff — 2026-09-13

Plan: `docs/USABILITY_IMPLEMENTATION_PLAN.md`. Commits 55ac165 (scope), 84c984d
(one-action willingness/recovery/private preview), d564fce (background offers).
Selected suggestions 2/3/5 are implemented; suggestion 1 has background discovery,
with optional Web Push registration, transport and browser delivery still pending
(steps E–H). Suggestions 4/6 remain excluded. No new credentials were needed for
these local milestones; no external publication/provider delivery took place.

Nineteen Chromium tests passed, including arrival-confirmation expiry returning the
JAM KËTU action without automatic renewal. `make verify-local`, `make test-store`,
`make simulate`, focused Go race tests and `make check-containers` passed. A final
cursor-lifetime review added an absolute participant-expiry bound checked even on
follower replicas; Go race tests and the fixed-clock real-store simulation passed
again after that small correction. The 3,000-person run was compiled at d564fce,
before that cursor-expiry correction; its exact source/dirty status and artifacts
are recorded, so it is not presented as a final release build.

`make simulate-population SCENARIO=tirana-population SEED=42 OUTPUT=reports/local/tirana-usability-3000-42`
and `make check-population REPORT=reports/local/tirana-usability-3000-42/report.json`
passed: 3,000 accepted, 1,921 uniquely invited, 1,112 unique arrivals / 1,589 arrival
confirmations, 46 gatherings (30 ever confirmed), 22 joins after JEMI KËTU, 456
cancellations and all remaining 2,544 expiries verified. Wall time 222.49 seconds;
functional time is accelerated, not a capacity benchmark or human usability study.
See `docs/reports/usability-3000.md` and its synthetic summary for scope/provenance.
The matching configuration remains schema 7 / 100 m default; no behavior thresholds
were changed. Current development UI/API have been rebuilt locally.

Next concrete work: step E of the usability plan—temporary opt-in subscription and
bounded/deduplicated outbox with cancellation/expiry tests, followed by Web Push
transport and explicit browser opt-in/resume. No working-push button is exposed
before delivery is connected. Preserve the unrelated untracked `commands.txt`.

## Optional push step E — 2026-09-13

Schema 8 adds explicit disabled-by-default push policy, operator contact/provider
allowlist and worker/transport/retry bounds. The store layer now holds one encrypted
transport slot/coalescing outbox per opted-in live signal, with atomic claim fences,
bounded retries, unchanged willingness, quota-preserving opt-out, cancellation and
TTL/index cleanup. Decision 0009 refines transition-written queues to a bounded
latest-state reconciler; short-lived intermediate transitions may coalesce. API,
encryption, provider transport and browser opt-in are the next integration work.

`make test-store` passed with actual restricted Valkey and the race detector:
idempotent registration/lifetime bounds, no enrollment/intent change, 20 concurrent
claims producing one winner, stale acknowledgements, retry ceiling, notification
gap across opt-out/re-opt-in, cancellation and native expiry. A Lua reserved-field
syntax error was found on the first run and fixed before the passing run.
Configuration rejection tests passed for concurrency/retries/timeouts/hostnames and
contact schemes. Existing Web Push interval/TTL settings are now used by the store
policy integration; no provider was contacted. Earlier schema-7 hashes/reports are
historical. Source and local API/client now use schema 8.

## Optional push step F — 2026-09-13

Provider transport and private subscription API are connected. Endpoints/keys are
AEAD-encrypted in Valkey, bound to capability hash + random delivery binding. Exact
HTTPS host allowlists, public-address DNS validation/direct IP dialing, verified TLS,
no proxy/redirect forwarding, bounded body/time/concurrency and queue/attempt fences
protect delivery. Public VAPID key/config is separate; registration/status/opt-out
are authenticated and no-store. `make push-keys` generated project-local service keys
without printing them; they are ignored by Git. The optional Compose overlay mounts
only that read-only file. Default push remains disabled; the localhost contact
placeholder cannot enable production-profile delivery. The user was asked for a
public project/security-contact URL while independent code/test work continues.

`make test-store` passed restricted-Valkey + race tests, including new encrypted
transport/fake-provider and API auth/query/origin/strict-input/opt-out checks.
Transport tests cover public/private/mixed DNS, direct checked-IP dialing and redirects.
The first fake-delivery fixture had a gathering shorter than the configured admission
minimum; it was corrected before the passing run. `go mod verify` passed. Pinned
`make security-check` reported no reachable/imported vulnerabilities, plus the unused
OpenPGP module advisory documented in THIRD_PARTY. Compose base/optional parsing
passed. No real endpoint/provider was contacted. Browser opt-in/resume/worker source
is being integrated and has not yet completed its dedicated tests; the existing
19 browser tests still pass with push off. Next: step G/H browser lifecycle checks.


## Notification race and index review — 2026-09-13

Opt-out now advances a temporary signal revision, rejecting delayed earlier
registrations. Ordinary signal responses strip the revision. Expiry cleanup removes
notification index entries before releasing shared admission capacity. Stable state
changes wait for the rate gap before starting queue TTL; otherwise the default
10-minute effective gap could discard a 5-minute queued JEMI KËTU update entirely.
`make test-store` passed actual restricted-Valkey race checks after these fixes,
including stale registration, quota preservation, stable post-gap presence and
index cleanup while another long-lived subscription keeps the shared key alive.
`make verify-local` passed. Browser lifecycle integration and final evidence follow.


## Optional push browser and local integration — 2026-09-13

Steps G/H now have local evidence: explicit permission/opt-in, first-party manifest
and worker, temporary metadata-only IndexedDB resume, live-state fetch before generic
Albanian display, no private response cache, and opt-out independent of willingness.
Missing/denied push and a delayed optional configuration request do not block the
first action. Provider/browser operations are bounded; late registration/cancellation
callbacks cannot recreate the resume record. Reopening recovers the existing
capability; uncertain recovery offers retry/cancel without enrolling again.

All **28 browser tests passed** against the rebuilt local API and current built
client: 19 participation/activity checks plus nine push lifecycle checks. The push
suite uses full Chromium headless; CDP injects a PushEvent into the installed worker
with the app page closed, exercising IndexedDB, private fetch and native generic
notification display. Provider registration and delivery are synthetic fixtures;
no live browser provider or real OS click was tested. Tests cover denial/unsupported,
revocation, expiry, lost reply/idempotent retry, cancellation during registration,
stale/forged push metadata, closed/reopened page and uncertain server recovery.

A preliminary full-suite run timed out in the existing collective-arrival test and
its cleanup obscured the failing step. A focused rerun passed; stage labels and
bounded cleanup were added. The final full suite passed in 24.7 seconds. The
original timeout's cause is unconfirmed; it is not counted as a passing run.
The initial push reopen test also needed to wait for a visible recovered section
before inspecting sessionStorage. Native persistent notifications worked in full
Chromium, whereas the headless-shell variant rejected notification permission.

`make verify-local`, `make test-store` (actual restricted Valkey/race, including HTTP
stale-revision rejection), `make simulate` and `make check-containers` passed. Local
API/web images were rebuilt; push remains off. Effective schema-8 config SHA-256:
13aedb29ea15f34e845cd4234e124d1047397fc1cfc631d4fb162cbdf51503aa.
The latest small browser cleanup timeout correction is tested by the final browser
build; a final development web-image refresh and schema-8 3,000-person regression
will follow. Setup, lifecycle and manual provider checks are in NOTIFICATIONS.md.

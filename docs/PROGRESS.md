# Development progress and handoff

Updated: 2026-09-14. Local implementation includes temporary willingness, continuous
crossroad matching, late admission, arrivals, delayed cell-only statistics/map,
optional push, private monitoring, measured 100k load and deployment/release tooling.
No public deployment or independent audit has occurred. Read this summary first;
entries below are dated historical evidence, not competing current instructions.

## Current task and authorization

The subsequent About-section request is implemented locally: Albanian peaceful
Flamingo protest support, daily Kryeministria gathering plus nationwide aspiration,
voluntary coordination and bounded privacy/security claims. The same-page section
preserves participation and uses the public GitHub source link. Production build,
41 Chromium tests and desktop/mobile accessibility/visual checks passed; see the
About entry at the end of this document. No public deployment or remote push.

Both auditability cleanup batches are complete. The user subsequently authorized
three lifecycle priorities: expired gathering metadata cleanup, ended public views
and explicit early presence renewal. They are implemented locally and the regression checks passed, with evidence
recorded below, while OVHcloud provisions the VPS. jamgati.com is the chosen domain; the accepted route is
OVHcloud + Cloudflare Free. Exact host sizing/OS, SSH, zone activation and private
reporting/alert channels remain unverified. Preserve commands.txt. Do not request
provider/identity decisions again; no remote Git push is authorized.

## Current implementation and evidence

| Plan milestone | Current state | Evidence / remaining work |
| --- | --- | --- |
| 01 — Product/threat boundaries | Documented; residual limits accepted | REQUIREMENTS, THREAT_MODEL, decisions 0001/0004/0007 |
| 02–05 — Toolchain, geography, willingness and UI | Implemented locally | Pinned tools; 6,491 imported crossroads; 100 m private cells; fresh device-only location; 30-minute minimum; expiry/retry/cancel tests |
| 06 — Tirana simulation | Implemented | Parameterized 3,000-person response scenarios and smaller adversarial fixtures; reports/population-3000.md and current cleanup evidence below |
| 07–08 — Activation/admission and arrival | Implemented locally | Continuous matching, frozen destinations/deadlines, late admission, nonce/replay/freshness/stability checks; real-store and fixed-clock journeys |
| 09 — Aggregate activity | Current map/cards implemented; history incomplete | Fixed delayed suppressed/bucketed releases, cell-only public gatherings; daily summaries remain |
| 10 — Notifications/follows | Optional session push implemented locally | Real-store/fake-provider/browser worker checks; real provider/device delivery, geographic follows and area alerts remain |
| 11 — Hardening/monitoring | Local automation implemented | Private bounded monitoring, replay/race/fuzz/recovery/pressure checks; realistic maximum-lifetime soak, real host/edge and independent audit remain |
| 12 — 100k capacity | Measured locally; deployment capacity unverified | reports/scaling-2026-09-14.md: component and mixed-load measurements; not a VPS or 1M-user guarantee |
| 13 — Deployment | Local production/rollback rehearsals implemented | API/worker roles, trusted ingress, private Compose, runbook; actual OVH/Cloudflare/TLS/host checks remain |
| 14 — Verifiable releases | Local artifact/runtime verification implemented | Exact image audits/SBOMs, two clean same-host builds, served-asset/runtime comparisons; independent build/signature/host evidence remains |
| 15 — Independent audit/pilot | Not performed | Requires external review, delivered host, actual edge/device checks and capped pilot |

Runtime remediation passed locally on 06fa798 with two exact unused-code exceptions
expiring 2026-10-14; see RUNTIME_SECURITY and reports/runtime-remediation-2026-09-14.md.
That audit covers its exact localhost release only. Source changes and jamgati.com
need a newly prepared release and fresh audit before public activation. Workstation
swap fails host preflight; it has not been disabled. No capacity claim is refreshed
by formatting/refactoring tests.

## Auditability cleanup

- 4560a64: readable release/audit/public-verification helpers, pinned formatting and
  unchanged Python syntax trees apart from docstring indentation; rollback passed.
- b0a1dd8: explicit own-session/private-invitation response allowlists, nested JSON
  contract tests and 33 passing Chromium tests.
- 26242b1: shared strict JSON parsing, malformed-input matrix, old/new differential
  fuzz comparison and permanent round-trip fuzz target.
- 3d3015c: removed unused bookkeeping/wrapper, labeled reserved settings, added
  audit entry points and reconciled stale status documents. First batch complete;
  final evidence is in reports/auditability-2026-09-14.md.
- fe8a5ea / 36616d6: shared private-map helpers, browser session controller and
  startup/cancellation/WebGL/map regressions. Final Chromium suite: 36 passed.
- f600502: readable Lua transaction layout and docs/TRANSACTIONS.md. All 34 script
  fragments passed syntax/token comparisons; real-store race and fixed-clock
  journey checks passed without changing atomic checks or deadlines.
- 05b2b68: test-only deterministic 3,000-participant planner replay with the full
  Tirana map/current 100 m config; CI target added. Runtime randomness unchanged.
- Final evidence: [follow-up report](reports/auditability-followup-2026-09-14.md).
  Full 3,000-person API scenario/checker, normal-clock public-map/late-join journey,
  two-API production rehearsal and byte-identical clean builds passed. One earlier
  injected-push test timeout is documented; focused repeats and full suites passed.

## Working commands and next action

Start with [AUDIT](AUDIT.md) for source→storage→test links and [DEVELOPMENT](DEVELOPMENT.md)
for setup. `make audit-local` includes real Valkey integrations; plain `make test`
does not. `make audit-journeys` runs Chromium, real delayed publication and the local
production-image rehearsal sequentially. `make simulate-population` uses current
settings; `make load` and monitored component benchmarks also exist. All are local
synthetic checks, not remote deployment or external push evidence.

The approved lifecycle priorities are implemented: bounded gathering-index and
session-link cleanup, ended public cards/map cells, and **JAM ENDE KËTU** renewal
with fresh device location during the last 120 seconds of presence freshness.
Schema 9 exposes `arrivals.renewal_window_seconds`. Renewal preserves one counted
contribution, stability and the original session/gathering deadlines; willingness
extension, gathering extension and destination relocation remain deferred.
See [decision 0014](decisions/0014-bounded-presence-renewal.md) and the completed
[lifecycle test report](reports/lifecycle-2026-09-14.md).
Hosting work still needs the delivered server's nonsecret connection details and
privately arranged SSH/Cloudflare access. Prepare and audit a fresh exact release
before activation; the older runtime audit does not cover this source. Keep
optional push off until provider/device verification.
Historical entries follow; their old next-action lists are superseded by this one.

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


## Selected usability slice: final evidence and next action — 2026-09-13

All selected code (1/2/3/5) is locally implemented and incrementally committed.
Suggestions 4/6 remain excluded. The final web image and running stack passed
`make check-containers`; `/api/push-config` confirms default push is off. Base and
optional Compose configuration parsing passed. The repeated pinned
`make security-check` again reported zero reachable/imported-package vulnerabilities;
the one unused-module OpenPGP advisory remains documented, not silently omitted.
Local documentation links and `git diff --check` passed.

The final schema-8 3,000-person Tirana seed-42 run and `make check-population` passed:
3,000 accepted, 1,920 uniquely invited, 1,715 going, 1,107 unique arrivals / 1,554
accepted confirmations, 46 gatherings (30 ever confirmed), 24 joins after JEMI KËTU,
456 cancellations and all other 2,544 session expiries checked. Wall time 239.73 s;
262,740 HTTP requests. This is accelerated functional simulation, with push disabled
and synthetic discovery; it is not external push verification or a capacity test.
Compiled source ead158a, source-dirty flag and artifact/config/map hashes are retained
in reports/usability-final-summary.json. Only documentation changed afterwards.
See reports/usability-final.md and the ignored local replay for exact scope.

Next input needed for live push setup: a real public operator project/security-contact
URL to replace the localhost placeholder, followed by supported browser/device
interoperability testing. Keys already exist locally; do not request passwords or
personal participant information. Default functionality works without push. No remote
Git push or public deployment occurred. Broader follows/daily summaries, hardening,
100k benchmark, production deployment and independently verified release/audit remain
incomplete. Preserve unrelated untracked commands.txt.

## Predeployment testing review — 2026-09-13

Reviewed current test sources, configuration, recorded usability evidence and the
remaining hardening/performance/release gates in response to the user's question.
Proposed additional work is in PREDEPLOYMENT_TEST_PLAN.md: isolated real end-to-end
journeys, crash/replica recovery, overnight expiry/churn, real-time load, fuzz/privacy
probes, browser/device/human checks and independent build/audit rehearsal. Pilot
monitoring should use bounded service-health metrics and existing protected releases,
with no participant tracking or synthetic contributions to real gathering counts.
No new test executions, implementation changes or deployment occurred in this review.
The next recommended local batch is test isolation, failure recovery and retention
soak/fuzzing. Live push contact/device verification remains unresolved separately;
it does not block these local checks. Preserve commands.txt.

## Automated hardening: monitoring and fuzz foundation — 2026-09-13

User authorized the proposed automated checks, including monitoring during scale
runs. Added an optional private Unix monitoring socket: fixed completed-window,
suppressed/bucketed request statistics and four task-health summaries, no participant
labels or raw errors. First-party collection bounds retention and validates fields;
its failure cannot gate participant requests. Decision 0010 and MONITORING describe
scope and limits. Functional schema/thresholds remain unchanged.

`make test-monitor` passed Go race/socket/suppression/expiry/canary checks and four
Python collector/alert/bounds checks. `make test-store` passed with real restricted
Valkey after instrumentation. Four bounded fuzz targets (configuration, request
parsers, capabilities and geography) ran for 30 seconds each and passed; the existing
configuration fuzz target was strengthened rather than duplicated. These runs do
not replace unbounded fuzzing or a security audit. The first compile exposed a
duplicate fuzz name and incorrect parser arity, corrected before passing checks.

The owned browser lab passed all 28 checks. It now uses authenticated store readiness;
the first TCP-only readiness attempt raced startup and failed, with no passing claim.
Monitored 1k/10k load smoke passed; the full 100k run, normal-delay real publisher
journey, failure recovery and 15-minute short-TTL churn tests are still running.
Initial memory collection sampled the container supervisor; the larger run measures
the actual Valkey process. Recovery exposed dynamic Docker port changes across a
restart; the owned lab now binds an explicit private test port. Results follow.

## Automated hardening: adversarial transitions and provider outage — 2026-09-13

Added a seeded 250-step arrival/retraction/nonce/cancellation model checked against
private state and independent-credential index counts after each operation. Added
configuration rollover checks: pending cohorts unlock on version mismatch while an
already activated destination/deadline stays frozen. Added 100 encrypted synthetic
push subscriptions with two competing workers and a failing fake provider; exclusive
claims, configured concurrency, retry backoff and cancellation passed without any
external delivery. `make test-store` passed with the race detector in 2.722s/1.092s/
1.651s for store/API/notification packages. Initial test assumptions about duplicate
nonce issuance and comparable map structs were corrected before this passing run.

The collector now creates temporary output files with mode 0600 from the start,
rejects invalid object/worker schemas and tests rejection of sensitive extra fields.
Python monitor and load-gate checks passed (eight total). Load checks distinguish
accepted-request latency from fast rejections; historical combined histograms do
not silently qualify mixed traffic. Browser compatibility, the normal-delay journey,
short-TTL churn and final load/reproduction evidence are still being completed.

## Automated hardening: owned monitored lab and retention — 2026-09-13

The owned production-binary lab now supplies private monitoring, independent ports,
a nonpersistent restricted Valkey, config/binary/source/hardware manifests and
teardown. It refuses to overwrite an existing run manifest. Load cannot accept an
arbitrary remote URL; synthetic capability values remain in generator memory.

Passing evidence: two-replica process/store recovery (six checks, 132.52s); the real
normal-delay publisher journey through public map/preview/late arrival (13.8-minute
test, 13.9-minute suite); and the 15-minute short-TTL storage soak (901.203s, 695
cycles, 69,500 creations, 23,196 cancellations, every cycle's indexes restored to
one expected live keeper). A first soak completed its checks but failed writing a
relative report path; the path is now absolute and the entire run passed again.
The initial publisher test allowed only 11 minutes; actual once-per-epoch capture
plus delayed release can approach 15 minutes from a user's action. The corrected
bounded wait passed without changing product timing or mocking API responses.

The owned 8-MiB no-eviction pressure test passed: saturation returns 503, a live
credential retains its deadline after pressure, uncertain creation can be retried,
and cancellation remains terminal. Temporary filler keys exist only in that lab.

Uniform and hotspot ramps each accepted 100,000 signals. Dense-hotspot client gates
passed with separate accepted/rejected histograms, no timeout/5xx or dropped jobs.
The 3,000/s burst accepted 500/s and rejected 2,500/s with explicit 429s; post-burst
reads recovered. This PC is an AMD Ryzen 7 7435HS, 16 logical CPUs; API/store/clients
share it, with other isolated checks also running. This is not the planned remote
4-vCPU/8-GB reference host, simultaneous mixed traffic, cached CDN test, or direct
matching/expiry-lag proof. Final uniform split-latency report and clean builds follow.

The final uniform ramp now also passes the explicit accepted-latency client gate.
Both final workloads accepted all 100,000 signals, with accepted p95 <=10ms in each
phase, zero client timeouts/5xx/dropped jobs, and no observed operational alerts.
Sampled peak RSS: uniform API 492.18 MiB / Valkey 88.72 MiB; hotspot API 506.37 MiB /
Valkey 90.22 MiB. Offline CPU/RSS charts and exact synthetic phase counts are retained
in reports/local; the tracked predeployment summary records scoped evidence/hashes.

Firefox and WebKit each passed 24 core/accessibility/resilience checks. The browser
adapter corrects only the automation-provided location timestamp (Firefox's future
value and WebKit's microsecond value); production stale/future/inaccurate location
rejection remains unchanged and is explicitly tested. The optional verified local
Ubuntu WebKit library setup avoids sudo and incompatible inherited Snap GIO paths.
The normal Chromium suite's final push injection test exposed a worker lifecycle
race and is being retested with explicit CDP worker startup. No production push
behavior has been changed to accommodate an injected provider event.

Final browser validation passed: 33 Chromium checks (1.0m), 24 Firefox and 24 WebKit
checks (about 1.1m each), plus the separate normal-delay journey above. The injected
push regression passed five consecutive focused runs (15.2s) after explicit CDP
worker startup, followed by the full Chromium suite; no retries hid failures and no
production push code changed. Verified browser prerequisites, CI fast gates, command
documentation and scoped evidence are included in this milestone. Clean-source byte
reproducibility is the next check and requires this implementation commit first.

## Automated predeployment batch complete — 2026-09-13

`make reproduce-build OUTPUT=reports/local/reproduction-final` passed for clean
commit 4da585e: two clean exports with separate compilation caches produced identical
hashes for the backend and all seven frontend artifacts. Backend SHA-256:
1dd8f14687a80bc69d43ab1ba40d92fcca83d093c6698deafab82a088d66f23b.
Go 1.27.1 / Node 24.21.0, same Linux host. Dependency inventories record Go modules
and 92 npm entries, all with license metadata; this is not a full license or
independent supply-chain audit. The production binary also rejected the simulation
configuration. Only evidence documentation changed after the reproduced commit.

The implemented local batch is complete: private monitoring; isolated real browser
journey; browser/accessibility/resilience checks; seeded state/fuzz/race probes;
process/store/replica recovery; bounded fake-provider fanout; short-TTL churn and
no-eviction pressure; two monitored 100k ramps/bursts; report gates/charts; and clean
local build comparison. Exact results and known failed attempts/corrections are in
reports/predeployment-automation.md and its summary JSON; commands and boundaries
are in AUTOMATED_TESTING.md. Long reports remain under ignored reports/local.

Remaining release work is explicit: overnight/full-API realistic 120-minute session
churn, simultaneous mixed traffic and mass admission/synchronized expiry, direct
matching/expiry-lag measurements, 10k cached public reads/s, remote reference-host
sizing, provider/public-release freshness monitoring, real devices/live optional
push, volunteer usability, pedestrian destination review, independent human audit,
actual host/edge controls, deployment/rollback and served-artifact verification.
A real public operator contact and device setup remain needed for live optional
push; they do not block default local functionality. This batch does not certify
production readiness, location authenticity, unique humans or absolute anonymity.
No remote push or public deployment occurred. Preserve untracked commands.txt.

Final `make security-check` reported zero reachable vulnerabilities and zero in
imported packages. The previously documented advisory in an unused required module
remains; it is not silently omitted. All owned lab processes exited and tracked
changes pass `git diff --check`. Next local testing priority: realistic-lifetime
full-API churn and direct matching/expiry-lag evidence before reference-host sizing.

## Architecture explanation and 1M estimate — 2026-09-13

User requested a scaling/memory assessment and visual architecture explanation.
Reviewed current code/configuration and existing monitored 100k samples; added
ARCHITECTURE.md and editable Graphviz DOT with rendered SVG/PDF. The diagram
separates current Compose components from optional monitoring/push and planned
Caddy/edge deployment. README links the explanation.

As configured, admission caps at 150k. The 500/s write budget, increasing-offset
population scans, one matching writer, background offer budget, publication capture
and cleanup backlog require further scale work. Offset traversal cost was checked
against official Valkey documentation. No 1M run or behavior/configuration change
occurred. The approximately 3–6 GiB combined Go/Valkey allowance at 1M is explicitly
an uncertain extrapolation for willingness-heavy, push-disabled traffic on the same
Tirana map, not measured capacity or a memory upper bound. Actual 100k baseline,
peak values and projection arithmetic are documented.

Validation: rendered SVG/PDF with existing Graphviz 2.43.0, visually inspected the
layout, checked SVG XML/PDF header, local document links and git diff whitespace.
No runtime test rerun was needed for these documentation/vector-only changes.
Next scaling action remains a bounded larger lab benchmark with representative
lifetimes/mixed traffic and direct matching/publication/expiry-lag measurement,
preceded by review of the identified scan/backlog bottlenecks. Preserve commands.txt.

## Deployment preparation and optimization proposal — 2026-09-13

Reviewed matching/store loops, proxy admission, cleanup, local Compose and current
vendor documentation for the user's deployment/cost/anonymity discussion. Added
DEPLOYMENT_PREPARATION.md with a proposed seven-step commit/release sequence,
linked from README/development plan/threat model. This is planning only: no runtime
or functional configuration changed, no account/purchase or public deployment occurred.

Priorities: measure direct matching/release/cleanup lag; remove growing-OFFSET
traversal; schedule expiring geographic eligibility and dirty/deadline work;
index gathering offers; separate API/worker roles before adding parallel ownership.
One matching lease and multi-key Lua operations mean more replicas or Valkey
Cluster do not automatically scale matching. Periodic full reconciliation may
remain linear; no claim of constant-time citywide matching was made.

Found a concrete deployment prerequisite: network rate limits deliberately use
RemoteAddr, so production forwarding needs an explicitly trusted proxy boundary
and spoof/shared-NAT tests. Current Compose is a development scaffold with Vite,
small Valkey limits and no Caddy/edge configuration. Proposed first-host setup is
Caddy/Go/private Valkey behind an outbound Cloudflare tunnel, subject to a documented
provider-visibility decision. Cloudflare's TLS proxy can inspect IPs, coarse inputs
and capabilities; origin concealment and WHOIS redaction are not provider anonymity.
Pricing is dated and qualified by stock/tax; it is not a purchased capacity promise.

Validation: reviewed cited primary vendor pages and source paths; checked 39 local
link targets across the four affected reference documents and git diff whitespace.
No runtime tests were rerun for these documentation-only edits. The new document
clearly labels proposed work and distinguishes a capped pilot from public/1M
readiness. Next local action is representative mixed-load lag instrumentation and
the bounded scan fix. Later staging needs domain/provider choice, credentials via
interactive/local secret setup, and explicit authorization for remote publication.
Preserve untracked commands.txt.

## Scaling baseline and delay instrumentation — work in progress

User authorized the deployment-preparation steps and explicitly requested measured
speedups. Added an owned-store benchmark at 10k/100k/1M, geographic lookup benchmark,
private v2 worker duration/deadline-lag observations and a simultaneous
status/public-read/going load phase. Algorithm changes have not landed yet.
Baseline raw output is reports/local/scaling-before: median snapshot 30.06 ms at
10k, 292.33 ms at 100k and 2.814 s at 1M, three samples of three iterations. The
fixture directly seeds expiring synthetic records and bypasses HTTP admission;
it is traversal evidence, not million-participant application capacity. Timings
were roughly linear on this fixture; theoretical offset costs did not dominate.
Geographic last-hit/miss lookup baseline is also recorded at 100/1k/10k candidates.

Focused monitor/race and disposable store/API/notification integration tests passed;
load report unit checks and make test (Go race, simulation build, vet and TypeScript)
passed. Pending: mixed 100k baseline,
optimization and identical-workload comparisons. Source remains local; no provider
credentials or publication were needed. Preserve commands.txt.

## Incremental geographic reads — 2026-09-14

Implemented ten-second coarse-cell change markers, private owner-only working
sets, periodic full reconciliation, independent deadline wakeups and cache discard
on failure/ownership change. Full snapshots now use an atomic indexed-rank cursor
with bounded retry, batch record reads and reusable decoder storage. Reachability
membership uses binary search without an additional per-cell map. Decision 0011
documents lifecycle and residual linear CPU/full-reconciliation work.

Measurements caught and rejected the first ZSCAN prototype: it reduced allocation
traffic but made the snapshot benchmark 10–20% slower. Raw evidence is retained at
reports/local/scaling-after-scan. The subsequent rank traversal without decoder
reuse was near baseline speed at 100k/1M, with about one-third less allocation
traffic; it was not described as a substantial throughput speedup. Final repeated
benchmarks will include the complete changes and changed-cell reads under both
uniform and hotspot distributions, using an unchanged baseline revision.

The first requested mixed-load run accidentally omitted the mixed flag at the
Python-to-Go boundary; it only establishes sequential 100k evidence. Fixed the
forwarding, added expected-phase checks and a clean-revision binary option to the
owned lab. Corrected reports/local/mixed-100k-before-corrected used original backend
f9a412e and passed: 100k accepted, simultaneous 66,680 status reads, 20k public
origin reads and 1,000 going confirmations, all successful with accepted p95 <=10ms.
It does not measure arrivals, full-lifetime expiry or edge traffic.

Validation: make test-store passed with the added worker integration package;
make test passed Go race/simulation builds, vet and TypeScript after correcting
unkeyed test literals reported by vet. Current seeded 3,000-person simulation
completed in 214.2 wall seconds with 3,000 accepted credentials, 1,922 invited
actors, 1,562 arrived actors and 44 observed gatherings; its paced synthetic time
is not a speedup benchmark. Remaining checks/results are recorded in the next
entry. Local deployment scaffolding is being prepared separately; preserve commands.txt.

## Measured optimization results — 2026-09-14

Final same-harness clean-export comparison f9a412e → 96db982 is retained in
[the performance report](reports/scaling-2026-09-14.md), including raw repeated
component samples and a standalone SVG. Distributed single-cell reads improved
296.38 ms → 0.394 ms (752× component-only); a 100k single-cell hotspot improved
303.88 → 223.78 ms (1.36×). Full snapshots: 10k became 8.3% slower, 100k became
1.085× faster, 1M became 1.059× faster. At 100k/1M allocation traffic dropped about
43%; this is not a live-memory measurement. Reachability last-hit/miss improved
3.97×/29.1×/267× at 100/1k/10k candidates. Materialization/planning and periodic
reconciliation/publication still include population-sized work.

Corrected before/after mixed-load runs both accepted 100k creations and all requested
simultaneous status/public/going traffic, with no 429/errors/generator drops and
p95 <=10ms. Sampled API CPU during the workload fell 176.42 → 142.55 CPU seconds
(19.2%), while peak sampled API RSS rose 505.18 → 559.40 MiB (10.7%). Valkey CPU/RSS
were essentially unchanged. No monitoring alerts; observed worker lag buckets were
<5 seconds at samples, not proof of zero delay. These are one fixed-rate run per
revision on this laptop, not maximum throughput, VPS sizing or 1M-user evidence.

Next local work: finish and verify the already prepared production roles/proxy,
immutable release/rollback tooling and host/edge runbook. No publication authorized.

## Production roles, ingress and release tooling — 2026-09-14

Added API-only/dedicated-worker roles, private docker-exec monitoring, trusted proxy
validation, credential-header cache bypass, pinned production web/Caddy image and
private Compose deployment with bounded resources. Default remains one combined
process. The optional two-replica topology shares atomic state/network budgets and
has one matching owner. Optional push has a separate outbound/mount overlay;
functional settings remain in the published YAML. Added local immutable release,
file/image/runtime/public-asset verification and compatible rollback commands plus
host preflight. User-local Buildx 0.37.1 was installed using its verified publisher
checksum; no administrator credentials or remote publication were used.

Validation: make test (race/simulation/vet/TypeScript), make test-store and make
test-monitor passed. Three release guard unit tests passed. Production Compose and
push overlay validate. The real production-image rehearsal at
reports/local/deployment-roles-cache-fixed passed eight grouped checks in 71.97s:
actual container/store controls, trusted ingress/spoof/query/body limits, cache
isolation, Chromium built-client map/willingness/cancellation, two-API activation
and concurrent arrival replay, one-API failure, separate worker/monitoring roles,
and network budgets across replicas. No real Cloudflare/TLS/push provider was used.

The rehearsal caught and fixed address allocation collisions, Caddy route/header
ordering, loss of the copied client-IP header and the upstream Caddy executable
capability preventing startup with ALL dropped. A pre-existing test sent an empty
Authorization header while expecting shared-cache eligibility; it now omits that
header for anonymous cases and keeps credential-cache rejection assertions.

The host preflight correctly fails on this developer PC because host swap is
active; it was not disabled. Container no-swap checks pass, which does not certify
host/provider privacy. Next: export the committed release, run its actual private
activation/served-artifact/rollback rehearsal, and compare two clean builds. Then
finish documentation/evidence and request only the domain/host/account access and
public authorization needed for the external staging step. Preserve commands.txt.

## Release rehearsal and reproducible artifacts — 2026-09-14

Committed release 81ea543 was exported and built at reports/local/release-81ea543.
Its source/config/files, four image identities and served browser assets are
recorded in release.json; the roughly 94 MiB images.tar contains no participant
state or credentials. Two clean exports with independent compilation caches
produced byte-identical native API and browser artifacts (Go1.27.1/Node24.21.0),
recorded at reports/local/reproduction-81ea543. This is same-host/toolchain evidence,
not independent-host or remote-backend attestation.

The actual combined-role activation, public-asset/config comparison and rollback
rehearsal passed in 124.87s. Two switches between identical code/config release
directories preserved Valkey's container ID and start time and a synthetic live
capability's original expiry; cancellation still worked. All owned rehearsal
containers/networks were removed afterward. It does not certify arbitrary schema
rollback. A subsequent guard rejects an existing operational ACL change without
rewriting it; four release guard tests pass. The guard and final evidence will be
included in one final local release export and rehearsal before handoff.

Sanitized deployment/browser/rollback/reproduction evidence is retained at
[deployment-preparation-summary.json](reports/deployment-preparation-summary.json).
Remaining external steps require a selected hostname/VPS, privately arranged SSH
and Cloudflare authentication, and explicit public-deployment authorization. Keep
host privacy controls, actual edge/cache behavior, target-host capacity, real-device
location/push and private alert routing as separate gates. No remote push or public
service has been created. Preserve commands.txt.

Release verification refinement: record SHA256 of the actual /gati executable from
its production image and reproduce the Dockerfile's stripped binary alongside the
native development binary. This avoids treating differing OCI metadata/image IDs
as proof of differing executable code. The earlier clean-build check remains valid;
a final comparison will explicitly connect the clean stripped artifact to the image.

## Local deployment preparation complete — 2026-09-14

Final guarded release-directory rehearsal b189e86 passed in 123.74s, including
stable live Valkey and capability expiry across both switches. No owned rehearsal
containers or networks remain. Release guards pass all four Python tests.

The reviewable final artifact bundle is reports/local/release-auditable, exported
from 52caa181f2f73a815af2c7b0ddd8d8d2e8b01b4f. Its release manifest SHA256 is
433e129fdef9e983b21fdf7588e6190fd9569082b3d4bfea6602deaff13d42dc. Both independent
clean builds in reports/local/reproduction-container-final matched each other,
and their stripped API executable and all browser files matched the actual
production images. Production /gati SHA256:
90b0c30dd83100208b271c9b81686c9620204e1accafa1e9460a1f2ea546e03d.
The latest evidence-only documentation commit is intentionally outside that source
artifact; executable/config behavior is unchanged. OCI image IDs differ between
builds, so reproducibility claims concern verified executable/browser bytes, not
OCI metadata or remote honesty.

Completed local checks: Go race/simulation/vet/TypeScript, live Valkey integrations,
monitoring/release guards, 3,000-person simulation, repeated 10k/100k/1M component
benchmarks, before/after real-time 100k mixed load, production Caddy/two-API/worker
security and browser rehearsal, compatible release rollback, served assets/config
and two-clean-build-to-image comparison. Exact scopes and exclusions are retained
in the two reports linked above. 1M live-user capacity, real VPS performance and
Cloudflare/TLS/device/push validation remain unproven. No source was pushed and no
service was publicly deployed.

Next action requires user input for the external staging step: chosen domain/VPS,
SSH access arranged privately, Cloudflare login/tunnel credentials provisioned on
that host, and explicit remote/public authorization. Never request tokens/passwords
in chat. Review host swap/hibernation/snapshots, provider plaintext visibility,
real edge cache/spoof limits and operator alert routing before the small pilot.
Keep advanced geographic parallelism and remaining linear planner/publication work
as measured follow-up optimization; current 100k local mixed traffic met its gates.
Preserve unrelated untracked commands.txt.

## Final security review — 2026-09-14

User requested a final code security review and deployment plan. Reviewed the
security-sensitive application/store/browser/push/ingress/release paths and exact
production artifacts. **Public deployment is on hold:** the Caddy image contains
outdated Go/OS dependencies (40 high-severity package/advisory pairs; binary scan
finds affected symbols for 25 advisories). Tunnel package findings need explicit
triage; its Go binary scan found zero affected symbols. Application Go/npm audits
pass, including the explanation of unused OpenPGP GO-2026-5932. Valkey image scan
has zero findings. These are scoped results, not exploit or anonymity guarantees.

Fixed a release-verification gap: compare command/entrypoint/user/environment,
working directory, resource bounds, mounts, network attachments and proxy IPs with
the release configuration and image defaults. Reject omitted live worker/tunnel
services before activation, so omitting --edge cannot silently leave a public
tunnel running while claiming a private launch. Errors omit secret values. Added
mutation regression tests and a live extra-network-attachment rejection rehearsal.
CI now invokes Go/npm dependency audits; full image scans remain a release gate.

Validation: make test (race/simulation/vet/TypeScript; cached unchanged Go results),
fresh make test-store with race/count=1, all four 30-second fuzz targets, 20 Python
tests including 11 release guards, and make security-check passed. Modified local
release/rollback rehearsal passed in 129.94s with actual network-drift rejection,
asset/config comparison and preserved Valkey/session expiry. Owned rehearsal
containers/networks were cleaned up. Trivy 0.74.0 was installed user-locally with
official checksum verification; all four actual image scans completed, and a
clean tracked-tree secret scan found nothing. This is not a full Git-history scan.
Host preflight still fails on developer-host swap, which was not changed.

See [findings and deployment plan](reports/security-2026-09-14.md) and
[sanitized scan/test inventory](reports/security-2026-09-14.json). Raw local outputs
are in reports/local/security-2026-09-14. Existing release-auditable still contains
the old verifier/runtime and must not be published as a cleared release. No source
was pushed and no remote/public service was created.

Next technical action needs no credentials: remediate the web runtime dependencies,
triage/update the tunnel image, add repeatable full-image release auditing, then
build/scan/rehearse a new immutable artifact. Only afterward proceed to independent
review and a chosen domain/VPS, privately arranged SSH/Cloudflare authentication,
security-reporting and alert channels, actual host/edge/device gates and explicit
public authorization. Preserve unrelated commands.txt.


## Runtime remediation in progress — 2026-09-14

User authorized proceeding with review remediation. Added locked Caddy and verified
cloudflared source builds using Go 1.27.1 and patched Go dependencies, scratch
runtimes, dependency license preservation and symbol retention for precise binary
scanning. Cloudflared's four Sentry initializers are disabled by a checked patch;
source tests confirm no client/event even with a synthetic DSN. Caddy now contains
only modules required by the reviewed Caddyfile and a fixed loopback health probe.

Added full image/SBOM auditing with exact, expiring package exceptions, unconditional
Go binary checks and a fresh-audit public activation gate. Release manifests now
record API/proxy/tunnel executable hashes; clean reproduction includes all three.
CEL's direct upgrade broke upstream interpreter APIs; kept its locked version with
an unused-function exception pending symbol verification, rather than changing the
interpreter behavior. Raw failures and corrected builds are retained locally.

Validation so far: patched access/tunnel privacy tests, source vulnerability scans,
Caddy health tests, Python audit/release guards and an initial actual production
browser/two-API rehearsal passed (72.0s). The final symbol-retaining images and
compression patch are undergoing scans and the final production rehearsal. Next:
commit tested source, prepare/scan/rehearse a new exact release and compare clean
builds before claiming the runtime remediation complete. No remote action taken;
keep commands.txt untouched. See RUNTIME_SECURITY.md and decision 0013.

The locked source runtime tests and final production browser/two-API rehearsal now
pass (71.88s, two observed transient 5xx responses during deliberate replica loss,
followed by successful bounded retry/cancellation). The harness now handles plain
proxy errors and does not mistake one successful request for complete DNS recovery.
Application tests/audits, fresh store race integrations and 25 Python tests pass.
Next is the exact committed artifact audit/reproduction; local pre-export scans
have only the scoped CEL/OpenPGP package findings and no affected binary symbols.


## Runtime remediation complete locally — 2026-09-14

Committed implementation 06fa798c409c491f078d36dec155694a67c52fd8 now has a new exact
release at reports/local/release-runtime-06fa798. Manifest SHA256:
ef9b61e9db5e4dfa0597e9f5d2ec2b85fb7a38d67fbdc749c21fefed3bc3ae87.
It uses localhost for rehearsal and must be prepared for the eventual public host.
All four final image audits pass: zero unresolved and zero high/critical package
findings. Four package matches are accepted by two exact unused-code exception
rules (OpenPGP and CEL), expiring 2026-10-14. All three symbol-retaining Go binaries
pass govulncheck; no package exception can override that result. Raw scans and four
CycloneDX SBOMs are retained at reports/local/runtime-remediation/release-audit.

Two clean exports with separate compilation caches produced identical API, Caddy,
cloudflared and browser bytes, and all matched the executable/assets extracted from
the release images. This is same-host/toolchain reproduction, not an independent
reviewer or remote attestation. The exact audit passed the public-audit validator
without running public activation. Actual host preflight still fails on workstation
swap; no workstation swap/hibernation controls were changed.

Final validation: application race/simulation/vet/TypeScript and Go/npm audits,
fresh live Valkey race integrations, 25 Python guard tests, source-patch tests,
patched access/tunnel Sentry no-client/no-event tests, Caddy health tests and source
vulnerability scans passed. Production browser/two-API rehearsal passed in 71.88s
with two explicitly observed transient 5xx during deliberate replica loss and
successful bounded retry/cancellation. The final exact-release/rollback rehearsal
passed in 134.26s, rejecting an added network and preserving store identity/session
expiry through two compatible switches. All owned rehearsal containers/networks
were removed. No scale tests were repeated and no new capacity claim is made.

See [runtime remediation report](reports/runtime-remediation-2026-09-14.md),
[sanitized evidence](reports/runtime-remediation-2026-09-14.json) and
[RUNTIME_SECURITY.md](RUNTIME_SECURITY.md). README, security status, threat model,
plan and deployment runbook now distinguish this resolved local runtime work from
the historical review blockers. Existing old bundles remain uncleared; no source
was pushed and no remote/public deployment or account operation was performed.

Next action needs the user's domain and VPS/hosting budget, then privately arranged
SSH/Cloudflare access and private reporting/alert channels. A concise asynchronous
question requested the nonsecret hostname/provider/budget/SSH-alias information;
no answer has arrived yet. Complete independent review and actual host/edge/device
checks before a capped public pilot, with explicit deployment authorization. Keep
push optional/off until real-provider/device verification. Preserve commands.txt.

## Domain purchased; VPS provisioning pending — 2026-09-14

The user reports ownership of jamgati.com (jegati.com was unavailable) and an
OVHcloud VPS order awaiting delivery. Canonical pilot hostname: jamgati.com.
The user accepted the hosting/registrar government-disclosure boundary and asked
to continue with OVHcloud + Cloudflare Free toward deployment. Preserve that
decision across handoffs; do not repeat the provider-choice/identity approval.
Provider delivery, exact VPS sizing/OS, SSH access, Cloudflare zone activation and
private reporting/alert contacts remain unverified. No remote Git push is authorized.

Updated DEPLOYMENT.md with the selected hostname and remaining handoff. This is
documentation-only preparation: no application settings, DNS, accounts or server
services changed, no new release built, and no runtime tests repeated. Validation:
reviewed the runbook against release.py's --public-host argument and checked the
documentation diff with git diff --check. Existing localhost artifacts remain
rehearsal evidence; the real hostname needs a new release and a fresh audit before
activation. Next: obtain the delivered server's nonsecret connection details,
arrange local SSH key access, then proceed through actual host/edge/pilot checks.
Keep commands.txt untouched and credentials outside chat/tracked files.

## Auditability cleanup: release formatting — 2026-09-14

The user authorized the proposed simplification work. Start with small, locally
committed changes that preserve product/privacy behavior; no remote push or public
activation. Preserve commands.txt. Remaining slices: response allowlists, strict
parsing, dead-code/configuration clarification and a current audit/evidence index;
browser-state and Lua refactoring need their own transition regression gates.

Formatted five release/audit/public-verification/rehearsal Python files and release
guard tests with development-only Ruff 0.11.13. Scope and version are pinned in
ruff.toml; DEVELOPMENT documents the ignored local environment and check command.
The downloaded Linux wheel SHA256 was checked against PyPI publisher metadata:
4ffbc82d70424b275b089166310448051afdc6e914fdab90e08df66c43bb5ca9.
No release orchestration or runtime guards changed. Syntax-tree comparison against
2335b81 passed after normalizing docstring indentation (the first exact comparison
correctly flagged that documentation whitespace change).

Validation: pinned doctor and Docker daemon ready; baseline make test (Go race,
simulation build tests, vet, TypeScript), 25 Python tests and make test-store passed.
Real-store tests use a disposable restricted Valkey; ordinary make test skips those
integrations. After formatting, all 25 Python tests and make format-python-check
passed. The formatted orchestration/verification tools passed a 127.36-second local
rollback rehearsal against the existing 06fa798 release images: unexpected network
attachment rejected, served assets/config verified, two compatible directory
switches preserved store identity and session expiry, cancellation passed. Evidence:
reports/local/audit-format-release/release-rehearsal.json. This validates the current
Python helpers with prior application images, not a new audited production release.
Baseline seeded population and additional response-contract tests are in progress.

## Auditability cleanup: explicit response contracts — 2026-09-14

Own-session, invitation and eligible-preview JSON now use explicit HTTP response
structs, including an explicit nested map-landmark view. Removed Signal.Public's
manual private-field scrubbing; storage structs no longer define these API views.
Wire names, values, optional fields, private destinations, public cell aggregates
and state transitions are unchanged. No new disclosure or storage lifetime.
API/threat documentation points reviewers to the response allowlists.

Validation: response allowlist assertions passed against the old implementation
before refactoring, including the real fixed-clock activation/admission/arrival/
preview journey (reports/local/audit-contract-baseline). Afterward, Go HTTP/store
race tests, make test-store, and make simulate with live simulation integration
checks passed (reports/local/audit-dto-journey). The same journey now checks exact
allowed nested fields on successful responses; unit tests populate private sentinel
bookkeeping and check map geometry/optional-arrival round trips. All 33 Chromium
browser tests passed in 58.0 seconds (reports/local/browser-20260914T080916).

Baseline Tirana population with seed 42 and unchanged config/gati.yaml completed:
3,000 synthetic people/accepted credentials, 1,897 invited, 1,486 accepted arrival
operations (1,067 distinct arrived credentials), 43 observed gatherings, 218.5 wall seconds (reports/local/audit-baseline-population-3000).
An earlier small success-fixture run used 80 people despite its output directory
name audit-baseline-3000; do not mistake that directory for the 3,000-person run.
Next: consolidate strict JSON parsing, compare malformed/valid input behavior,
remove confirmed dead code and reconcile current audit/status documentation.

## Auditability cleanup: shared strict JSON parser — 2026-09-14

Willingness, join/preview, going/decline, arrival and push now share one small
exact-object reader. Endpoint-owned media type/body limit, auth/rate ordering,
domain validation and atomic storage transitions are unchanged. Duplicate names
(including escaped equivalents), unknown/case-variant fields, nulls, missing fields,
wrong types and trailing documents remain rejected. Join no longer reserializes
its willingness fields through a second separate parser.

Validation: the three existing parser schemas passed the new malformed/valid
matrix before refactoring. A temporary differential fuzz harness compared old
parsers with the replacements: 798,358 executions in 31 seconds, identical
accept/reject decisions and accepted values. The duplicate legacy implementations
were removed after comparison. Permanent matrix coverage includes single-field
arrival/intent schemas and reader limits. The new round-trip fuzz target passed
511,121 executions in 31 seconds and is included in make test-fuzz.
Go HTTP race tests, make test-store, make simulate with live fixed-clock journey
checks (reports/local/audit-parser-journey), and make test passed afterward.
Full real-publisher browser and final seeded population checks are running; their
results are not yet counted as passed. Next: dead-code/configuration clarity and
current audit/status documentation, then record the final regression evidence.

## Auditability cleanup: audit entry points and current status — 2026-09-14

Removed the write-only Engine.lastSweep field/assignment and unused MatchingLease
wrapper. The activity publisher's separate lease script remains in use and was
preserved. No Lua transition, worker scheduling or retention policy changed.
The production YAML now labels six reserved fields (including the unused static
intersection-index batch size); their schema membership/values remain unchanged.
Canonical effective configuration before/after is byte-identical.

Added docs/AUDIT.md linking flows, validation, state owners and tests. Reconciled
stale current summaries in PROGRESS, FUNCTIONALITY_STATUS, THREAT_MODEL, STORAGE,
API and development docs with implemented monitoring, 100k measurements, optional
push and production/release evidence. Dated old evidence remains identifiable.
Added make audit-local (native/build/config/smoke, Python guards, real-store checks)
and sequential audit-journeys (Chromium, normal-clock publisher, production images).
CI now calls audit-local in place of the same constituent checks and has a current
workflow name; remote CI has not run. Formatting stays separately invocable and
plain make still selects bootstrap; dry-run checked its default target.

Validation: make audit-local and format-python-check passed; all 55 local links in
AUDIT resolve. Small before/after scenario JSON is identical across baseline, DTO
and parser runs. Both seed-42 3,000-person reports passed expiry/funnel/configuration
invariants with 3,000 accepted credentials, 456 cancellations and 2,544 verified
expiries. Larger scenario totals are not identical: baseline/after has 1,897/1,898
invited credentials, 1,067/1,063 distinct arrived credentials, 1,486/1,501 accepted
arrival operations and 43/45 observed gatherings. Driver credentials and worker
IDs remain cryptographically random; matching uses their hashes/IDs as tie-breaks.
Thus this is invariant/scenario evidence, not deterministic equivalence or a speed
comparison. No 100k/1M benchmark was repeated for this cleanup.

The current production-image two-API/Caddy rehearsal passed in 72.06 seconds with
zero observed proxy 5xx during deliberate replica loss (reports/local/audit-deployment).
It covers runtime privacy settings, proxy/cache boundaries, real Chromium creation/
cancellation, activation/arrival replay, worker separation and shared rate limits.
The full normal-clock delayed publisher journey is still running and not yet
counted as passed. No public deployment, remote push or fresh exact-release audit.


## Auditability cleanup: final first-batch evidence — 2026-09-14

The actual production-binary browser journey passed in 15.4 minutes with normal
matching/arrival/publication settings and clock: willingness, going, fresh arrival,
JEMI KËTU, delayed public statistics, eligible private preview, late admission and
cancellation. The currently released snapshot initially preceded the test gathering;
the test waited for the next scheduled release rather than bypassing privacy delay.
Evidence: reports/local/audit-full-journey. All 33 Chromium checks then passed again
against the final code in 57.1 seconds (reports/local/audit-final-browser).

Two clean source exports of 3d3015c with separate compilation caches produced
identical API/browser/Caddy/cloudflared artifacts using the pinned toolchain.
Evidence: reports/local/audit-reproduction/reproduction.json. This is same-host
reproduction, not independent/remote attestation. The final summary and sanitized
log copies are under reports/local/audit-cleanup; the tracked report is
[the auditability report](reports/auditability-2026-09-14.md). Added the development
formatter's MIT notice to THIRD_PARTY. No new runtime dependency was introduced.

Final checks: audit-local, formatting, real-store/fixed-clock journeys, malformed
request matrix and differential/round-trip fuzz, 33 browser cases, normal-delay full
journey, production-image two-API rehearsal, prior-release rollback rehearsal and
two clean builds passed within the scopes recorded above. No scale benchmark or
new exact-release image audit was run. Owned test labs cleaned up; commands.txt
remains untouched. No source push, account operation or public deployment.

This completes the first cleanup batch. Deeper Lua/controller reorganization is
still a separate unimplemented follow-up, with atomic transitions, cancellation
races and startup-failure expiry to preserve. Actual hosting checks await the
previously requested VPS/SSH/Cloudflare handoff; provider identity decisions stand.

## September 14 — behavior-preserving cleanup, second batch

The user approved the follow-ups only if functionality/usability stay unchanged.
Added three Chromium regressions before production edits: startup-failure expiry,
cancellation during a delayed arrival challenge, and participation without WebGL;
also pinned destination/preview accessibility and map removal. Baseline: 36 passed
(`reports/local/cleanup2-browser-baseline`). Private-map creation/decoration now
share a helper; public cell maps retain their separate renderer. Initial map run:
35 passed, one existing injected service-worker push test timed out; three focused
repeats passed, followed by 36/36 in the full repeat
(`reports/local/cleanup2-map-confirm`). TypeScript and diff checks passed.
Controller/Lua work remains in progress. No remote push or deployment occurred.

The session controller extraction passed TypeScript checking and all 36 Chromium
journeys (`reports/local/cleanup2-controller`). Defaults, Albanian copy, markup,
CSS, map options, one-second scheduling, poll jitter and cancellation/expiry
ordering are preserved. Full delayed-publication verification is running; Lua
formatting and a test-only planner replay remain to be committed separately.

All 34 embedded Lua fragments were expanded into readable layout with StyLua's AST
verification and a second 6,414-token comparison. Non-Lua Go source is byte-for-byte
unchanged. Added docs/TRANSACTIONS.md with source, guard/lifecycle and test links.
Real Valkey race tests passed before and after (`make test-store`); the fixed-clock
simulation accepted 40/40 credentials, rejected zero, verified expiry for 34, and
passed both real simulation-store and HTTP integration packages
(`reports/local/cleanup2-simulation`). `make test` passed race, simulation-build,
vet and TypeScript checks. These layout changes do not change atomic boundaries,
configuration, real/simulation clocks or script assembly.

Added `make test-planner-replay` and CI wiring: 3,000 synthetic participants, the
full 6,491-crossroad map, current 100 m configuration, eight time snapshots and a
reviewed decision fixture. It passed in 33.35 seconds, including shuffled-input
replays, founder reachability/deadline and decline-exclusion checks. Fixed labels
exist only in the test file; production and normal simulation randomness remain
unchanged. This is planner regression evidence, not a deterministic full API
simulation or a new capacity claim. Python guards: 25 passed; formatting check:
5 files unchanged. The normal-clock full browser journey is still running.

Second batch completed: the normal-clock browser journey passed in 15.2 minutes,
including real publication, cell-only map, private preview and late joining.
Private monitoring recorded no alerts. The 3,000-person API population run and
checker passed: 3,000 accepted, 1,923 invited, 1,718 chose to go, 1,119 distinct
credentials arrived; 456 cancelled and 2,544 verified expired. The local
production-image rehearsal passed in 85.16 seconds with two API replicas, one
worker and zero transient proxy failures. Two clean builds of 05b2b68 matched all
compared bytes. Four code/test commits plus this evidence document complete the
conditional cleanup scope. Details and the isolated push-test failure are in
reports/auditability-followup-2026-09-14.md. No public deployment, remote Git push,
new product decision, new capacity claim or fresh exact-release security audit.
Next action is the pending VPS/access preparation described in the summary above.


## September 14 — lifecycle questions and static review

Reviewed current configuration, admission/activation/arrival transactions, private
and public UI, expiry cleanup and existing tests. Answers and candidate features
are in docs/LIFECYCLE_REVIEW.md. Late joining and post-expiry arrival reconfirmation
are implemented; willingness extension, automatic gathering extension and active
location changes are not. No runtime code/configuration changes or new test runs
in this review. The expired gathering-index and stale-reference retention findings
need targeted keeper/churn tests and a fix; do not infer complete physical cleanup
from the earlier passing logical-expiry journeys. Extending/moving gatherings
would change agreed frozen-lifetime/destination policy and needs explicit design.

## September 14 — lifecycle priorities implementation (in progress)

User authorized bounded expired-metadata cleanup, clear ended public presentation
and explicit early presence renewal. No willingness/gathering lifetime extension
or active destination movement is included. Toolchain doctor passes after sourcing
scripts/env.sh (the initial unsourced shell did not expose installed tools).

Cleanup now prunes expired gathering-index members in bounded batches alongside
willingness expiry, and includes gathering backlog in private cleanup-lag metrics.
Expired session associations/nonces are cleared on own-session reads and when the
existing bounded matching snapshots encounter them. No new participant index or
persistent scan cursor is introduced; old records are covered by reconciliation.
The mutation rereads current state, preserving any newer assignment and original
willingness expiry. Tests cover repeated expiry with a live keeper, stale links
without client polling, and delayed cleanup after reassignment. `make test-store`
passed the real Valkey/race store, HTTP, notification and worker packages.


## Lifecycle presence and presentation implementation — 2026-09-14

Implemented schema 9 early presence renewal using mode/member-bound challenges,
the existing one-shot coarse device fix and rate budgets. Renewal atomically
updates one uninterrupted contribution and does not extend willingness/gathering
expiry. Ended public cards retain labeled historical buckets until snapshot
expiry; ended gatherings leave the map layer without hiding another live gathering
in the same cell. Product, API, storage, configuration and threat docs now describe
these boundaries. No credentials, push, deployment or new dependency was required.

Validation completed so far:

- `make test-store`: disposable Valkey race tests passed, including renewal replay,
  one-count/stability preservation, early rejection, expired challenges, missing
  contributions, retraction/cancellation and the cleanup keeper/churn regressions.
- `OUTPUT=reports/local/lifecycle-simulation make simulate`: passed the 40-person
  scenario plus fixed-clock store/HTTP integration, including the new exact renewal
  window, wrong-location/body rejection, endpoint separation and deadline cap.
- `make test-browser OUTPUT=reports/local/lifecycle-browser-2`: 39 Chromium tests
  passed (1.1 minutes). The first run passed 37/38; the new test's native location
  timestamp was real while its browser clock advanced 13 minutes. The test adapter
  now aligns its timestamp to the simulated clock; production validation unchanged.
- `make test-planner-replay`: passed (34.02 seconds). The first comparison detected
  the expected schema/config hash change. All other fixture fields, inputs and
  eight matching checkpoints were equal; only that reviewed hash was updated.
- `make test`: passed Go race/unit tests, simulation-build tests, vet and TypeScript;
  environment-gated integration checks are counted only under their commands above.
- Changed arrival Lua fragments passed StyLua 2.3.0 syntax/AST verification and
  token comparison during formatting; real-store tests passed afterward.

Completed in `be1077c` (cleanup) and `5ced2db` (renewal/ended views). Final response
simulation and checker passed: all 3,000 credentials accepted, 1,929 distinct
credentials invited, 1,743 chose to go, 1,179 arrived; 53 gatherings/36 confirmed,
24 joins after JEMI KËTU, 456 cancellations + 2,544 verified expiries. Driver wall
time was 230.7 seconds. The unchanged driver covers ordinary arrival/re-arrival;
new tests above cover explicit early renewal.

The production-image/two-API/Caddy rehearsal passed in 75.03 seconds excluding
builds. It observed one transient proxy 5xx during forced replica termination and
then recovered. Exact scope, first-run test fixes, hashes and remaining limits are
in [the lifecycle report](reports/lifecycle-2026-09-14.md). No public activation,
new capacity measurement or independent release audit occurred.

Next: wait for the delivered VPS connection details and privately arranged access;
prepare and audit a fresh exact release before public activation. Local work in
this approved lifecycle batch is complete. Preserve untracked commands.txt.


## Albanian About section — 2026-09-14

Added a top-navigation “Rreth nesh” anchor and a warm cream/rose section on the
existing page. It explicitly supports the Flamingo protests, values continuing
every day at the usual hour in front of Kryeministria, and aims to complement that
with continuous peaceful participation across Albania. It identifies this version
as Tirana-only, explains the voluntary willingness/invitation/arrival flow and
leaves subsequent peaceful protest decisions to participants. No exact protest
hour, official organizer identity or nationwide availability was invented.

Privacy cards describe no registration/identity fields, locally coarsened location,
automatically expiring coordination data, no personal participation history,
suppressed/delayed public cells and open-source inspection. An expandable explanation
covers temporary retention, rate/replay controls, inference, device/Sybil limits,
network/server-operator exposure, optional push and absence of independent audit.
It does not claim zero retention, absolute anonymity or proof of remote behavior.
The configured origin's public GitHub repository was verified accessible without
authentication and linked with explicit new-tab/no-referrer behavior. No automatic
external asset/request, tracking, dependency, runtime setting or backend changed.

Validation:

- Sourced `scripts/env.sh`; `bash scripts/doctor.sh` passed required tool checks;
  optional gh remains absent. No tools needed installation.
- `npm --prefix web run build`: TypeScript and production Vite build passed.
  The existing large MapLibre bundle warning remains; no new dependency was added.
- `make test-browser OUTPUT=reports/local/about-browser`: **41 Chromium tests
  passed in 1.1 minutes**, including existing real API participation, arrival,
  renewal, map, retry/cancellation and optional-push emulation checks.
- New tests confirm reading/navigation creates no participation or location request,
  preserves an active capability/deadline, supports keyboard navigation/disclosure,
  and remains readable when the API fails. Automated WCAG A/AA scans and horizontal
  overflow checks pass at 390 px and 200% text, including expanded privacy details.
- Focused About accessibility/visual test rerun passed after making screenshot
  paths relative to the test module. Desktop, mobile and full-section screenshots
  are in ignored `reports/local/about-{desktop,mobile,section}.png`; inspected visually.
- No backend/load/deployment audit was rerun for this static HTML/CSS change.
  Browser GPS and push are emulated; these checks do not certify real devices or
  a public deployment. Existing host/edge and exact-release audit gates remain.

Requirements and plan now record the approved public purpose, superseding the
purpose-copy exclusion from the earlier usability batch. README and THREAT_MODEL
are consistent. Next: prepare a fresh audited release once VPS/access is ready;
this About request is complete locally. Keep commands.txt untouched.

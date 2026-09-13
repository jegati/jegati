# Functionality and test coverage

Reviewed 2026-09-13 against REQUIREMENTS.md, the current API routes/client, configuration,
test sources and locally executed evidence. Optional future availability is excluded.
“Tested” means the checks below cover the stated behavior; it does not certify all
security properties or a public deployment. A configured field alone is not a feature.

The proposed next steps and acceptance gates are in
[ACTIVITY_IMPLEMENTATION_PLAN.md](ACTIVITY_IMPLEMENTATION_PLAN.md). The missing
features below remain unimplemented.

## The four user-requested journeys

| Journey | Verdict | Present and tested | Missing / qualification |
| --- | --- | --- | --- |
| 1. First visit → JAM GATI → nearby willingness count | **Partial** | First visit, duration/radius/device fix, JAM GATI state and expiry/cancel are covered by browser and real-store checks. | No nearby bucketed willingness statistic such as 50+ GATI appears after activation of the user's willingness. |
| 2. Nearby threshold → invitation → accept/decline | **Partial** | Compatible activation (current threshold 30, stable for 10 seconds), JEMI GATI, destination/deadline, PO, PO SHKOJ and JO TANI are covered by planner, fixed-clock API and browser tests. | Delivery is an in-page update through visible-tab polling, normally about 30–36 seconds between status polls. No OS/background push or area-follow alerts. JO TANI declines the invitation and keeps JAM GATI active; Mbyll gatishmërinë separately cancels willingness. |
| 3. Going → location-checked JAM KËTU → continuing approximate going/arrival statistics | **Partial** | Going enables JAM KËTU. Fresh device coordinates are coarsened locally; the API checks the destination cell (currently zero neighboring rings), nonce, admission and deadlines. Success records temporary presence, supports retraction and can trigger JEMI KËTU at 20 stable arrival claims. Browser, real-store and fixed-clock tests cover this. | Neither approximate counts of people going nor approximate arrival counts are exposed. The UI shows the user's own state and collective JEMI GATI/JEMI KËTU, not 50+-style totals. Presence lasts at most 15 minutes and is capped by session/gathering expiry; it is not continuous tracking or proof against spoofing. |
| 4. Tirana activity map: current gatherings/statistics or willingness heatmap | **Missing as an activity feature** | A first-party Tirana basemap and private invited-gathering destination map render and are browser-tested. | No citywide public gathering/activity layer, heatmap or bucketed statistics. The synthetic replay demonstrates simulation data only and is not available as a live product activity map. |

Source inspection: [client state/rendering](../web/src/main.ts),
[product markup](../web/index.html), [API routes](../internal/httpapi/server.go),
[arrival checks](../internal/httpapi/arrivals.go). Behavior evidence:
[browser journeys](../web/tests/willingness.spec.ts),
[continuous activation/late arrival integration](../internal/httpapi/gathering_simulation_test.go),
[store arrival tests](../internal/store/arrivals_test.go) and the population reports.

Implementing the missing statistics requires the planned public aggregate release
layer and UI, including suppression, buckets and delayed fixed releases. Going and
arrival counts must be reviewed together to avoid revealing small groups through
subtraction. Notification delivery and public discovery are separate unfinished work.
No new statistics or notification feature was implemented as part of this audit.

## Present and tested

| Requirement / behavior | Implementation | Evidence and scope |
| --- | --- | --- |
| Anonymous, temporary willingness | No account; random expiring capability; duration/radius choices; immediate neutral cancellation | Real-store creation/expiry/cancellation/retry tests; real-browser create/reload/cancel checks. Network/operator anonymity is a separate unresolved property. |
| Minimum 30-minute availability | Config validation and client choices; server deadlines | Config rejection tests, strict API validation, browser minimum-duration check |
| Device-only one-shot location | Fresh fix required; local coarsening; no manual map/address input; exact coordinates not uploaded | Browser checks for payload/storage, denied/missing location, poor/stale/out-of-area fixes and map interaction. Device origin and physical accuracy are not authenticated by the API. |
| Default 100 m cells and fractional radii | Current gati.yaml: 100 m / 50 m maximum reported error; radii 0.1/0.5/1/3 km | Config/geographic tests, running config/hash inspection and all 13 browser checks on the new grid. Short-radius eligibility is still conservative. |
| Continuous compatible matching | Reachability, remaining availability, stability deadlines and timer/reconciliation processing | Planner radius/unaligned-time tests, exhaustive small-cohort oracle, real-store reservation/activation races and fixed-clock API integration |
| Automatic nearest mapped crossroad | Imported Tirana junctions; nearest common eligible destination; fixed activated destination/deadline | Map/connectivity, nearest/tie/missing-common-destination tests, frozen-destination simulation assertions, browser destination rendering |
| JEMI GATI and decisions | Private invitation, PO, PO SHKOJ, JO TANI; decline retains willingness | Browser full flow; planner/atomic intent/fixed-clock tests; population acceptance/decline/ignore scenarios |
| Late admission, including after JEMI KËTU | Existing-session admission and atomic new-session /api/join; revalidate radius/deadlines/live state | Fixed-clock API tests and 3,000-person response simulations. Notification/map discovery is injected in simulations, not a shipped public user flow. |
| JAM KËTU / JEMI KËTU | Nonce-backed coarse arrival claim, freshness/stability, retraction and replay protection | Real-store concurrent arrival/replay/retraction tests; founding/late/wrong-cell/stability/expiry integration; browser arrival retry/retraction. Distinct credentials are not independent humans. |
| First-party map backdrop and private destination | Local road assets and own-area/destination maps | Browser rendering/network checks. This is not the public aggregate activity map. |
| Albanian minimal flow | Albanian screens, errors, controls and synthetic replays; flamingo identity | Browser language/mobile/keyboard/full-flow checks. No consent-based human usability study yet. |
| Parameterized local Tirana simulations | 3,000-person scenarios, decisions/journeys, offline/late/cancel/replay behavior, isolated real API/store, timeline exports | 18 final city-case checks, controls and exact functional repeatability; see reports/population-3000.md. The original baseline was 1 km; its 100 m comparison has the same functional settings as the new default. |

## Partial coverage and missing requirements

| Requirement | Status | What remains |
| --- | --- | --- |
| Notifications | Partial | Foreground own-status polling shows invitations and gathering state. No background push, area-follow subscriptions, large-nearby alerts or notification queue/delivery lifecycle. |
| Newly notified people joining | Partial end-to-end | Admission API is implemented/tested; the public discovery/notification/deep-link user journey is unfinished. |
| Anonymous public collective map/counts/history | Missing | No released GATI buckets, public activated/confirmed gathering layer, suppressed/delayed aggregate snapshots or protected statistics over time. Public configuration fields reserve these behaviors only. |
| Architectural privacy | Partial; known unmet requirement | Coarse inputs, temporary hashes, TTL checks, private nonpersistent store and no individual public views have evidence. A prior-config colluding-input probe reconstructs a target cell. No complete cross-surface/differencing or privileged-operator protection; the new 100 m default is not an anonymity proof. |
| Abuse resistance | Partial | Request/body/create/global/arrival limits, replay controls, atomic transitions and rate-secret rollover regression exist. Distributed Sybils, fabricated device claims, effective anomaly detection and upstream/edge DoS protection remain unresolved. Shared-network exclusions are measured, not solved. |
| Independent arrivals / accurate physical presence | Unmet as a guarantee | One credential is not one human, and client location can be spoofed. Existing arrival checks establish bounded accepted coarse claims only. |
| 100k+ capacity and spikes | Unverified | No 100k real-time benchmark, reference capacity result or complete worker failover/overload tests. 3,000-person accelerated simulations do not establish this. |
| Simple local maintenance | Partial | Pinned toolchain, Compose, ephemeral-store settings, health endpoint and local checks exist. Full operational monitoring, production deployment/rollback, host hardening and nonparticipant backup/restore evidence remain later work. |
| Open-source auditability | Partial | AGPL source, schema, infrastructure, tests, threat/data-flow docs, config/map hashes and reviewed reports exist. Independent security/privacy audit has not run. |
| Actual deployment verification | Missing | No published production release/deployment, signed release/SBOM/independent build comparison or privileged host verification. A config hash is not remote attestation. |
| PWA packaging/offline installation | Missing | Browser app exists; service-worker/offline shell/installability work remains. Native apps are outside the MVP. |
| All configurable knobs affect behavior | Partial | Matching/arrival/device settings are active. Public notification/aggregate settings are pending, and matching.intersection_index_batch_size is validated but unused. See MATCHING_PARAMETERS.md. |

## Validation refreshed for the 100 m default

- `make config-check`; simulation-build validation of the population profile.
- `make verify-local`: unit/race and simulation-tag tests, vet, TypeScript, builds,
  config checks, Compose parsing and HTTP smoke. Integration tests requiring a
  real store are not counted as passed just because this target skips them.
- `make test-store`: actual temporary Valkey tests for lifecycle, concurrency,
  arrival replay, ACLs, HTTP validation and rate-secret rotation.
- `make simulate SCENARIO=tirana-evening`: legacy willingness driver plus separate
  real-store fixed-clock activation/late-admission/arrival integration, on its
  explicit smaller regression profile.
- `make test-browser`: all 13 checks passed against the real local API using the
  new 100 m default. Device geolocation is mocked; no real GPS is requested.

The previously completed 18-run population suite and the 100 m comparison were
reviewed, not rerun for this config-only change. Their frozen inputs/hashes and
limitations remain in reports/population-3000.md. Remote CI and a human study have
not run. Public release readiness is not established by the passing local checks.

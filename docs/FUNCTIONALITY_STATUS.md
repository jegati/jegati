# Functionality and test coverage

Reviewed 2026-09-13 against REQUIREMENTS.md, the current API routes/client, configuration,
test sources and locally executed evidence. Optional future availability is excluded.
“Tested” means the checks below cover the stated behavior; it does not certify all
security properties or a public deployment. A configured field alone is not a feature.

Implementation order and remaining work are in
[ACTIVITY_IMPLEMENTATION_PLAN.md](ACTIVITY_IMPLEMENTATION_PLAN.md).

## The four user-requested journeys

| Journey | Verdict | Present and tested | Missing / qualification |
| --- | --- | --- | --- |
| 1. First visit → JAM GATI → nearby willingness count | Implemented locally | Device-only willingness, cancellation/expiry and a nearby published-area bucket. Real-store capture/API tests plus browser presentation with fixed public fixtures. | Public figures describe a delayed observation, not immediate occupancy. Below-minimum values remain unavailable, never zero. |
| 2. Nearby threshold → invitation → accept/decline | Partial | Continuous private activation, foreground JEMI GATI, PO, PO SHKOJ / JO TANI; browser and fixed-clock API tests. | No background push, follows or server-side notification outbox yet. Push stays optional/off by default; JO TANI preserves willingness. |
| 3. Going → location-checked JAM KËTU → approximate going/arrival statistics | Implemented locally | Real API/browser arrival, nonce retry/retraction and expiry; same published going/here cards before/after arrival. Real fixed-clock publisher tests verify public suppression independently of private JEMI KËTU. | Presentation tests use fixed public fixtures alongside real arrival tests. No live-device presence proof; fresh claims last at most 15 minutes. Full browser-to-real-publisher test under the normal 5–10 minute delay remains a later integration extension. |
| 4. Tirana activity map | Implemented locally | Always-visible first-party city map; willingness polygons and gathering entries in the same public cells; statistics/list, explicit public-map join and capability-preserving retry. | Exact crossroads appear only in private invitations/admission. No daily-history summaries yet. Population discovery still uses its existing synthetic scheduling; it does not yet consume the new map API. |

Evidence: [publication policy](../internal/activity/release_test.go),
[real-store publication](../internal/store/activity_test.go),
[public API/cache](../internal/httpapi/activity_test.go),
[browser statistics/map fixtures](../web/tests/activity.spec.ts),
[real browser participation](../web/tests/willingness.spec.ts), and
[fixed-clock real publisher/arrival flow](../internal/httpapi/gathering_simulation_test.go).
Decision 0007 accepts inferred information while preserving direct-data boundaries,
cell-only public geography, small-count suppression and expiry. This is not formal
anonymity, a 100k benchmark, or a production deployment certification.

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
| Late admission, including after JEMI KËTU | Existing-session admission and atomic new-session /api/join; revalidate radius/deadlines/live state | Fixed-clock API tests and 3,000-person response simulations. Public-map joining is now shipped locally; population discovery remains injected in simulations. |
| JAM KËTU / JEMI KËTU | Nonce-backed coarse arrival claim, freshness/stability, retraction and replay protection | Real-store concurrent arrival/replay/retraction tests; founding/late/wrong-cell/stability/expiry integration; browser arrival retry/retraction. Distinct credentials are not independent humans. |
| First-party map backdrop and private destination | Local road assets and own-area/destination maps | Browser rendering/network checks. Public aggregate layers now share the city map; the private destination remains separate. |
| Albanian minimal flow | Albanian screens, errors, controls and synthetic replays; flamingo identity | Browser language/mobile/keyboard/full-flow checks. No consent-based human usability study yet. |
| Parameterized local Tirana simulations | 3,000-person scenarios, decisions/journeys, offline/late/cancel/replay behavior, isolated real API/store, timeline exports | 18 final city-case checks, controls and exact functional repeatability; see reports/population-3000.md. The original baseline was 1 km; its 100 m comparison has the same functional settings as the new default. |

## Partial coverage and missing requirements

| Requirement | Status | What remains |
| --- | --- | --- |
| Notifications | Partial | Foreground own-status polling shows invitations and gathering state. No background push, area-follow subscriptions, large-nearby alerts or notification queue/delivery lifecycle. |
| Newly notified people joining | Partial end-to-end | Public-map discovery/join is implemented; background notification/follow discovery and deep links remain. |
| Public collective map/counts/history | Partial | Delayed bucketed snapshots, nearby/going/here cards and cell-only map are implemented. Daily aggregate history is missing; inferred information is explicitly accepted. |
| Architectural privacy | Partial; known unmet requirement | Coarse inputs, temporary hashes, TTL checks, private nonpersistent store and no individual public views have evidence. A prior-config colluding-input probe reconstructs a target cell. No complete cross-surface/differencing or privileged-operator protection; the new 100 m default is not an anonymity proof. |
| Abuse resistance | Partial | Request/body/create/global/arrival limits, replay controls, atomic transitions and rate-secret rollover regression exist. Distributed Sybils, fabricated device claims, effective anomaly detection and upstream/edge DoS protection remain unresolved. Shared-network exclusions are measured, not solved. |
| Independent arrivals / accurate physical presence | Unmet as a guarantee | One credential is not one human, and client location can be spoofed. Existing arrival checks establish bounded accepted coarse claims only. |
| 100k+ capacity and spikes | Unverified | No 100k real-time benchmark, reference capacity result or complete worker failover/overload tests. 3,000-person accelerated simulations do not establish this. |
| Simple local maintenance | Partial | Pinned toolchain, Compose, ephemeral-store settings, health endpoint and local checks exist. Full operational monitoring, production deployment/rollback, host hardening and nonparticipant backup/restore evidence remain later work. |
| Open-source auditability | Partial | AGPL source, schema, infrastructure, tests, threat/data-flow docs, config/map hashes and reviewed reports exist. Independent security/privacy audit has not run. |
| Actual deployment verification | Missing | No published production release/deployment, signed release/SBOM/independent build comparison or privileged host verification. A config hash is not remote attestation. |
| PWA packaging/offline installation | Missing | Browser app exists; service-worker/offline shell/installability work remains. Native apps are outside the MVP. |
| All configurable knobs affect behavior | Partial | Matching/arrival/device settings are active. Public snapshots are active; notification delivery and daily-summary settings are pending, and matching.intersection_index_batch_size is validated but unused. See MATCHING_PARAMETERS.md. |

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

## Activity slice verification

`make verify-local`, `make test-store` and `make simulate SCENARIO=tirana-evening`
passed after the publisher/API implementation. Browser suite: 17 checks passed,
including four explicit public-response fixtures plus the real participation flow.
Fixtures are deliberately distinguished from real publisher evidence above. A
focused activity recheck verifies the final Albanian time formatting and map-ready
screenshot. The 3,000-person publisher-enabled seed-42 run and report checks passed (see
reports/activity-3000.md); no full 18-case rerun or 100k result is claimed.

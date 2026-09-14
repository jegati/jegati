# Auditing GATI

Start with [current functionality](FUNCTIONALITY_STATUS.md) and the current summary
in [PROGRESS](PROGRESS.md). Older milestone entries are historical evidence, not the
current implementation contract. [REQUIREMENTS](REQUIREMENTS.md) describes agreed
behavior; [API](API.md), [STORAGE](STORAGE.md) and [THREAT_MODEL](THREAT_MODEL.md)
describe the observable surfaces, data lifetimes and remaining limitations.
The [first cleanup report](reports/auditability-2026-09-14.md) and
[follow-up report](reports/auditability-followup-2026-09-14.md) record the refactors
and exact scope of their local regression evidence.

## Follow a request through the code

| Flow | Entry and checks | State/data owner | Regression evidence |
| --- | --- | --- | --- |
| Device location → willingness | [client location](../web/src/location.ts), [session storage](../web/src/session.ts), [HTTP auth/limits](../internal/httpapi/signals.go) and [strict object parsing](../internal/httpapi/strict_json.go) | [create/status/cancel](../internal/store/store.go); bounded cell/expiry indexes and capability replay tombstone | [HTTP validation](../internal/httpapi/signals_test.go), [strict-body matrix/fuzz](../internal/httpapi/strict_json_test.go), [browser willingness](../web/tests/willingness.spec.ts) |
| Continuous matching → invitation | [worker](../internal/worker/engine.go), [planner](../internal/matching/planner.go), [map reachability](../internal/geography/index.go) | [atomic reservations/activation](../internal/store/gatherings.go), [incremental working set](../internal/worker/population.go) | [reservation races](../internal/store/gatherings_test.go), [full fixed-clock journey](../internal/httpapi/gathering_simulation_test.go) |
| Preview, going, decline, late join | [preview](../internal/httpapi/preview.go) and [invitation/admission](../internal/httpapi/invitations.go); explicit capability, coarse eligibility and live deadlines | [intent/join transactions](../internal/store/intent.go); preview creates no participation record | [fixed-clock journey](../internal/httpapi/gathering_simulation_test.go), [browser activity](../web/tests/activity.spec.ts) |
| Arrival, early presence renewal, replay and retraction | [arrival HTTP](../internal/httpapi/arrivals.go); capability, one-use nonce, coarse area and expiry | [atomic arrival transitions](../internal/store/arrivals.go); bounded nonce and arrival-index lifetimes | [arrival state model/races](../internal/store/arrivals_test.go), [renewal invariants](../internal/store/arrival_renewal_test.go), [renewal HTTP boundaries](../internal/httpapi/arrival_renewal_simulation_test.go), [fixed-clock journey](../internal/httpapi/gathering_simulation_test.go) |
| Own-session/preview JSON | [explicit response allowlists](../internal/httpapi/responses.go); authenticated/no-store | Only enumerated fields leave the storage boundary; private invitations include map landmarks | [nested response contracts](../internal/httpapi/response_contract_test.go), assertions throughout the fixed-clock journey |
| Public map/statistics | [public activity API](../internal/httpapi/activity.go); fixed delayed releases, no arbitrary query | [publication policy](../internal/activity/release.go), [publisher](../internal/worker/activity.go), [bounded aggregate storage](../internal/store/activity.go) | [suppression/bucket tests](../internal/activity/release_test.go), [public cache/query tests](../internal/httpapi/activity_test.go), [real-publisher browser journey](../web/tests/full.spec.ts) |
| Optional push | [client opt-in/resume](../web/src/push.ts), [HTTP registration](../internal/httpapi/push.go) | [temporary subscriptions/outbox](../internal/store/push.go), [delivery service](../internal/notification/service.go); see [notification lifecycle](NOTIFICATIONS.md) | [browser push](../web/tests/push.spec.ts), real-store notification tests in `make test-store`; external provider/device checks remain |
| Deployment and operations | [trusted proxy](../internal/httpapi/proxy.go), [production Compose](../compose.production.yaml), [release guards](../scripts/release.py) | Memory-only Valkey, restricted service roles, bounded [private monitoring](MONITORING.md) | [release guard tests](../scripts/test_release.py), [deployment rehearsal](../scripts/deployment-lab.py), [runtime audit](RUNTIME_SECURITY.md) |

The browser entry point [main.ts](../web/src/main.ts) owns DOM rendering, startup,
map/preview cleanup, optional push wiring and the existing timer/event scheduling.
[session-controller.ts](../web/src/session-controller.ts) owns the mutable session
state, authenticated requests, retry backoff, intent/arrival transitions and deadline
checks. [session.ts](../web/src/session.ts) owns credential persistence. The shared
[private-map helper](../web/src/destination-map.ts) keeps preview/destination styles
consistent; public activity rendering remains cell-only. The controller's cleanup
hooks preserve push/preview cancellation before credentials are cleared and map/UI
cleanup afterward. Startup-failure expiry still runs independently of successful
initialization; cancellation remains available during other in-flight operations.

The [transaction catalog](TRANSACTIONS.md) maps each embedded Lua transaction to
its guards, lifecycle and integration tests, including shared fragments and clocks.

For each flow, inspect authorization, the whole atomic transition, server deadlines,
index cleanup, retries and every response. Storage TTL alone does not expire sorted
set members, and the matching owner's private memory cache has a separate lifecycle.
Public cell aggregates and private actionable destinations are different surfaces;
review their combined inference exposure using decisions 0004 and 0007.

## Run the evidence gates

Follow [development prerequisites](DEVELOPMENT.md), then:

```sh
make audit-local
make browser-install
make audit-journeys
```

`audit-local` runs native race/simulation-build/vet/type/build/config/smoke checks,
Python guard tests and **real disposable Valkey integration tests**. Plain `make test`
still skips environment-gated integrations; a skipped check is not evidence of a
working store. The CI workflow uses `audit-local` alongside its existing fuzz,
simulation, runtime and browser/recovery checks. Remote CI execution is not implied
by a passing local run.

`audit-journeys` sequentially runs the Chromium participation suite, the normal-clock
browser→matching→arrival→delayed-publication journey, and the production-image
rehearsal with two API replicas. The real-publisher journey may take 20 minutes.
These commands use owned synthetic labs; they do not deploy publicly. Each command
stops on failure and preserves its local report/logs. See [AUTOMATED_TESTING](AUTOMATED_TESTING.md)
for individual commands, fixture boundaries, fuzzing, recovery, other browsers,
capacity checks and report locations. Release-helper formatting has its own
`make format-python-check` prerequisite described in DEVELOPMENT.

Capacity is a separate gate: use the recorded hardware/configuration and load
commands when changing storage/worker execution. A 3,000-person accelerated scenario
is a functional check, not a 100k capacity test. Compare deterministic behavior,
not random tokens or wall-clock duration, and record rejected traffic as well.

Before publishing, follow [DEPLOYMENT](DEPLOYMENT.md) for a fresh exact-release
image audit, immutable artifact verification, compatible rollback, served-asset
comparison and privileged host/edge inspection. Two clean builds on one host are
not independent attestation. Nothing here proves absence of hidden remote logging,
verified physical presence, unique humans or immunity to inference/Sybil attacks.

## Simplification boundaries

Keep formatter-only changes separate from behavior changes. Request schemas and
response allowlists belong at the HTTP boundary; private storage data must not
become a public contract by embedding a storage record. Preserve atomic Lua checks,
production/simulation isolation, deadlines, client cancellation races and startup
failure cleanup when reorganizing those modules. The browser controller extraction is covered by the same browser journeys and
explicit startup/cancellation regressions. Lua layout changes have a separate
real-store and fixed-clock regression gate.

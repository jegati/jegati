# Behavior-preserving cleanup follow-up — 2026-09-14

Scope: the user authorized cleanup only while preserving functionality and
usability. This batch reorganizes existing code and adds regression evidence; it
introduces no product feature, changed matching/arrival/privacy setting or runtime
dependency. The earlier batch is in [auditability-2026-09-14](auditability-2026-09-14.md).

## Changes

| Commit | Change and preserved boundary |
| --- | --- |
| `fe8a5ea` | Shared private destination/preview map creation and decoration. Constructor assignment, cleanup after partial failure, map options, labels and preview/destination accessibility differences are preserved; public cell rendering stays separate. Added three browser regressions and map lifecycle assertions before the refactors. |
| `36616d6` | Extracted own-session state, authenticated requests, retry/intent/arrival/deadline transitions into `web/src/session-controller.ts`. The entry point keeps rendering, map/preview cleanup, push wiring, startup and event/timer scheduling. |
| `f600502` | Expanded 34 embedded Lua fragments into readable layout and added the [transaction catalog](../TRANSACTIONS.md). Kept atomic execution, argument order, return values, shared helpers, TTL operations and clock/namespace selection. |
| `05b2b68` | Added `make test-planner-replay`, reviewed JSON fixture and CI wiring. Fixed test-only labels support exact 3,000-participant planner comparisons; production and normal simulation randomness are unchanged. |

No changes to application configuration, HTML/CSS, device location validation,
session persistence, push implementation/worker, production matching/worker/map
code, dependency manifests or lockfiles. The UI remains Albanian. The private map
helper does not change the public map's cell-only publication policy.

## Equivalence and regression evidence

- **Browser baseline before refactoring: 36/36 Chromium tests passed.** Added checks
  cover expiry when startup fails, cancellation during an in-flight nonce request,
  no subsequent location/arrival submission after that cancellation, WebGL failure
  without blocking participation, preview map removal and accessible map labels.
- **Private-map repeat: 36/36 passed (58.0 seconds).** The initial map run passed
  35 and timed out in the existing injected service-worker push notification test.
  Three focused repetitions passed, then the full suite passed. No push code or
  timeout was changed to obtain a pass. Treat the isolated failure as an observed
  intermittent browser/provider-emulation issue, not evidence of reliable real
  provider delivery; its root cause was not established.
- **Session-controller suite: 36/36 passed (about one minute).** Includes ordinary
  participation, small-screen/keyboard accessibility, retry/cancellation races,
  preview/late join, arrival/retraction, public buckets and optional push fixtures.
- **Additional browser syntax comparison:** 13 transferred/view routines have
  identical normalized syntax after replacing state-property references with the
  former local names and inlining the accepted-signal hook. Startup plus the moved
  tick body, and end plus its two view cleanup hooks, also match after reassembly.
  TypeScript 5.9.3's parser/printer was used only as an isolated audit tool; the
  application still builds with its pinned TypeScript 7.0.2. This comparison is
  supplementary evidence, not a proof covering every browser/device.
- **Lua comparison:** StyLua 2.3.0 verified the parsed fragments, followed by a
  second comparison of 6,414 executable tokens. Allowed layout differences are
  whitespace/comments, optional semicolons/trailing table commas and equivalent
  unescaped quote styles. Go source outside the Lua strings is byte-for-byte
  unchanged. See the transaction catalog for formatter settings and archive hash.
- **`make test-store`: passed before and after Lua formatting**, using disposable
  restricted Valkey and the race detector for store, HTTP, notification and worker
  integrations. **`make simulate`: passed**, including simulation-store/HTTP
  checks; 40 accepted, zero rejected, 34 expiries verified in the scenario.
- **`make test`: passed** Go race tests, simulation-build tests, vet and TypeScript.
  **Python guards: 25 passed.** **Python format check: 5 files unchanged.** Existing
  environment-gated tests skipped in plain Go runs are covered only where their
  separate commands are listed; they are not counted as extra passes.
- **Planner replay: passed in 33.35 seconds.** Seed 42, 3,000 participants, all
  6,491 mapped crossroads, current 100 m config, eight time snapshots, shuffled
  input replays, founder reachability/deadlines and exclusion after declining.
  This fixture is bounded to three proposals per snapshot, uses a synthetic
  halfway cancellation rule and does not send requests or model arrivals. Its
  counts are regression checkpoints, not a predicted end-to-end funnel.
- **Full 3,000-person API population simulation: passed**, followed by
  `make check-population`. All 3,000 credentials accepted; 1,923 distinct credentials
  invited, 1,718 chose to go and 1,119 arrived. There were 1,586 accepted arrival
  operations (distinct from people), 189 verified replays, 46 rejected arrival
  attempts, 46 observed gatherings/32 confirmed, and 22 joins after JEMI KËTU.
  456 cancellations plus 2,544 verified expiries account for all 3,000 accepted
  credentials. The driver took 215.6 seconds, excluding startup. Cryptographic
  tie-breakers remain random, so this funnel is not an exact equality assertion
  against another seeded API run or a performance comparison.
- **Production-image rehearsal: passed in 85.16 seconds.** Built client through
  Caddy/CSP, two API replicas plus one worker, actual container/store controls,
  admission limits across replicas, concurrent arrival replay and surviving API
  after one replica stops. Zero transient proxy 5xx responses during this test.
- **Normal-clock full browser journey: passed in 15.2 minutes.** Real API/store,
  matching, arrival and delayed publication, aggregate map, private preview and
  late joining from a new browser; no API-response mocks or shortened product
  timings. Private monitoring ran throughout and recorded no alerts.
- **Two clean builds of `05b2b68`: byte-identical.** Includes API binaries, browser
  assets, Caddy and cloudflared, with separate source exports/dependency installs
  and compilation caches on this same workstation. This is not an independent
  build, host attestation or proof of remote operator behavior.

## Local evidence paths

Under `reports/local/`: `cleanup2-browser-baseline`, `cleanup2-private-map`
(initial intermittent failure), `cleanup2-map-push-repeat2`, `cleanup2-map-confirm`,
`cleanup2-controller`, `cleanup2-simulation`, `cleanup2-population`, `cleanup2-deployment`,
`cleanup2-reproduction`, `cleanup2-full-journey`. Go/type/Python and the planner target return their result
to the command line; the reviewed planner fixture is checked in under
`internal/simulation/testdata/`.

## Limitations and next step

No public deployment, remote Git push or independent audit occurred. The localhost
production rehearsal does not test the pending OVH VPS, Cloudflare, external TLS,
real phone GPS or push-provider delivery. New source requires a fresh exact-release
runtime/image audit before public activation; the older `06fa798` audit is not
transferred to this revision. Existing product gaps and privacy limitations in
PROGRESS/THREAT_MODEL remain. This batch makes no new capacity or security claim.

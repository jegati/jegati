# Gathering lifecycle priorities — 2026-09-14

The approved priorities are implemented locally:

- `be1077c`: bounded pruning of expired gathering index entries; own-session reads
  and existing matching snapshots remove stale gathering links/challenges without
  renewing session TTLs or erasing a newer assignment. Cleanup lag observes both
  gathering and signal expiry indexes.
- `5ced2db`: explicit **JAM ENDE KËTU** presence renewal and ended public views.
  Schema 9 adds `arrivals.renewal_window_seconds` (120 by default, positive and at
  most 300, strictly shorter than freshness). A fresh challenge and one-shot coarse
  device fix refresh one contribution within the original session/gathering limits.
  Ended cards retain labeled historical buckets until snapshot expiry; ended
  gatherings leave the map layer while another live gathering can retain the cell.

No automatic renewal, willingness/gathering extension, destination movement,
location history, permanent identifier or additional runtime service was added.
The complete boundary is in [decision 0014](../decisions/0014-bounded-presence-renewal.md).

## Executed regression checks

| Command | Result and scope |
| --- | --- |
| `make test` | Passed Go race/unit tests, simulation-build tests, vet and TypeScript. Environment-gated integrations are counted only under their separate commands. |
| `make test-store` | Passed disposable restricted Valkey integrations with the race detector. New keeper/churn, stale/new assignment, renewal concurrency, one-count/cohort preservation, mode separation, retraction/cancellation, missing-member and expired-challenge tests passed. Final package durations: store 4.56 s, HTTP 1.15 s, notification 1.79 s, worker 1.14 s. |
| `OUTPUT=reports/local/lifecycle-simulation make simulate` | Passed: 40 accepted synthetic credentials, zero rejected, 34 expiries verified, plus fixed-clock store/HTTP tests. New API test verifies the exact two-minute renewal boundary, wrong-cell/extra-coordinate rejection, private response allowlists/cache headers, replay and the original deadline cap, including renewal after the late-admission cutoff. |
| `make test-browser OUTPUT=reports/local/lifecycle-browser-2` | 39 Chromium tests passed in 1.1 minutes. Includes explicit fresh-fix renewal/retry, no automatic renewal, cancellation during a pending renewal challenge, original session deadline, ended public cards and shared-cell map removal. Public renderer fixtures and accelerated-clock renewal are separate from the existing real API browser journey. |
| `make test-planner-replay` | Passed in 34.02 seconds: 3,000 seeded participants, 100 m cells, all 6,491 mapped crossroads, eight reviewed checkpoints and shuffled input replay. |
| `make simulate-population OUTPUT=reports/local/lifecycle-population` followed by `make check-population REPORT=reports/local/lifecycle-population/report.json` | Passed the real isolated API/store response simulation and expiry/funnel/config checker. See counts below. |
| `make test-deployment OUTPUT=reports/local/lifecycle-deployment` | Production images built; local Caddy/client, two API replicas and one worker, actual container/store controls, private monitoring, replay and shared admission budgets passed. Forced replica termination caused one transient proxy 5xx, followed by successful recovery. Rehearsal took 75.03 seconds excluding builds. |

All existing tools were available after sourcing `scripts/env.sh`; no installation
was necessary. Changed Lua fragments also passed StyLua 2.3.0 AST verification and
executable-token comparison during formatting, followed by real-store tests.

The first new browser test run passed 37/38 and failed renewal because native
geolocation timestamps stayed on real time while Playwright advanced 13 minutes.
The test adapter now aligns its callback timestamp with the accelerated clock;
production location validation was unchanged. The final 39-test run passed.
The first planner replay detected only the intended schema/config hash change;
all other fixture fields and decisions matched. Only its configuration hash was
updated before the passing rerun. Config tests also reject zero, excessive and
whole-freshness renewal windows.

## Population results

Seed 42, production functional defaults with only the isolated simulation profile
substituted, 180 virtual minutes, 230.7 driver wall seconds excluding startup:

- 3,000 generated and accepted credentials; 1,929 distinct credentials invited.
- 1,743 distinct credentials chose to go; 1,179 distinct credentials arrived.
- 1,748 accepted arrival operations, 215 verified arrival retries, 56 rejected
  arrival attempts and 170 arrival retractions.
- 53 gatherings observed, 36 confirmed; 24 joins after JEMI KËTU.
- 456 cancellations + 2,544 verified expiries account for all accepted credentials.

The unchanged population driver exercises normal arrival/re-arrival and cleanup;
early renewal is covered by the new store, HTTP and browser tests above. Crypto
IDs/tie-breakers remain random, so this funnel is not an equality assertion against
prior seeded API runs, a prediction of human behavior or a speedup measurement.

Local detailed artifacts are under `reports/local/lifecycle-*` and remain ignored.
The production-config hash is
`3f16cecde7a990dfd9f75340f29180fe110f4f0032fbe373d0a551e7e0137b14`;
the simulation-profile hash is
`4998437927e2da3a9388c5cb228b11ce7bb3d828a1cd5215b913f17c8358b40b`.
Runs began on the working tree later committed as `5ced2db`; the population report
honestly records base revision `be1077c` and `source_dirty: true`, with binary/map
hashes. It is not a clean-release artifact attestation.

## Remaining limits

These checks do not establish genuine device location, independent humans, formal
anonymity or an honest remote operator. Accepted inference limits remain documented.
This batch did not rerun the 15-minute normal-clock public-publisher journey,
100k capacity measurements, maximum-lifetime soak, independent reproducible builds
or exact release vulnerability audit. No new capacity or public-release claim is
made. The local deployment rehearsal uses a fake loopback connector; OVH host,
Cloudflare/TLS and real optional-push/device checks still need the actual deployment.

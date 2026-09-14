# Fast activity and nearby arrival verification — 2026-09-14

Implementation source: **12431f34d503b0c107fcd1e03690e05d4dd37ff0**.
[Decision 0016](../decisions/0016-fast-publication-and-nearby-arrivals.md) records the
user-directed privacy tradeoff. [Sanitized JSON](usability-fast-2026-09-14.json)
contains phase measurements, image identities and the running overnight checkpoint.

## Implemented behavior

Server push is available by default; each participant explicitly opts in. Nearby
arrival accepts the destination cell plus eight adjacent 100-m cells. Public
activity uses fixed ten-second epochs with no extra delayed epoch, a five-second
capture budget and one-minute retention. Cache age/polling follow that cadence;
public geography stays at 1 km, counts remain bucketed and values below 20 hidden.
The form starts at 30 minutes/1 km and explains travel, waiting, tab recovery and
pedestrian meeting space. Gathering/presence deadlines remain bounded and explicit.

Faster updates weaken temporal privacy: colluding inputs and bucket crossings can
reveal changes within seconds. Expanded arrival tolerance also broadens false
presence claims. Neither suppression nor device location is a proof of anonymity
or physical attendance. No exact participant coordinates, tracking or automatic
push permission requests were introduced.

## Verification

- Go race/simulation/vet/type checks, actual store/HTTP/notification/worker race
  integration, 45 Python guard tests and configuration checks passed.
- 42 Chromium cases, one real-clock full journey and 108 device-preset cases passed.
  The full journey first rejects an out-of-range arrival, then accepts a neighbouring
  cell, observes the fast release and admits a new browser to JEMI KËTU.
- Native production build and storeless protocol smoke passed. The fixed smoke
  fixture remains push-disabled; real-store tests exercise the current defaults.
- Private VAPID directory/retained-key and missing-publication monitoring regressions
  passed. A missing publisher now accumulates lag across short intervals.
- Built production-image two-API/worker rehearsal passed in 79.83s, including key
  mounts/egress, network/auth/cache boundaries, store settings and replica failover.

Initial assertions expected old cache/poll timings and assumed a public 20+ arrival
count implied completed presence stability. Corrected them to check the actual
intermediate state; stability remains required for JEMI KËTU. All final runs above
passed. Device presets and synthetic provider/service-worker inputs remain distinct
from physical phones, actual GPS and external push delivery.

## 100,000-signal load

Uniform, seeded, real-clock local run: 100,000 accepted creates in 333.34s. Status
phase served 200,040 requests/60s; public-origin phase 60,000/60s. Mixed status,
public reads and explicit going, plus cancellation, all passed with no rejected
requests, errors or generator drops. p95 upper buckets were 10ms for creation,
public/mixed/going/cancellation and 50ms for the standalone status phase.

Maximum sampled RSS was 573.8 MiB API and 90.9 MiB Valkey; no sampled monitoring
alerts. Two direct public observations during load captured in 435ms and 549ms,
inside the five-second budget. Measurements used the same AMD Ryzen 7 7435HS host
(16 logical CPUs), with a 512-MiB store limit and other test/build jobs running.
They are not a total machine budget or proof of deployed capacity.

This run used the worktree API before the final missing-publication monitor
correction; current capture/release/config behavior was already in place. The JSON
records that binary's hash rather than falsely assigning the later release hash.
Production-image verification includes the final correction. Traffic covers 100k
stored signals, not 100k browsers each polling maps at ten-second intervals; no
cached CDN, real push provider, remote host or maximum-lifetime claim follows.
Raw results are ignored under reports/local/usability-fast-100k.

## Current release and overnight status

Prepared reports/local/usability-fast-release for jamgati.com. Manifest SHA256:
816b1ffe9d25ef7a37cd0c159a83ab6be13065543fdd1f83a2359a3e929d5a41.
Four exact-image audits passed with zero unresolved findings; the two existing
scoped exceptions expire 2026-10-14. Exact-audit freshness/identity checks passed.
Native API/browser hashes match the prepared artifact. Same-code/config local
activation and rollback rehearsal, with a disposable VAPID key, passed in 127.49s.
Rescan within 24 hours of real activation. No new independent two-build review,
real Cloudflare/TLS or remote deployment is claimed.

The new eight-hour run is active at
reports/local/usability-fast-overnight/lifecycle.json, expected to finish
**2026-09-15 00:26:51 Europe/Rome**. It uses this committed API/config and a binary
identical to the prepared release, with no external push subscriptions. It has a
temporary sleep inhibitor and writes its final result automatically; keep the
machine powered on. Running is not passed. The earlier jamgati-alpha-overnight run
retains the old settings and cannot certify these changes.

The local Compose preview was rebuilt and its served configuration verified at
http://127.0.0.1:5173. Nothing was pushed remotely or publicly deployed. Real host,
SMTP delivery and phone/provider checks remain deployment work.

# Local automated hardening evidence — 2026-09-13

Synthetic tests only. Nothing was deployed or pushed remotely. Commands and
remaining gates are in [AUTOMATED_TESTING](../AUTOMATED_TESTING.md); machine,
binary/config hashes, exact phase counts and limitations are in the adjacent
[summary JSON](predeployment-automation-summary.json). Source-dirty flags are
retained rather than describing an in-progress build as a clean release.

Both uniform and single-cell hotspot workloads reached 100,000 accepted signals
with normal production functional configuration, 100-metre private cells and
30-minute availability. On this Ryzen 7 7435HS/16-logical-CPU PC, with API, store,
clients and other isolated tests sharing the host:

| Measured phase | Result in each final run |
| --- | --- |
| Ramp | 100,000 accepted creations at approximately 300/s |
| 100k status polling | 66,680 requests over 20s, all accepted |
| Public origin at 100k | 20,000 requests over 20s, all successful; no CDN |
| 60s write burst | 180,000 attempts; 30,000 accepted and 150,000 explicit 429s |
| Post-burst recovery | 1,000/1,000 successful reads |
| Cancellation sample | 1,000/1,000 successful deletes |
| Client p95 accepted latency | At most 10ms in every measured phase |
| Timeouts / 5xx / dropped scheduled jobs | Zero |
| Operational alerts during final load runs | None observed |

These are sequential traffic phases using synthetic credentials and 512 modeled
network addresses. They do not establish 100k simultaneous connections/humans,
combined status+public load, remote reference-host capacity, cache hit rate or
per-gathering matching/expiry lag. Absence of suppressed operational error buckets
alone does not establish zero failures; exact synthetic client results establish
the response counts above. Raw resource samples can miss instantaneous peaks.

Local offline charts (ignored artifacts, available on this PC):
[uniform](../../reports/local/load-100k-uniform-final/index.html) and
[hotspot](../../reports/local/load-100k-hotspot-full/index.html). They chart CPU/RSS
and tabulate accepted/rejected traffic without third-party scripts or telemetry.

Other passing checks:

- Normal-delay end-to-end journey: 13.8 minutes through willingness, commitment,
  arrival, real delayed publication, map, private preview and a new browser's late
  admission/arrival. No API-response fixtures or shortened production delay.
- Six two-replica recovery checks in 132.52s: paused owner takeover, replay/retraction,
  API crash/restart, interrupted and restarted Valkey, monitoring stopped.
- 901.203s short-TTL storage churn: 695 cycles, 69,500 creations, 23,196 cancellations;
  expiry/cell/push indexes returned to their live-keeper baseline after each cycle.
- Owned 8-MiB no-eviction saturation: writes fail closed with 503; existing credentials
  retain deadlines; creation retries and terminal cancellation work after recovery.
- Restricted-store race tests, seeded 250-step arrival model, configuration rollover,
  100-subscription fake-provider outage with competing workers, bounded fuzzing and
  private-monitor schema/retention/failure checks.
- `make verify-local`: race/simulation/vet, frontend checks/build, configuration,
  Compose validation and smoke checks passed. The existing large MapLibre bundle
  warning is still present; it is not an accessibility or performance pass.

Initial failed attempts remain explicitly distinguished: TCP-only store readiness,
Docker auto-port changes on restart, a relative soak report path, insufficient real
publication wait, overlapping preview ports and two browser-emulation timestamp
quirks were diagnosed and corrected. The first uniform report lacked a separate
accepted-latency histogram during mixed burst traffic, so it did not pass the new
client evidence checker; a complete rerun did. The injected Chromium push test raced worker lifecycle after app closure; explicit CDP worker startup passed five focused repeats and the complete final suite. Final browser results: 33 Chromium tests and 24 each in Firefox/WebKit, including automated accessibility and storage/cancellation resilience. No production push behavior was changed. Two clean exports of commit 4da585e, using independent compilation caches on this PC, produced byte-identical Go and all seven frontend artifacts. The recorded inventory contains Go module metadata and 92 npm entries with license metadata; this is not a complete independent supply-chain/license audit. Hashes are in the summary JSON.

# Automated predeployment checks

All synthetic participation runs below create owned disposable API/Valkey instances.
They do not use the ordinary development app, accept arbitrary remote targets, or
add production test routes. Start with `make doctor`, `make deps`, working Docker
and `make browser-install`. Make selects the pinned project toolchain. Python 3.12+
is needed for the clean-export build rehearsal; other helpers use the standard library.

| Command | Scope and approximate local duration |
| --- | --- |
| `make audit-local` | Native/build/config/smoke checks, Python guard suite and real-store integrations; recommended audit starting point |
| `make audit-journeys` | Sequential Chromium suite, normal-clock full journey and two-API production-image rehearsal; potentially 20+ minutes |
| `make verify-local` | Go race/simulation/vet, frontend type/build, configuration and smoke checks; minutes |
| `make test-store` | Restricted real-store behavior, state model, rollover, push outage/fanout; seconds |
| `make test-planner-replay` | Deterministic 3,000-participant planner snapshots on the full Tirana map and current 100 m configuration; about 35 seconds locally |
| `make test-monitor` | Private socket, suppression, retention, concurrency and collector checks |
| `python3 -m unittest discover -s scripts -p 'test_*.py'` | Release/runtime guards, image audit, collector and synthetic load evidence rejection checks |
| `make test-fuzz` | Five native fuzz targets, including strict request round trips, 30 seconds each by default |
| `make test-browser` | Chromium journeys, optional push fixtures, accessibility and resilience; about one minute |
| `make test-browser-matrix` | Core, accessibility and resilience in Firefox then WebKit; separate stores, a few minutes |
| `make test-devices` | Pixel 7/Chromium, iPhone 13/WebKit and iPad Mini/WebKit presets; core/map/arrival/resilience/accessibility plus touch, rotation, offline cancellation and resume-after-expiry; a few minutes |
| `make test-full-journey` | Actual production matching, arrival and delayed publisher, public map, private preview, late join; up to 20 minutes |
| `make test-recovery` | Two API replicas, paused leader, arrival replay, killed API, disconnected/restarted store, disabled monitoring; about two minutes |
| `make test-pressure` | Owned 8-MiB no-eviction saturation, fail-closed writes and recovery; about one minute |
| `make test-soak SOAK_SECONDS=900 OUTPUT=reports/local/my-soak` | 15-minute real-clock storage churn with deliberately short synthetic TTLs |
| `make test-lifecycle-smoke` | 90-second harness validation using real production lifetimes; checks activation/arrivals/late joining, then destroys the synthetic lab; not expiry/renewal proof |
| `make test-lifecycle-overnight HOURS=8` | Real-clock 30/60/90/120-minute API lifetimes, repeated cohorts, renewal/replay/late joining and deadline/index checks with private memory/health monitoring |
| `make test-ops-alerts` | Synthetic local TLS receiver, failure/retry/recovery, payload allowlist and private webhook configuration checks |
| `make load SIZES=1000,10000,100000 BURST=1 OUTPUT=reports/local/my-uniform` | Real-clock synthetic client ramp and write burst with monitoring; about 9 minutes |
| `make load SIZES=1000,10000,100000 DISTRIBUTION=hotspot BURST=1 OUTPUT=reports/local/my-hotspot` | Same measurements with all willingness in one coarse cell |
| `make check-load REPORT=reports/local/my-uniform/load.json` | Reject incomplete/slow/rejection-masked client evidence |
| `make reproduce-build` | Two clean committed-source exports, separate dependency installs, byte comparisons and inventories; minutes |

Use a fresh output directory per run. Load commands emit an offline `index.html`,
exact **synthetic** client counts/statuses/latency histograms, bounded private
monitoring samples, and machine/source/config/binary metadata. They return failure
when the implemented client gate fails. A report with 100k accepted signals does
not establish 100k simultaneous browser connections, 100k humans or the complete
reference-host capacity gate. The generator uses 512 loopback network addresses,
128 workers and real HTTP/production limits, with 30-minute availability. Status
polling and public-origin traffic are sequential measured phases, not a simultaneous
combined workload. There is no cached-CDN traffic in this local origin test.

The normal-delay journey changes neither functional configuration nor the clock.
A release captures only once per five-minute epoch. A signal submitted after that
capture waits for the next capture; with the additional delayed epoch, button-to-
public-visibility can approach **15 minutes**. This differs from observation-to-
release age. Private matching continues immediately under normal stability rules.
The full journey uses only synthetic coordinates/peers, without API-response mocks.
The faster browser suites deliberately use public-response/provider fixtures where
identified in their source; do not treat them as actual provider delivery.

The soak uses 1.2-second storage fixtures to repeat index/expiry/cancellation and
push-binding cleanup. It retains an expected live keeper so lingering set members
cannot be hidden by whole-key expiry. Set `SOAK_SECONDS=28800` for eight hours, but
even this is **not** a full-API, 120-minute-session, maximum-TTL retention proof.
An extended realistic-lifetime churn scenario and direct
per-gathering matching/expiry-lag measurements remain separate coverage gaps.

## Browser prerequisites and emulation boundaries

On a supported machine with administrator-installed dependencies:

```sh
source scripts/env.sh
cd web
npx playwright install --with-deps firefox webkit
```

An optional Ubuntu 24.04 amd64 fallback avoids administrator access:

```sh
source scripts/env.sh
python3 scripts/browser-local.py
GATI_WEBKIT_EXECUTABLE="$PWD/.runtime/browser-deps/webkit-local" make test-browser-matrix
GATI_WEBKIT_EXECUTABLE="$PWD/.runtime/browser-deps/webkit-local" make test-devices
```

This verifies pinned Debian-package hashes, extracts libraries locally, and creates
a separate launcher for the original pinned WebKit binaries. The launcher excludes
incompatible Snap/editor GTK/GIO paths. Upstream system-dependency validation may
still warn because it checks the system linker cache; the actual browser tests
validate this local launcher. No system libraries or upstream binaries are edited.

`web/tests/fixtures.ts` documents two observed emulation timestamp quirks: Firefox
sets location a day ahead ([upstream source](https://github.com/microsoft/playwright/blob/main/browser_patches/firefox/juggler/content/main.js)); this pinned WebKit
returns its override timestamp in microseconds. The test adapter corrects only
these emulated values. It does not change production location validation. Stale,
future and inaccurate location rejection remain tested separately. These engine
checks are not real Android/iPhone GPS, background behavior or external push tests.

Automated axe checks cover initial/active screens at mobile width, keyboard actions
and enlarged text; they are not a complete accessibility or screen-reader audit.
Synthetic browser failure artifacts must never be enabled for real participants.
Private capabilities stay in disposable test process memory; trace/screenshot
capture is disabled in these suites.

## Monitoring and remaining release gates

See [MONITORING](MONITORING.md) for the owner-only socket, allowlisted completed-window
summaries, bounded collection, alert semantics and failure handling. The load report
charts archived samples; it must not be presented as a live availability dashboard.
Monitoring measures operational health, not individual participation or conversion.

The automation supplements [the broader plan](PREDEPLOYMENT_TEST_PLAN.md). Separate
work remains for overnight realistic lifetimes, gradual production-sized pressure, simultaneous mixed
traffic, mass late admission and synchronized API expiry at scale, 10k cached public
reads/s, per-gathering matching/expiry latency, a separate 4-vCPU/8-GB host with remote
generator, actual edge/network DoS protection, independent human audit, real devices,
volunteer usability, public-map pedestrian suitability, actual-host deployment/rollback and
remote served-artifact verification. Clean builds on this PC demonstrate local byte
reproducibility, not another operator's honesty. No public deployment is performed.

## Device presets and public alpha

The device suite runs browser presets, not Android/iOS virtual machines. Presets
supply viewport, touch, pixel density and user agent. It reuses the full core suites
and their known fixture distinctions; new touch/offline flows use the owned real
API, while suspended-page expiry uses a synthetic session, response and wall-clock
jump. No real participant credentials or location are used. Permission prompts,
OS sleep/process eviction, radio handover, installed-PWA and external push behavior
remain real-device work. Retained diagnostics stay in ignored local reports.

Decision 0015 supersedes earlier volunteer-stage sequencing: the first public use
is an alpha, followed by optional sanitized feedback and synthetic regression tests.
The realistic maximum-lifetime overnight scenario remains outside this batch.

## Real-lifetime overnight run

The subsequent alpha-preparation batch adds scripts/lifecycle_run.py. It owns a
disposable API/Valkey and accepts no target URL; the API is built from the recorded
HEAD export. Commit the driver before starting so its recorded hash/revision are
reviewable. The Make targets set CGO_ENABLED=0 to match the release API build; use
the same environment when invoking the Python driver directly. Compare the lab's
binary_sha256 with the release's api_binary_sha256 before claiming an exact match.
Production clocks/settings stay unchanged, push stays off, and synthetic
capabilities remain in driver memory. The report records only totals and bounded
minute samples; monitoring retains its existing one-hour rolling window.

At eight hours the driver creates three cohorts two hours apart, each with 120-minute
founders and below-threshold 30/60/90/120-minute keepers. The final interval drains
and observes baseline. It checks frozen destinations/deadlines, explicit arrivals,
idempotent retries, repeated renewal, joining after JEMI KËTU, decline/cancellation,
expired HTTP access and known participant keys/index entries. Final expiry/gathering/
push indexes must be empty; conservative RSS drift limits are 256 MiB API/64 MiB
Valkey. This is a small-cohort leak check, not scale proof or complete forensic/store
inventory. No privileged production monitoring endpoint is added.

The computer must stay awake. A wall-clock/suspension discrepancy over 30 seconds
fails the run. The private output's lifecycle.json says running, passed, failed or
interrupted and includes its expected completion time; never count running as passed.
SIGTERM to the recorded driver PID exits the owned lab and removes its test store.
An unclean host crash can leave an orphan lab; inspect recorded ownership before
removing it. Do not stop the regular development stack to clean a synthetic run.
The 90-second smoke has a distinct smoke_passed result and does not shorten TTLs.

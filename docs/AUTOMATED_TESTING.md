# Automated predeployment checks

All synthetic participation runs below create owned disposable API/Valkey instances.
They do not use the ordinary development app, accept arbitrary remote targets, or
add production test routes. Start with `make doctor`, `make deps`, working Docker
and `make browser-install`. Make selects the pinned project toolchain. Python 3.12+
is needed for the clean-export build rehearsal; other helpers use the standard library.

| Command | Scope and approximate local duration |
| --- | --- |
| `make verify-local` | Go race/simulation/vet, frontend type/build, configuration and smoke checks; minutes |
| `make test-store` | Restricted real-store behavior, state model, rollover, push outage/fanout; seconds |
| `make test-monitor` | Private socket, suppression, retention, concurrency and collector checks |
| `python3 -m unittest discover -s scripts -p 'test_*.py'` | Collector and synthetic load evidence rejection checks |
| `make test-fuzz` | Four native fuzz targets, 30 seconds each by default |
| `make test-browser` | Chromium journeys, optional push fixtures, accessibility and resilience; about one minute |
| `make test-browser-matrix` | Core, accessibility and resilience in Firefox then WebKit; separate stores, a few minutes |
| `make test-full-journey` | Actual production matching, arrival and delayed publisher, public map, private preview, late join; up to 20 minutes |
| `make test-recovery` | Two API replicas, paused leader, arrival replay, killed API, disconnected/restarted store, disabled monitoring; about two minutes |
| `make test-pressure` | Owned 8-MiB no-eviction saturation, fail-closed writes and recovery; about one minute |
| `make test-soak SOAK_SECONDS=900 OUTPUT=reports/local/my-soak` | 15-minute real-clock storage churn with deliberately short synthetic TTLs |
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
volunteer usability, public-map pedestrian suitability, deployment/rollback and
remote served-artifact verification. Clean builds on this PC demonstrate local byte
reproducibility, not another operator's honesty. No public deployment is performed.

# Device and release refresh — 2026-09-14

Source: **`e5a374dd4972370c839b3ce43b588e9d93f314d2`**. The user selected a direct public alpha, without a
separate volunteer stage; see decision 0015. This batch changes tests, CI and
planning documentation only. Manual Albanian HTML copy remains unchanged.

## Browser results

| Target | Result |
| --- | --- |
| Pixel 7 preset / Chromium | 35 passed |
| iPhone 13 preset / WebKit | 35 passed |
| iPad Mini preset / WebKit | 35 passed |
| Desktop Firefox | 32 passed |
| Desktop WebKit | 32 passed |

The final mobile suite took 2.6 minutes; desktop suites took 1.2 and 1.3 minutes.
These durations are verification runtimes, not capacity measurements. Each suite
uses an owned disposable API/Valkey lab with production functional settings.
The three new cases use injected touch, viewport rotation, offline recovery and
resume-after-expiry. The existing 32 cases cover willingness, map/private preview,
late admission, arrival, renewal, cancellation, errors and accessibility. GPS and
selected public/private responses remain fixtures as documented in test sources.
No new production routes, dependencies, tracking or functional settings were added.

The first mobile run passed 102/105: WebKit exposed zero maxTouchPoints despite
supporting injected touch, and a Chromium protocol event delivered a valid startup
request after the runner began counting expiry requests. Assertions now observe
actual touchstart events and count authorization at fetch creation against the
expiry deadline. The subsequent complete 105-case run passed without retries or
application changes. These are test-harness corrections, not device bugs fixed.

Emulation does not establish real GPS/permission prompts, iOS/Android OS behavior,
battery management, cellular handover, installed-PWA/background behavior or external
push delivery. It does not establish suitable pedestrian space at a crossroad.
Push remains disabled. The maximum-lifetime overnight scenario was not run.

## Exact release and verification

Local release: `reports/local/alpha-release-refresh`, hostname **localhost**.
Manifest SHA256: `2f4cc2f99ece22c281c0c0211774591a4bf404f9f135cc624773410234adca8a`.
Four image audits passed with zero unresolved findings; API, Caddy and tunnel binary
symbol scans passed. The inventory retains four accepted package matches across
three images for two scoped advisories (OpenPGP and CEL), expiring **2026-10-14**.
The exact-release manifest/image/exception/freshness gate passed. Scanner and
vulnerability-database dates, image identities and exceptions are in the
[sanitized JSON evidence](alpha-refresh-2026-09-14.json).

Two clean exports with separate compilation caches produced identical API, Caddy,
cloudflared and browser assets. All three runtime binary hashes and all seven
browser assets match the exact audited release. This is same-host reproducibility,
not another reviewer's independent build or proof of remote behavior.

The release rehearsal passed in 136.06 seconds: actual combined-role container
activation; rejection of an extra network attachment; served assets/effective config
comparison; two compatible directory switches preserving store identity/start time,
capability and original expiry; final cancellation. It uses identical code/config
in both directories, not arbitrary schema migration. Its connector is local and
synthetic; no real Cloudflare tunnel, DNS or host settings were changed.

Additional validation: make test (race/simulation/vet/types), all 32 Python guard
tests, production builds, source security checks and staged publication checks
passed. Source scanning found no called Go vulnerabilities and zero npm findings;
the unused Go module advisory remains in the exact image inventory. Environment-
gated store tests skipped by make test are not separately credited as executed.
The real-store browser journeys and release rehearsal were actually executed.

Raw evidence remains ignored under reports/local/alpha-browser-matrix,
alpha-devices-final, alpha-release-image-audit, alpha-release-reproduction and
alpha-release-rollback. Retain it with this release. Local tools were already
installed; the existing WebKit library launcher was reused. Remote CI was updated
but was not run or published during this batch.

## Next deployment work

Complete OVH access and host/Cloudflare/TLS/cache verification, choose a private
operator alert destination and add the alpha label/optional feedback instructions.
Prepare the final jamgati.com artifact and rescan within 24 hours of activation;
this localhost artifact is rehearsal evidence, not a public release. Review host
capacity before selecting an operational cap. No new 100k/1M capacity claim is made.
Independent review and real-device gaps remain disclosed. There is no separate
volunteer-stage requirement. Feed sanitized alpha reports into synthetic regression
tests and small audited updates; no individual usage tracking. Nothing was pushed
or publicly deployed.

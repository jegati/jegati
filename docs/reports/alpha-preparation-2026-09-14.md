# Four alpha preparations — 2026-09-14

Release source: **6efad344de81fdd474aeb1185e4242ffc7094019**, configured for
**jamgati.com**. Manifest SHA256:
9965dc5223b897faf2bcfe4498e976d76c46596d247a2f9b5c2c740e00a5e52f.
[Sanitized machine-readable evidence](alpha-preparation-2026-09-14.json) records
exact audit identities, exceptions and the overnight status at this checkpoint.

## Alpha notice and feedback

Added Albanian alpha notice, optional static feedback guidance and an Albanian
GitHub issue form. Fixed external links transmit no application state. Email and
public GitHub identity/retention boundaries are explicit; reporting is optional and
creates no in-app record. Existing manual About copy is preserved. No analytics,
new runtime dependency or behavioral threshold change.

Chromium passed 42/42 checks; Pixel/iPhone/iPad presets passed 108/108. Initial
200% heading overflow was corrected with scoped CSS wrapping. The new feedback
regression verifies no automatic external request, state leakage or enrollment.
Device emulation remains distinct from actual OS/GPS/background behavior.

## Private operator alerts

Optional generic HTTPS JSON sender and systemd collector/sender templates are
prepared. Only fixed health labels leave the host; private endpoint/auth files,
TLS verification, no redirects, debounce, bounded retries/reminders and recovery
are implemented. All 40 Python guard tests passed, including four alert tests and
four lifecycle tests. A local TLS receiver exercised failed delivery, retry,
recovery and redirect rejection. Both service templates passed systemd verification.

No external destination or credentials were configured, no external notification
was sent, and no systemd service was installed. This is a generic webhook contract;
a provider-specific or SMTP adapter depends on the operator's choice. Same-host
alerts cannot report a dead VPS/network; independent availability remains host work.

## Domain package and reproducibility

Prepared reports/local/jamgati-alpha with hostname jamgati.com. Four exact images
passed the audit with zero unresolved findings and three runtime binary scans
passed. Two existing scoped advisory exceptions remain, expiring 2026-10-14.
The exact manifest/image/freshness/exception gate passed; rescan within 24 hours
of actual activation. This does not mean there are no possible vulnerabilities.

Two clean source exports with separate compilation caches produced identical
artifacts. Three runtime binaries and all seven browser assets match the audited
release. This is same-machine/toolchain reproducibility, not independent review.

Hostname-aware local rehearsal passed in 145.38 seconds: container activation,
unexpected network rejection, served assets/effective settings verification and
two compatible release-directory switches preserving store/session/original expiry.
The connector forwards the verified configured hostname only to its local Caddy;
it does not use public DNS or a real Cloudflare tunnel. Rollback uses identical
code/config, not a schema migration. Actual OVH/TLS/edge behavior remains untested.

Raw evidence: reports/local/jamgati-alpha-audit, jamgati-alpha-reproduction and
jamgati-alpha-rehearsal. Host/edge steps are in ALPHA_DEPLOYMENT.md and DEPLOYMENT.md.
The later Make/documentation-only commit does not change these recorded artifacts.

## Eight-hour lifecycle run — still running at this checkpoint

The active run is reports/local/jamgati-alpha-overnight/lifecycle.json. It started
with unchanged production settings and CGO_ENABLED=0; its API binary is byte-identical
to this release. The driver/API source is pinned to the recorded committed export.
The lab's dirty-worktree flag reflects subsequent Make/documentation work, not an
uncommitted API build. The source/API configuration contains no simulation timing
changes. Synthetic cohorts exercise normal 30/60/90/120-minute lifetimes and monitor
memory/health locally. Expected completion: **2026-09-14 23:29:05 Europe/Rome**.

The earlier workstation-default-build run was stopped gracefully and is recorded
as interrupted in reports/local/alpha-lifecycle-overnight; it is not passed evidence.
The Make targets now explicitly use the release CGO setting. Final 90-second smoke
passed activation, arrival/retry/retraction, cancellation, decline and late joining;
it did not prove renewal, expiry or overnight stability.

The active driver writes its final result and cleans up its owned API/store
without a new chat turn. Keep the workstation powered on; its temporary sleep
inhibitor ends with the driver. A crash/reboot or clock discontinuity invalidates
the run. Check status=passed and the terminal report before crediting real-lifetime
renewal/expiry/baseline gates. A running report is not successful overnight evidence.
No capacity or complete forensic-erasure claim follows from this small-cohort test.

Next: review that completed report, choose a private alert channel, obtain VPS/SSH
access and perform actual host/edge checks. Nothing was pushed or publicly deployed.

# Additional testing before deployment

2026-09-13 review. The table below is the original scope. See [automated testing](AUTOMATED_TESTING.md) and PROGRESS.md for implemented checks, passing evidence and remaining gaps. The current
baseline is [the selected usability report](reports/usability-final.md): 28 browser
checks, real-store/race tests and a checked 3,000-person accelerated simulation.
No tests were rerun for this review. No deployment or product changes are authorized
by this document. Known inference/spoofing limits remain documented and accepted
for implementation; these tests must not silently claim to solve them.

## Priority order

| Work | Concrete checks | Acceptance evidence |
| --- | --- | --- |
| 1. Isolated complete journeys | Give browser suites their own API/store and fresh state. Drive actual willingness → activation → going → arrival → delayed real publisher → map → private preview → late join. Exercise the normal publication delay at least once; keep fast isolated-clock cases for routine regression. Repeat the previously timed-out collective test with bounded stage diagnostics. | No public-response mocks in this dedicated journey; exact source/config/seed recorded; cleanup verified even on failure. Diagnose repeatable timeouts rather than hide them with retries. |
| 2. Failure and concurrency recovery | Kill/restart API/worker during creation, activation, arrival and push claims; run competing workers/API replicas; pause the lease owner; disconnect/restart Valkey; drop successful replies; change config/map while gatherings are live. | No duplicate counting, stale-worker writes, moved destinations or resurrected cancelled/expired state. Explicit temporary-state loss after a nonpersistent-store restart; recovery never silently enrolls again. |
| 3. Long-running expiry and churn | Run overnight with repeated create/decline/join/arrival/opt-in/opt-out/cancel cycles, offline tabs and expired credentials. Inspect bounded key/index/queue populations through multiple 120-minute session lifetimes. Inject storage pressure. | Participant state and indexes return to expected baselines after deadlines/cleanup; memory usage stabilizes rather than growing with total historical users. Browser expired bytes are distinguished from server-valid credentials. |
| 4. Real-time load and abuse | Ramp 1k → 10k → 100k live signals; uniform geography and dense hotspots, mass joins/cancels, synchronized expiry, polling bursts and fake-provider outages/fanout. Measure shared-network rejection and hostile traffic effects on ordinary clients. | Report attempted/accepted/rejected traffic, latency, worker delay, expiry lag, CPU/memory and recovery after bursts. Use DEVELOPMENT_PLAN targets; do not count rejection throughput as accepted-user capacity. Local PC sizing precedes a separate reference-host result. |
| 5. Adversarial input and privacy regression | Add bounded native Go fuzzing for untrusted parsers and geographic edge cases; randomized transition sequences with invariant checks. Probe wrong capabilities, replay, cross-origin requests, script injection and push URL/DNS boundaries. Trace synthetic canaries across requests, caches, storage and allowed diagnostics, including failures. | No crashes, hangs, unauthorized transitions, exact-coordinate transmission, secret exposure or private shared caching. Preserve failing fuzz cases. Known nearest-crossroad/differencing inference remains reported separately from direct-data disclosure. |
| 6. Browsers, devices and human usability | Add Firefox/WebKit coverage, slow/flaky connections, storage denial/eviction, multi-tab races, background/resume, zoom, keyboard, screen reader and automated accessibility checks. Test real Android/iPhone location and push in an isolated trusted-HTTPS test environment. Observe a small first round of Albanian-speaking volunteers using synthetic gatherings. | Record first-action success/time, recovery/cancellation success and whether users understand willingness, commitment, delayed buckets and arrival expiry. Inspect sampled mapped crossroads for practical pedestrian meeting space; automated selection does not establish real-world suitability. No real-location/identity/session recordings. |
| 7. Clean build and audit rehearsal | Rebuild from a clean checkout with pinned dependencies on a separate environment. Compare unsigned frontend/backend artifacts, inventory dependencies/licenses, inspect production artifacts for simulation controls, review data flows and deletion contracts. | Reproducible results or explained differences and a scoped review report. Independent human review remains distinct from agent checks; local artifacts do not attest a remote operator. |

Native Go fuzzing is documented by [Go](https://go.dev/doc/security/fuzz/).
Automated accessibility checks supplement manual review; they cannot establish
complete accessibility, as [Playwright's guidance](https://playwright.dev/docs/accessibility-testing)
explains. Adapt security checks to GATI's capability model rather than adding accounts.

The high-value starting batch is 1–3 plus input-boundary fuzzing. It gives the load
and browser/device work a repeatable environment. Remaining features such as area
follows and daily summaries are implementation gaps, not gaps that testing can fix.

## Prepare privacy-preserving pilot monitoring locally

- Collect allowlisted service-health measurements: service-wide latency/error
  buckets, worker/expiry lag, resource use and bounded queue-age ranges. Do not emit
  capabilities, endpoint URLs, IPs, device IDs, cell/gathering labels, request bodies
  or per-participant traces. Set explicit aggregation and retention bounds; suppress
  sparse activity-derived metrics and review combinations for inference. Keep raw
  operational metrics out of the public participant/map interface.
- Use the existing delayed/suppressed public releases for collective activity
  trends. Do not reconstruct individual willingness → going → arrival funnels or
  interpret a decision not to attend as a usability failure.
- Combine synthetic read-only health checks with periodic, voluntary usability
  sessions and sanitized issue reports. Full synthetic participation journeys belong
  in a separate test environment, never in real public gathering counts. Production
  artifacts must remain free of test clocks and control routes.
- Predefine alerts and a response guide for unavailable API/store, stale public
  releases, stuck worker/cleanup and push failure. Exercise alerts locally with
  injected faults. The private monitoring foundation is implemented; see MONITORING.md for its narrower measured scope and alert limitations.
- Test that monitoring failure cannot block participation or leak private data.
  Review sensitive-data exclusions alongside [OWASP's logging guidance](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html);
  GATI deliberately omits individual access/event logs.

Live provider delivery needs a real operator contact and a suitable test device;
a trusted local test environment does not require public launch. Real network/edge
DoS protection, actual host controls, rollout/rollback and served-artifact verification
remain deployment-stage checks. These local tests improve evidence without claiming
complete location authenticity, unique humans or absolute anonymity.

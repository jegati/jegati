# Selected usability slice: final local evidence

2026-09-13. Implements suggestions **1, 2, 3 and 5** from
[the selected plan](../USABILITY_IMPLEMENTATION_PLAN.md): optional temporary push,
one-action device-based JAM GATI, eligible private crossroad preview before
commitment, and clearer waiting/recovery/arrival-expiry guidance. Sharing links/QR
and new purpose copy are excluded. Local commits contain code, tests and docs.

## Validation

- `make verify-local`: Go race/simulation-tag tests, vet, TypeScript, builds,
  configuration validation, Compose parsing and native HTTP smoke passed.
- `make test-store`: actual restricted ephemeral Valkey and race tests passed,
  including strict push API, encryption, opt-out revision/stale-write rejection,
  bounded queues/retries, claim fencing, rate-gap behavior and expiry-index cleanup.
- `make simulate`: isolated real-store fixed-clock activation/admission/arrival and
  private-preview/background-offer checks passed.
- `make test-browser`: **28 passed** against current built client and rebuilt API.
  Nine push tests cover explicit/denied/unsupported opt-in, expiry, lost replies,
  cancellation races, permission revocation, recovery and closed-page display.
- Refreshed Compose API/web, `make check-containers` and optional overlay parsing
  passed. The configured 100 m private / 1 km public grid remains unchanged.
- Pinned `make security-check`: zero reachable or imported-package vulnerabilities;
  one advisory in an unused OpenPGP package of a required module remains documented
  in THIRD_PARTY. This scan is not an independent audit.

The worker test uses full Chromium's native notification API with the app closed;
CDP replaces provider delivery and fixtures replace subscription registration/private
state. Separate Go tests exercise actual store plus fake provider transport. No
external Google/Apple/Mozilla delivery, OS click or human usability study is claimed.
A preliminary collective-flow browser timeout remains unexplained; its focused
rerun and final full suite passed. PROGRESS records the failed runs and corrections.

## Final Tirana regression

```sh
make simulate-population SCENARIO=tirana-population SEED=42 OUTPUT=reports/local/tirana-usability-final-3000-42
make check-population REPORT=reports/local/tirana-usability-final-3000-42/report.json
```

The checker passed all configuration, funnel and expiry invariants. Source ead158a;
full revision, reported dirty flag, scenario, hardware metadata and artifact hashes
are preserved in [the synthetic summary](usability-final-summary.json). This is
recorded local-build evidence, not an independently reproduced clean release.

| Metric | Result |
| --- | ---: |
| Accepted credentials | 3,000 |
| Distinct invited / going | 1,920 / 1,715 |
| Distinct arrivals / accepted confirmations including renewals | 1,107 / 1,554 |
| Gatherings observed / ever confirmed | 46 / 30 |
| Joins after JEMI KËTU | 24 |
| Cancelled / expiry verified | 456 / 2,544 |
| HTTP requests | 262,740 |
| Wall time | 239.73 seconds |

The runner retained current functional settings and changed only the profile for
isolation. Push was disabled. Discovery is synthetic; this population run does not
exercise browser previews or push delivery. Accelerated time and concurrent local
work make this a functional regression, not a throughput comparison or 100k test.
[Local replay](../../reports/local/tirana-usability-final-3000-42/index.html).

## Remaining input and limits

[NOTIFICATIONS.md](../NOTIFICATIONS.md) gives the concrete setup/manual checks.
A real public operator contact URL is needed before enabling live push; service
keys already exist in ignored local storage. Device/provider interoperability is
unverified. The app works without push. Known location spoofing, Sybil and inference
limits remain; no public deployment, production readiness or remote Git push is
claimed. Follows, daily summaries and release/audit gates remain separate work.

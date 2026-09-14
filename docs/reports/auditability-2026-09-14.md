# Auditability cleanup — 2026-09-14

This first cleanup batch changes code organization and reviewability, preserving
product behavior and privacy boundaries. No remote push, public deployment or
new feature. Source baseline: `2335b81`; implementation through `3d3015c`.

| Commit | Result |
| --- | --- |
| `4560a64` | Readable Python release/verification/audit/rehearsal code; Ruff version and formatting scope pinned. |
| `b0a1dd8` | Explicit own-session and private-invitation JSON allowlists, including nested map landmarks; storage no longer defines these response types. |
| `26242b1` | One strict JSON object parser used by willingness, join/preview, intent, arrival and push; original endpoint checks retained. |
| `3d3015c` | Removed unused matching field/wrapper; labeled reserved settings; added audit index/commands and reconciled stale current documentation. |

## Regression evidence

- Python syntax trees before/after formatting match after normalizing only docstring
  indentation. All 25 Python guard tests pass. The formatting check passes.
- New response-field assertions passed against the original implementation, then
  passed after refactoring, including the live fixed-clock gathering lifecycle.
  Tests check private sentinel bookkeeping, nested fields and optional geometry.
- Strict-body cases passed against the original parsers. Old/new differential fuzz:
  798,358 executions in 31 seconds, identical acceptance and accepted values. The
  temporary old-parser copies were removed afterward. Permanent matrix and round-trip
  fuzz cover all five body schemas; the new fuzz target passed 511,121 executions.
- `make audit-local` passed: Go race/simulation-build/vet, TypeScript, builds,
  production/simulation config, Compose parsing, HTTP smoke, Python guards and
  actual disposable Valkey integrations. Environment-skipped tests in plain
  `make test` are not counted as store integration evidence.
- All 33 Chromium checks passed after the response-type change and again against
  the final code (57.1 seconds), with device/provider fixtures where identified
  in the test sources. The production-image browser and
  two-API/worker rehearsal passed after the parser/dead-code cleanup in 72.06 seconds,
  with zero observed proxy 5xx during deliberate replica loss.
- The full browser→real matching→arrival→normal delayed publication→private
  preview→late admission journey passed in 15.4 minutes, with no API-response
  fixtures or altered functional clocks/settings.
- Formatted release helpers passed a 127.36-second compatible rollback rehearsal
  against the prior `06fa798` application images. This was not a new release audit
  or a cross-schema migration test.
- Two clean exports of `3d3015c`, with separate compilation caches, produced identical
  API, browser, Caddy and cloudflared bytes on the same host/toolchain. This is local
  reproduction, not independent attestation or remote backend honesty.

Detailed synthetic logs/manifests remain under `reports/local/audit-cleanup`,
`audit-format-release`, `audit-deployment`, `audit-reproduction`, `audit-dto-journey`,
`audit-parser-journey`, `audit-full-journey`, `audit-final-browser` and
`browser-20260914T080916`. Live production-style labs record
source-dirty flags and binary/config hashes rather than claiming a clean release
for an intermediate working-tree build.

## Seeded Tirana comparison

Both runs used seed 42, 3,000 synthetic people, the same scenario/input/map hashes,
and identical effective production settings (100 m private cells). Both pass the
existing report checks for expiry, admission, retries, configuration and funnel
invariants. The smaller fixed-clock willingness report is identical across baseline,
response-type and parser revisions.

| Measure | Baseline | After API cleanup |
| --- | ---: | ---: |
| Accepted credentials | 3,000 | 3,000 |
| Invited credentials | 1,897 | 1,898 |
| Distinct credentials with an accepted arrival | 1,067 | 1,063 |
| Accepted arrival operations | 1,486 | 1,501 |
| Cancelled / verified expired | 456 / 2,544 | 456 / 2,544 |
| Direct joins accepted / rejected | 219 / 67 | 219 / 67 |
| Gatherings observed | 43 | 45 |
| Wall seconds | 218.5 | 216.2 |

The larger runs are **not exact replay equivalence**. Actor behavior is seeded,
but the driver generates cryptographically random credentials and the worker
random gathering IDs; hashes/IDs participate in matching tie-breaks. These runs
validate scenario invariants, not unchanged outcomes for every actor. Their paced
wall times are not speedup measurements. Detailed reports:
`reports/local/audit-baseline-population-3000` and `audit-after-population-3000`.

## Limits and next scope

Functional configuration is byte-identical when rendered canonically. YAML comments
change the file's raw hash: immutable release identity and the existing strict
rollback compatibility gate still apply. No actual threshold or lifetime changed.

Deeper browser-state/controller and Lua layout/catalog refactors remain separate
follow-ups. This batch deliberately preserves those state machines and scripts.
No 100k/1M load test, external push/GPS test, independent audit or fresh exact-release
image vulnerability audit was performed here. Before deployment, prepare a new
jamgati.com release and apply the existing host/edge/runtime audit gates.

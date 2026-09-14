# Runtime remediation and release evidence — 2026-09-14

The runtime blockers from the [final review](security-2026-09-14.md) are addressed
for source **06fa798c409c491f078d36dec155694a67c52fd8**. All four exact release images pass the
new audit with zero unresolved findings and zero high/critical package findings.
Two explicit unused-code exception rules remain; this is not an exception-free
inventory or a promise that the software has no vulnerabilities.

The release is ready for the next controlled staging steps. Public launch still
requires independent review, a selected/prepared VPS/domain, actual Cloudflare and
real-device checks, private reporting/alert channels and deployment authorization.
No public deployment, account creation or remote Git push was performed.

## Changes and residual findings

| Component | Before | New release |
| --- | --- | --- |
| API | Stripped Go binary; one unused-module finding | Symbol-retaining Go 1.27.1 binary; same unused OpenPGP finding, no affected symbols |
| Web/Caddy | 117 package/advisory pairs, including 40 high; 25 binary-symbol advisories | Locked minimal Caddy source build; 2 package findings, no affected symbols |
| Tunnel | 28 package/advisory pairs, including 3 high; upstream Sentry initialization | Patched source build in scratch; 1 package finding, no affected symbols; Sentry initialization disabled |
| Valkey | Zero scan findings | Same pinned image; zero scan findings |

Caddy and cloudflared use patched x/crypto, x/net, x/text, gRPC and compression
versions recorded in their Go locks. Scratch final images omit the unrelated OS
libraries and tools that produced many earlier findings. Dependency licenses/notices
and module inventories are retained in the runtime images. All three Go binaries
retain symbols, enabling precise binary vulnerability checks. Stripped binaries
make govulncheck 1.8.0 fall back to module-level precision; apparent affected symbols
from that fallback are not proof of actual linkage.

The two exception rules in `deploy/security-exceptions.json` expire **2026-10-14**:

- GO-2026-5932, x/crypto v0.57.0, API/web/tunnel: deprecated OpenPGP is not imported.
  [Go advisory](https://pkg.go.dev/vuln/GO-2026-5932).
- GHSA-gcjh-h69q-9w9g, cel-go v0.28.1, web only: Caddy does not call the affected
  NativeTypes / ParseStructTag functions; exact binary scanning agrees. CEL 0.29/0.30
  breaks the pinned Caddy interpreter API, so this needs a compatible upstream
  update, not an unexplained blanket exception.
  [Go advisory](https://pkg.go.dev/vuln/GO-2026-6094).

Both rules name the service, exact package/version, owner, evidence and review
period. New/mismatched/expired findings fail the image audit. A failed binary scan
is never overridden by a package exception. The audit contains four accepted
package/advisory matches across three images, representing these two advisory IDs.

Cloudflared source review found four explicit Sentry initializers in the tunnel and
Access commands. Discarding Docker logs did not disable that outbound reporting
path. The checked GATI patch replaces those initializers with no-ops and rejects
unexpected source locations/counts. Tests exercise the patched code with a DSN set
and verify that no reporting client or event exists. The connector reports version
`2026.9.1-gati.1`; it is a modified source build, not an unmodified upstream binary.
Cloudflare's own visibility into traffic remains a separate documented boundary.

## Exact artifact and evidence

Local artifact: `reports/local/release-runtime-06fa798`.
Manifest SHA256: `ef9b61e9db5e4dfa0597e9f5d2ec2b85fb7a38d67fbdc749c21fefed3bc3ae87`.
The image archive is approximately 84 MiB and contains no participant state or
operational secrets. The hostname is `localhost`, deliberately for local rehearsal;
prepare a new manifest for the actual hostname before deployment.

| Executable | SHA256 from production image and both clean builds |
| --- | --- |
| API | `3d719390c3781eddb9b7b0beab426c19cc5d68380a9e7aa91ec642b1803bf167` |
| Caddy | `d9e84314255cd1b9b56149f2c4f14152699d4fb09ec3828df10d73cc0fccfd28` |
| cloudflared | `e363c0a5795373726f98315070f4d30b0839b44af0ab992c1312361a321dee2d` |

Two clean source exports with independent compilation caches matched each other,
and every executable/browser hash matched the actual release image. This is a
same-host Go 1.27.1 / Node 24.21.0 comparison, not an independent person's review,
bit-identical OCI metadata, full upstream Valkey reproduction or remote attestation.
The evidence-only documentation commit following 06fa798 changes no executable or
configuration behavior and is intentionally outside this source artifact.

The [sanitized audit and test summary](runtime-remediation-2026-09-14.json) records
identities, database dates, exceptions, binary checks and rehearsal scopes.
Raw image scans, per-image CycloneDX SBOMs and logs are under
`reports/local/runtime-remediation/release-audit/`. Clean build evidence is under
`reports/local/runtime-remediation/reproduction/`. Preserve these alongside the
release when preparing public audit material.

## Validation and corrected failures

- `make test`: application race/simulation tests, vet and TypeScript pass.
- `make security-check`: application Go/npm audits pass.
- `make test-store`: fresh live Valkey race integrations pass.
- Python tests: **25 pass**, including expired/mismatched image exceptions, failed,
  stale, future-dated and wrong-release audits, and runtime configuration drift.
- `make test-runtimes`: both patched Sentry tests, Caddy health checks and runtime
  source vulnerability scans pass. Source scans retain the documented unused-code
  warnings; they do not report called vulnerabilities.
- `make audit-images`: every image scanned by pinned Trivy 0.74.0; all exact Go
  binaries pass govulncheck. Passing audit evidence is bound to the manifest and
  binary/image identities; public activation requires evidence at most 24 hours old.
- Production proxy/browser/two-API rehearsal: **71.88s**, covering
  location-derived willingness, real activation/arrival/replay, cache bypass,
  forged headers, network budgets, private monitoring and deliberate API loss.
  **2 transient 5xx responses** occurred
  during the deliberate failure; bounded retries recovered and cancellation passed.
- Actual release/rollback rehearsal: **134.26s**, including
  extra-network rejection, served assets/config matching and two compatible release
  switches preserving Valkey identity, original session expiry and cancellation.
- All owned rehearsal containers/networks were removed afterward.

The work exposed and corrected three build/test assumptions: incompatible CEL APIs,
read-only directories copied from the Go source cache, and a failover harness that
assumed one successful request meant DNS had fully recovered. The harness now keeps
plain proxy failures visible and tests bounded retry. No production retry behavior
was weakened to make a test pass. Earlier failed outputs remain in local reports.

No new 100k/1M capacity claim is made. These changes concern runtime dependencies,
build verifiability and release gates; existing [scale measurements](scaling-2026-09-14.md)
remain scoped to their tested workload and host.

## Next operator steps

1. Provide the chosen domain and VPS/hosting budget. If a VPS already exists, arrange
   SSH keys privately and provide a local SSH alias; never paste credentials.
2. Establish private security-reporting and operational alert destinations. Arrange
   independent review of source, artifacts and privacy claims.
3. Prepare the dedicated host and run preflight. This development PC still has swap
   enabled and fails the host privacy gate; its swap settings were not changed.
4. Prepare a release for the actual hostname and run a new image audit. Follow
   [runtime auditing](../RUNTIME_SECURITY.md) and the [deployment runbook](../DEPLOYMENT.md).
5. Privately provision Cloudflare credentials, validate the actual named tunnel,
   origin isolation, cache/expiry/authentication bypass and edge rate limits. Real
   tunnel connectivity has not been tested without an account/token.
6. After explicit authorization, use a capped pilot to validate the full Albanian
   journey on actual devices, recovery and operator alerts. Keep push off unless
   separately enabled and tested. Grow only from measurements on the target host.

Location spoofing, Sybil manipulation, accepted inference and privileged provider
visibility remain architectural limitations. Today's audit is evidence for this
artifact and date, not a permanent security or anonymity certificate.

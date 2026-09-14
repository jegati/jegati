# Runtime builds and image auditing

Production API, proxy and tunnel images now contain symbol-retaining Go executables
in scratch images. Valkey remains the pinned upstream image. See
[decision 0013](decisions/0013-auditable-runtime-builds.md) for the reason and limits.
Caddy uses `deploy/caddy/go.mod` / `go.sum`; cloudflared uses `deploy/tunnel/source.json`
and its Go locks. `scripts/runtimeprep` disables the reviewed Sentry initialization
sites. The tunnel is a GATI source build, not the unmodified upstream image.

## Local checks and release audit

```sh
source scripts/env.sh
make bootstrap
make test-runtimes OUTPUT=reports/local/runtime-check
make production-build
make test-deployment OUTPUT=reports/local/runtime-deployment
# Commit the intended code/config before exporting a release.
python3 scripts/release.py prepare --release reports/local/RELEASE --public-host localhost
make audit-images RELEASE=reports/local/RELEASE OUTPUT=reports/local/RELEASE-audit
python3 scripts/rehearse-release.py --release reports/local/RELEASE --output reports/local/RELEASE-rehearsal
make reproduce-build OUTPUT=reports/local/RELEASE-reproduction
```

Use new output directories for each run. `bootstrap` supports Linux x86_64 and
installs checksum-pinned Trivy 0.74.0 alongside the existing toolchain. Image audits
use local Docker images, not a hosted scan service. Only public vulnerability
metadata/dependencies are downloaded. Every image gets raw JSON and a CycloneDX
SBOM; `audit.json` summarizes exact image/binary hashes, accepted and unresolved
findings, database/version metadata and symbol checks. These reports contain no
participant state. Retain the exact audit with the release evidence.

`deploy/security-exceptions.json` lists precisely scoped, expiring exceptions,
with owner, rationale and evidence. Current exceptions expire 2026-10-14:

- OpenPGP GO-2026-5932 in x/crypto v0.57.0 for API/web/tunnel: package not imported.
- CEL GHSA-gcjh-h69q-9w9g in v0.28.1 for web only: affected NativeTypes and
  ParseStructTag functions are not called or linked by this build. Upgrading CEL
  directly breaks the pinned Caddy interpreter API; revisit with a compatible
  upstream update. [Go advisory](https://pkg.go.dev/vuln/GO-2026-6094).

The full scanner inventory remains visible. A failed binary scan is never waived
by these package exceptions. Retained symbols matter: stripped Go binaries cause
this scanner to fall back to conservative module-level results. No exception means
an unresolved finding, even for UNKNOWN severity. Exception expiry requires review
and a fresh scan; it is not permission to mechanically extend a date.

## Public activation gate

On the separately authorized, prepared deployment host, after choosing a real
hostname and completing the host/edge checks:

```sh
python3 scripts/release.py up --release /opt/gati/releases/RELEASE --edge --audit /opt/gati/audits/RELEASE/audit.json
```

The audit must pass, identify that exact manifest/images/binaries, be at most 24
hours old and contain no expired exception. This gate does not configure Cloudflare,
prove host privacy or authorize publication. Files can be forged by a malicious
operator; independent artifact/host review remains necessary. Rollback with --edge
also requires a fresh audit for the rollback target. Private local activation
continues to work without an image audit, allowing remediation/testing.

## Reproduction and maintenance

`make reproduce-build` compiles API, Caddy and cloudflared and builds browser assets
from two clean exports with separate compilation caches. Compare their hashes with
`runtime_binary_sha256` and `browser_assets` in release.json. OCI metadata may differ;
this verifies executable/browser bytes. It does not independently reproduce every
third-party upstream Valkey binary or prove remote operator honesty.

Final proxy/tunnel images retain dependency license/notice files under `/licenses`
and module inventories. Upstream source is retrieved by pinned module checksum;
GATI's wrappers, patch and locks are in the exported repository. Changing source
versions or module sets requires renewed patch review, source/binary scans, licenses,
production tests and reproduction. Re-run audits before each deployment and review
upstream advisories regularly; today's successful scan is not a permanent clearance.

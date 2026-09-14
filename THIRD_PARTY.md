# Third-party material

Original GATI source and documentation: AGPL-3.0-or-later (LICENSE).

Dependencies are pinned in go.mod/go.sum and web/package-lock.json. Their upstream
licenses remain applicable: Go YAML v3 uses MIT/Apache-2.0 terms; Vite is MIT;
TypeScript is Apache-2.0. The client lockfile also records transitive packages;
complete license/SBOM verification is a release gate, not yet a completed audit.

The bootstrap downloads Go, Node and Docker Compose from their publishers. They
retain their own licenses. Tool binaries and their bundled notices remain outside
this repository. Container references are pinned by digest in deploy/images.env
and Dockerfiles; they contain separately licensed operating-system/runtime packages.

Release-helper formatting uses Ruff 0.11.13 (MIT), pinned in ruff.toml. Its binary
and bundled license remain in the ignored development environment; it is not an
application dependency or part of the deployed images. DEVELOPMENT documents setup.

Tirana public map data is now included under `data/tirana`, with OpenStreetMap
attribution, ODbL-1.0 terms, a source manifest, archived input and reproducible import.
See data/tirana/README.md. Geographic calculations use github.com/paulmach/orb
(MIT); its exact version/checksum is pinned in go.mod/go.sum.

Valkey access uses github.com/redis/go-redis/v9 (BSD-2-Clause), with pinned versions
and transitive checksums in go.mod/go.sum. No client-side state caching is enabled.

The map renderer MapLibre GL JS (6.9.0, BSD-3-Clause) and browser testing framework
Playwright (1.63.0, Apache-2.0) are pinned in web/package-lock.json. Chromium and
FFmpeg test artifacts remain in the user's Playwright cache with upstream terms.

Optional Web Push uses github.com/SherClockHolmes/webpush-go v1.4.0 (MIT), pinned
with its module checksum; upstream tag commit f5c3e9f7b642a8dd66cb844050520526f721971d.
It supplies RFC 8291 encryption and VAPID; GATI validates keys/endpoints before use
and supplies a bounded protected HTTP client. Transitives are explicitly pinned:
golang-jwt/jwt/v5 v5.3.1 (MIT), golang.org/x/crypto v0.57.0 and x/sys v0.48.0 (BSD-3-Clause).
The scratch image includes the pinned builder's CA certificate bundle for verified
provider TLS. These services are contacted only by optional, configured delivery.

`make security-check` uses pinned Go govulncheck v1.8.0. Its 2026-09-13 scan reported
no reachable or imported-package vulnerabilities. The module scan flagged
GO-2026-5932 for x/crypto/openpgp, which GATI does not import or link; the notification
library uses x/crypto/hkdf. This is not an independent security audit. Tool modules
are developer tooling, separate from the deployed application dependency graph.

Automated accessibility tests use @axe-core/playwright and axe-core 4.13.0
(MPL-2.0), pinned as development-only dependencies. They are absent from the
shipped browser bundle. Firefox/WebKit test binaries retain upstream licenses.
The optional Ubuntu 24.04 amd64 browser fallback pins Debian package versions
and SHA-256 values in scripts/toolchain/browser-deps-ubuntu24-amd64.json; extracted
libraries and their notices remain under ignored .runtime/browser-deps, not in
production artifacts. It does not change system libraries or the browser binaries.

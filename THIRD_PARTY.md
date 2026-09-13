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

Tirana public map data is now included under `data/tirana`, with OpenStreetMap
attribution, ODbL-1.0 terms, a source manifest, archived input and reproducible import.
See data/tirana/README.md. Geographic calculations use github.com/paulmach/orb
(MIT); its exact version/checksum is pinned in go.mod/go.sum.

Valkey access uses github.com/redis/go-redis/v9 (BSD-2-Clause), with pinned versions
and transitive checksums in go.mod/go.sum. No client-side state caching is enabled.

The map renderer MapLibre GL JS (6.9.0, BSD-3-Clause) and browser testing framework
Playwright (1.63.0, Apache-2.0) are pinned in web/package-lock.json. Chromium and
FFmpeg test artifacts remain in the user's Playwright cache with upstream terms.

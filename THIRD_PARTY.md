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

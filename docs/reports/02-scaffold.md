# Milestone 02 scaffold verification — 2026-09-13

Result: milestone 02 complete. Native checks, pinned image builds and running
Compose services were verified on 2026-09-13 after Docker installation.

## Implemented behavior

- Pinned, checksum-verified user-local Go/Node/Compose bootstrap; no sudo used.
- AGPL-3.0-or-later source/docs license, contributor/security/third-party guidance.
- Strict typed public YAML config, canonical JSON/SHA-256, production privacy floors
  and separately compiled simulation profile support.
- Read-only `/api/config` and `/healthz`, bounded HTTP transport settings, no-store
  responses, no cookies/access logging, Albanian generic errors, graceful shutdown.
- Albanian Vite/TypeScript preparation screen, first-party assets and config fetch.
  It requests no location, stores no participant data and has no willingness button.
- Default API Docker build excludes simulation. Pinned development Compose definition
  with isolated nonpersistent Valkey, no published store port and no API-store link yet.
- Local test/build/start commands and pinned CI workflow. No remote CI run or push.

## Validation executed

| Check | Result |
| --- | --- |
| Bootstrap then repeat bootstrap | Passed; pinned executable versions verified; repeat reused installed tools |
| `go test -race ./...` | Passed config/API/CLI tests |
| `go test -tags simulation ./...` | Passed with simulation build mode enabled |
| `go vet ./...` | Passed |
| `go test ./internal/config -fuzz=FuzzDecode -fuzztime=5s -parallel=2` | Passed, 351 generated executions plus seed corpus; bounded smoke fuzzing, not exhaustive verification |
| `npm --prefix web run check` | TypeScript strict check passed |
| `npm --prefix web run build` | Client production asset build passed |
| `npm --prefix web audit --audit-level=high` | Reported zero known vulnerabilities at check time; not a security audit |
| Production and simulation configuration checks | Passed; default executable rejects simulation config |
| `docker-compose -f compose.yaml config --quiet` | Definition parsed successfully without a daemon |
| `bash scripts/smoke.sh` | Native API health/config/hash/headers/method checks, absent participant/test routes, Albanian HTML and Vite API proxy passed |
| `bash -n scripts/*.sh` | Shell syntax passed |

Config tests reject unknown/duplicate/missing/null/aliased/coerced fields, extra YAML
documents, oversized input, excessive retention, invalid timing/geography, low
production thresholds, non-bucket alerts and integer overflow. They check that
changing activation count does not change arrival/nearby counts. They do not yet
prove matching/notification behavior, since those features are not implemented.

## Pending and limitations

Docker Engine 29.1.3 is now available to the agent without sudo. Both pinned API
and frontend images built successfully; Compose 5.5.1 started all three services.
`make check-containers` verifies the API, Albanian client, config proxy, Valkey PONG,
`save=""`, `appendonly=no`, `maxmemory-policy=noeviction`, no host store port and no
persistent writable store mount. This verifies the scaffold on this PC, not a clean
VM production deployment or participant storage (not implemented at this milestone).

No browser automation/visual review, arrival/matching tests, privacy-inference audit,
100k load test, independently reproducible build or production deployment has run.
No participant state is accepted or stored yet. The native workflow is a temporary
scaffold option; the intended app still uses the planned temporary Valkey store.

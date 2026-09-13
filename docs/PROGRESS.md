# Development progress and handoff

Updated: 2026-09-13. Status: milestone 02 native scaffold tested; container runtime
validation blocked on host administrator authentication.

## Current task and authorization

The user resumed implementation after documentation preparation. Installing needed
tools, implementing/testing milestones and making incremental local commits are
authorized. Continue once the Docker prerequisite is resolved. No push or deployment
is authorized, and GitHub authentication/write access remains unverified.

## Actual repository state

- Product requirements, plan, agent instructions and initial threat/decision records.
- Checksum-pinned user-local bootstrap for Go 1.27.1, Node 24.21.0/npm 11.19.0 and
  Compose 5.5.1. Source `scripts/env.sh` or use Make to select installed tools.
- Typed YAML configuration, production/simulation validation, canonical JSON/hash,
  Go config/health API, Albanian client preparation screen and native HTTP smoke.
- AGPL-3.0-or-later LICENSE and contributor/security/third-party guidance.
- Pinned Dockerfiles/development Compose and CI definition; container runtime checks
  and remote CI have not run. The API does not yet connect to Valkey.
- No willingness, matching, arrival, map, notification or participant-storage code.
  The configuration fields for those features describe future functional inputs.

## Milestone tracker

| Plan milestone | State | Evidence / next work |
| --- | --- | --- |
| 01 — Product/threat boundaries | Documented | Requirements, plan, decision 0001, THREAT_MODEL |
| 02 — Toolchain, scaffold, license, typed config | Partial; runtime blocked | Native checks passed; see reports/02-scaffold.md; need usable Docker Engine |
| 03 — Coarse geography and crossing fixtures | Not started | Reproducible map import, attribution and boundary/selection checks |
| 04 — Expiring willingness | Not started | Capability API, restricted store credentials/ACLs, retention/index and replay tests |
| 05 — Albanian willingness UI | Not started | Duration/radius/coarse area participation flow and browser network inspection |
| 06 — Seeded Tirana simulation | Not started | Isolated real-API driver and deterministic scenarios |
| 07 — Continuous activation and late admission | Not started | Matching/timers/crosswalk selection, races and joining |
| 08 — Arrival claims | Not started | Founding/late participants, nonces/freshness/retraction |
| 09 — Aggregate map/statistics | Not started | Canonical releases and inference review |
| 10 — Notifications and follows | Not started | Foreground, fake sink, expiring optional push and joins |
| 11 — Security and lifecycle hardening | Not started | Store/host/runtime controls and hostile-use scenarios |
| 12 — 100k benchmark | Not started | Reference hardware, real-time workload and capacity report |
| 13 — Deployment | Not started | Production stack, backup/restore/rollback and clean VM validation |
| 14 — Verifiable releases | Not started | SBOM, signed manifests and independent artifact comparison |
| 15 — Independent audit/pilot evidence | Not started | Resolve findings; no production-ready claim yet |

## Working commands

```sh
make deps
make config-check
make config-show
make config-check-simulation
make verify-local
make dev-native
```

`make verify-local` runs native tests/builds/config checks, Compose **parsing** and
an HTTP smoke. It does not start containers. `make dev-native` serves the preparation
screen at http://127.0.0.1:5173; Ctrl-C stops its API/client processes. It is not yet
the actual willingness flow. Read [DEVELOPMENT.md](DEVELOPMENT.md) for setup.

`make dev` / `make down` are implemented for Compose but require Docker Engine.
The planned `make simulate` and `make load` do not exist yet. A simulation-capable
config build is not the population simulator or fake-clock implementation.

## Validation and commits

[Milestone 02 report](reports/02-scaffold.md) records exact validation and limitations.
Use `git log` for authoritative commit IDs. The preparation commit is `a082b7a`;
toolchain installation was committed as `b9b362e`. Subsequent scaffold work is a
separate coherent local commit. Do not embed a commit's own hash in its files.

## Blocker and next action

On this Ubuntu 24.04.3 PC, there is no Docker daemon/CLI and no rootless UID-map
helpers. `sudo -n true` requires a password. User-local tools were installed, and
independent native scaffold work is complete. The next milestone acceptance step
requires administrator-authenticated Docker installation/access. Never request
passwords in chat; the user can follow the exact commands in DEVELOPMENT.md.

Once `docker info` works in the agent's shell:

1. Recheck git state and tool versions.
2. Run `make dev` / container build and startup checks; check API, Vite config proxy
   and Valkey health/persistence settings. Fix failures before marking 02 complete.
3. Start milestone 03: map import/attribution, coarse grid and crossing index/tests.
4. Update this tracker/report and make incremental commits; do not push without
   explicit authorization.

Keep secrets and real participant data out of this file. Record failures, skipped
checks and remaining work, rather than treating an intended control as implemented.

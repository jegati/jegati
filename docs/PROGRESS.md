# Development progress and handoff

Updated: 2026-09-13. Status: milestone 02 in progress; user-local toolchain installed.

## Current task and authorization

The user resumed implementation after the documentation preparation. Installation
of development/testing tools, milestone implementation/testing and incremental
local commits are authorized. Continue independent work until a credential or
material product decision is needed.

No remote push or deployment is currently authorized. No credentials have been
requested or stored. Do not infer GitHub write access from the configured remote.

## Actual repository state

- AGENTS.md, requirements summary, development plan, initial decision/threat
  records, this tracker and a read-only toolchain doctor are present.
- No application source, storage schema, Makefile, package manifests, runtime
  configuration, containers or application tests exist yet.
- Source licensing remains a milestone 02 deliverable; the plan proposes
  AGPL-3.0-or-later. This preparation does not claim an existing license grant.
- Use `git log` for authoritative commit hashes; do not embed a commit's own hash
  in files that are part of it.

## Milestone tracker

| Plan milestone | State | Evidence / next work |
| --- | --- | --- |
| 01 — Product/threat boundaries | Documented | REQUIREMENTS, plan, decision 0001, THREAT_MODEL; planned claims distinguished from evidence |
| 02 — Toolchain, scaffold, license, typed config | In progress | Pinned user-local tools installed; scaffold/config work next; Docker Engine blocked on host administrator authentication |
| 03 — Coarse geography and crossing fixtures | Not started | Reproducible map import, attribution and boundary/selection checks |
| 04 — Expiring willingness | Not started | Capability API, retention/index and replay tests |
| 05 — Albanian willingness UI | Not started | Accessible duration/radius/coarse area flow and network inspection |
| 06 — Seeded Tirana simulation | Not started | Isolated real-API driver and deterministic scenarios |
| 07 — Continuous activation and late admission | Not started | Matching/timers/crosswalk selection, races and joining |
| 08 — Arrival claims | Not started | Founding/late participants, nonces/freshness/retraction |
| 09 — Aggregate map/statistics | Not started | Canonical releases and inference review |
| 10 — Notifications and follows | Not started | Foreground, fake sink, expiring optional push and joins |
| 11 — Security and lifecycle hardening | Not started | Enforced store/host/runtime controls and hostile-use scenarios |
| 12 — 100k benchmark | Not started | Reference hardware, real-time workload and honest capacity report |
| 13 — Deployment | Not started | Production stack, backup/restore/rollback and clean VM validation |
| 14 — Verifiable releases | Not started | SBOM, signed manifests and independent artifact comparison |
| 15 — Independent audit/pilot evidence | Not started | Resolve material findings; no production-ready claim yet |

## Environment observations — recheck before relying on them

On 2026-09-13 the shell reported Ubuntu 24.04.3 LTS, Git, curl and make available.
Initially Go, Node/npm, Docker and GitHub CLI were not on PATH. Go 1.27.1,
Node 24.21.0/npm 11.19.0 and standalone Compose 5.5.1 are now installed under
`~/.local/share/gati/tools`; source `scripts/env.sh` to select them. `sudo -n true`
reports that a password is required; Docker Engine and rootless UID-map helpers
remain absent. These are observations
of the agent's execution environment, not proof of everything installed on the PC.

The repository remote is `https://github.com/jegati/jegati.git`, branch `main` at
the time of preparation. Push authentication and write permission are unverified.

Prefer verified user-local toolchains where practical. A root-requiring operation
may need the user to run a specific command in their terminal. Never ask for their
password in chat. Do not stop independent work merely because Docker needs setup.

## Available checks and planned commands

Available now:

```sh
bash scripts/doctor.sh
git diff --check
```

Doctor reports command availability/basic versions and checks Docker Compose when
the CLI exists. Exit 1 means required tools are missing/unusable. It deliberately
does not start daemons, install packages, inspect credentials or certify the app.

The plan's `make dev`, `make test`, `make simulate`, `make load`, `make verify-local`,
`make config-check` and `make config-show` are **not implemented yet**. Add commands
to the README only when they work, or label them explicitly as future targets.

## Next implementation action

Finish the milestone 02 scaffold/config tests independently of Docker, then record
container checks as pending. The user must authenticate an administrator setup
step before full Compose runtime testing; see DEVELOPMENT.md for exact commands.

## Toolchain validation — 2026-09-13

- `bash scripts/bootstrap.sh`: installed pinned artifacts after SHA-256 verification.
- With `source scripts/env.sh`: `go version`, `node --version`, `npm --version`
  and `docker-compose version` passed with the pinned versions above.
- Docker Engine is absent; no daemon or container startup claim is made.

## Preparation validation — 2026-09-13

- `bash -n scripts/doctor.sh`: passed shell syntax validation.
- `bash scripts/doctor.sh`: exited 1 as expected; Git/curl/make available,
  Go/Node/npm/Docker missing from PATH, optional GitHub CLI missing.
- Local Markdown file-link, code-fence and whitespace inspection: passed for all
  seven Markdown files. No external link availability or application behavior was
  tested by this inspection.
- `git diff --check`: passed. No application tests exist or were claimed as run.

## Recording subsequent work

After each milestone or interruption, update its state and record: changed
behavior/files, exact tests and results, skipped checks, unresolved concerns and
the next concrete action. Link reports rather than dumping command logs. Do not
store credentials, host secrets, real participant data or transient authentication
failures containing sensitive details in this file.

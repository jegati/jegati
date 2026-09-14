# GATI 🦩

**Bëje vullnetin tënd të dukshëm.**

An anonymous, temporary expression of willingness to join collective real-world activity.

**JAM GATI → JEMI GATI → JAM KËTU → JEMI KËTU**

## Project status

**The [runtime remediation](docs/reports/runtime-remediation-2026-09-14.md) is locally validated:**
pinned source builds and exact-image audits address the review blockers, with two
scoped unused-code exceptions. Public launch still needs independent review and
the actual host, Cloudflare and device checks. See [runtime auditing](docs/RUNTIME_SECURITY.md).

See the [functionality and test-coverage audit](docs/FUNCTIONALITY_STATUS.md).
The operating default is 100 m cells with a 50 m maximum reported device error.

Milestones 02–08 are complete: configuration, coarse Tirana geography, expiring
willingness APIs and the Albanian browser participation flow. The UI supports
one-action JAM GATI with required one-shot device location, availability/radius choices,
reload recovery and cancellation. Continuous matching, automatic crossroad
invitations, going/decline, late admission and temporary arrival confirmation are
implemented. **JAM ENDE KËTU** allows explicit renewal near presence expiry,
with a fresh device fix and the original session/gathering deadlines preserved.
Ended gatherings are labeled and removed from the gathering map layer; expired
metadata is pruned. See [lifecycle boundaries](docs/decisions/0014-bounded-presence-renewal.md).
Delayed aggregate statistics and a cell-only public activity map now
work locally, including explicit joining from map entries. Public cells default to
1 km; private matching remains 100 m. Eligible users can preview a private destination
before committing. [Optional background push](docs/NOTIFICATIONS.md) is implemented
with temporary opt-in/resume and remains disabled by default; real provider/device
verification and area follows remain. The Albanian **Rreth nesh** section explains
peaceful Flamingo protest support, voluntary coordination and bounded privacy
claims, with a source link and no added enrollment step. Seeded isolated [Tirana simulations](docs/SIMULATION.md)
include a 3,000-person response/journey driver, a standalone map replay,
expiry/replay checks and smaller Sybil cases. Run `make simulate-population` using
your current configuration, or `make simulate-suite` for a scenario comparison.
See the [3,000-person findings and replay links](docs/reports/population-3000.md)
and [matching configuration reference](docs/MATCHING_PARAMETERS.md).

**Known privacy limitation:** a reproducible synthetic collusion test reconstructs
one participant's coarse cell from the ordinary nearest-intersection invitation.
See the [evidence and product decision](docs/reports/09-inference-gate.md).
The prototype does not yet meet the full inference-protection requirement and is
not ready for public deployment. The user retained nearest-crossroad matching with
this limitation and explicitly accepted documented public inference for local
implementation (see [decision 0007](docs/decisions/0007-cell-only-public-activity.md)).

Read the [local development guide](docs/DEVELOPMENT.md) for setup and working commands.

Read the [development and verification plan](docs/DEVELOPMENT_PLAN.md) for the proposed architecture, 15 implementation commits, local Tirana simulation, privacy/security boundaries, and deployment verification process.

The intended MVP is entirely Albanian, requires no account, and displays collective activity rather than individual participants. Local development and testing come first; public deployment depends on the evidence and release gates in the plan.

## Contributor starting points

- [Audit starting point](docs/AUDIT.md): request-to-storage/test map and verification gates.
- [Agent instructions](AGENTS.md): workflow, autonomy, product and privacy invariants.
- [Progress and handoff](docs/PROGRESS.md): actual implementation state, checks and next action.
- [Agreed requirements](docs/REQUIREMENTS.md): product scope and conversation corrections.
- [Initial architecture decisions](docs/decisions/0001-mvp-boundaries.md): choices and consequences.
- [Threat model](docs/THREAT_MODEL.md): intended controls, residual risks and evidence needed.

Check local prerequisites without changing the computer:

```sh
bash scripts/doctor.sh
```

Missing required tools produce exit status 1. The checker does not install tools or
verify Docker daemon access. To run the willingness flow locally:

```sh
bash scripts/bootstrap.sh
make deps
make verify-local
make dev
```

Open http://127.0.0.1:5173. Docker is required for the temporary store.
Public snapshots appear after the configured delay (normally 5–10 minutes) and
small groups remain suppressed. Matching and personal state continue immediately.
`make dev-native` remains a read-only UI/API preview without a store.
Run `make browser-install` once, then `make test-browser` with the Compose API up.

## License

Original source and documentation are licensed under [AGPL-3.0-or-later](LICENSE).
See [third-party notices](THIRD_PARTY.md) for dependency and future map-data terms.

Selected usability work is tracked in [the implementation plan](docs/USABILITY_IMPLEMENTATION_PLAN.md).
Private destination preview before a map join and clearer waiting/offline/arrival-expiry
guidance are implemented. Previewing or cancelling a preview never enrolls you.

Additional local validation: [automated testing and monitored load](docs/AUTOMATED_TESTING.md), [private operational monitoring](docs/MONITORING.md). These checks create disposable synthetic stores and do not deploy the app.

[System architecture, diagram and 1M-signal estimate](docs/ARCHITECTURE.md) distinguishes measured 100k evidence from untested scaling assumptions.

[Deployment preparation and scaling priorities](docs/DEPLOYMENT_PREPARATION.md)
records the first-host topology, Cloudflare trust boundary and remaining release gates.
[Measured optimizations](docs/reports/scaling-2026-09-14.md) include repeated 10k–1M
component benchmarks and a monitored 100k mixed-load comparison, including regressions.
[Production images, local replica tests and the release runbook](docs/DEPLOYMENT.md)
are available; no public deployment has been performed. Run `make test-deployment`
for the isolated production proxy/browser/worker rehearsal.

# GATI 🦩

**Bëje vullnetin tënd të dukshëm.**

An anonymous, temporary expression of willingness to join collective real-world activity.

**JAM GATI → JEMI GATI → JAM KËTU → JEMI KËTU**

## Project status

Milestones 02–04 are complete: configuration, coarse Tirana geography and expiring
willingness APIs are implemented and tested against real Valkey. The browser is
still a preparation screen; its participation flow is the next milestone. Matching,
arrivals, collective maps and notifications are not implemented yet.

Read the [local development guide](docs/DEVELOPMENT.md) for setup and working commands.

Read the [development and verification plan](docs/DEVELOPMENT_PLAN.md) for the proposed architecture, 15 implementation commits, local Tirana simulation, privacy/security boundaries, and deployment verification process.

The intended MVP is entirely Albanian, requires no account, and displays collective activity rather than individual participants. Local development and testing come first; public deployment depends on the evidence and release gates in the plan.

## Contributor starting points

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
verify Docker daemon access. To run the current scaffold without Docker:

```sh
bash scripts/bootstrap.sh
make deps
make verify-local
make dev-native
```

Open http://127.0.0.1:5173. This is a preparation screen, not the full participation
flow. Use `make dev` once Docker is available to validate the Compose stack.

## License

Original source and documentation are licensed under [AGPL-3.0-or-later](LICENSE).
See [third-party notices](THIRD_PARTY.md) for dependency and future map-data terms.

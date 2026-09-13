# Contributing to GATI

Start with [AGENTS.md](AGENTS.md), [the requirements](docs/REQUIREMENTS.md) and
[the progress tracker](docs/PROGRESS.md). Set up the pinned tools using
[DEVELOPMENT.md](docs/DEVELOPMENT.md). Keep changes small and include behavioral
validation plus documentation of changed data flow or retention.

Use synthetic inputs only. Do not add analytics, permanent identifiers, personal
fixtures or participant debugging interfaces. User-facing copy is Albanian.
Formatting: `gofmt` for Go; TypeScript strict checking and the existing client style.
Run `make verify-local` for the current scaffold and document skipped container or
browser checks. Later milestone-specific checks will be added as features exist.

Project source and original documentation are licensed under AGPL-3.0-or-later;
contributions are made under that license. Third-party dependencies and future map
extracts keep their own licenses and attribution. See LICENSE and THIRD_PARTY.md.

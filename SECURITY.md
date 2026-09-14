# Security status and reporting

GATI is under development and has no supported public production release. Read
[THREAT_MODEL.md](docs/THREAT_MODEL.md) for planned controls and unresolved risks.
No anonymity, independent-human or deployment-honesty guarantee is claimed.

The [2026-09-14 review](docs/reports/security-2026-09-14.md) holds public deployment
pending production runtime dependency remediation. Local functional test success
does not clear those findings. Run `make security-check` for Go/npm audits and scan
all exact release images as described in the review before publishing a release.

A private reporting channel has not yet been configured. Do not submit credentials,
real participant records or exploitable details about a live deployment in a public
issue. For this local prototype, report synthetic reproductions to the repository
owner through an existing private channel. A documented working private reporting
channel is required before the public pilot; this file does not invent an address.

Security fixes must include an appropriate regression test and an updated claim or
retention description where relevant. Never use real participants as test subjects.

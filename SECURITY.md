# Security status and reporting

GATI is under development and has no supported public production release. Read
[THREAT_MODEL.md](docs/THREAT_MODEL.md) for planned controls and unresolved risks.
No anonymity, independent-human or deployment-honesty guarantee is claimed.

The [2026-09-14 review](docs/reports/security-2026-09-14.md) identified runtime
issues addressed by [pinned source builds and fresh image evidence](docs/reports/runtime-remediation-2026-09-14.md).
Two unused-code package exceptions expire 2026-10-14. Run `make security-check`,
`make test-runtimes` and `make audit-images` before releasing; public activation
requires a passing audit no older than 24 hours. Independent/host/edge checks remain
separate gates. See [runtime security](docs/RUNTIME_SECURITY.md).

A private reporting channel has not yet been configured. Do not submit credentials,
real participant records or exploitable details about a live deployment in a public
issue. For this local prototype, report synthetic reproductions to the repository
owner through an existing private channel. A documented working private reporting
channel is required before the public pilot; this file does not invent an address.

Security fixes must include an appropriate regression test and an updated claim or
retention description where relevant. Never use real participants as test subjects.

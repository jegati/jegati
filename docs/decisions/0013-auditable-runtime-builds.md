# 0013 — Pinned source runtimes and release vulnerability gate

Date: 2026-09-14. Status: accepted for local implementation.

The user authorized remediation of the final review findings. Latest consumable
Caddy/cloudflared images still contained old dependencies. Cloudflared source also
initialized a third-party Sentry client in tunnel/access paths, incompatible with
GATI's no-tracking intent even when Docker logs are disabled.

Build Caddy 2.11.4 as a separate locked Go module importing only the modules needed
by the reviewed Caddyfile. Build cloudflared from commit
f11dea9cb7079e90a982c1a2d5548ab40847fdcf with verified module content hashes, tracked
Go locks and a small checked source patch that disables all four reviewed Sentry
initializers. Unexpected initializer locations/counts fail the patch. Tests call
the patched initialization with a synthetic DSN and verify that no client/event
exists. The SDK can remain a dependency; GATI never initializes it. Retain the
normal named-tunnel protocol and upstream dependency replacements. Do not claim
an unmodified Cloudflare-supported binary: its version is 2026.9.1-gati.1.

Both runtimes use the same pinned Go 1.27.1 builder as the application, upgraded
x/crypto, x/net, x/text, gRPC and compression dependencies, and scratch final images.
Ship executable, CA roots, configuration/assets and dependency license/notice files;
no shell, curl, libc or package manager is needed. Caddy's fixed loopback health
command replaces wget; it accepts only HTTP 200 and does not follow redirects.
Container read-only/capability/network/logging/TTL boundaries stay in place.

Retain Go symbols in all three Go executable images. govulncheck 1.8.0 explicitly
falls back to module-level findings on stripped binaries; its printed symbol
results in that mode are not evidence of actual linkage. Larger artifact size is
accepted to improve exact-binary auditability. No runtime throughput improvement
is claimed from changing symbol retention.

CEL 0.29/0.30 is not source-compatible with Caddy 2.11.4's interpreter integration.
Keep 0.28.1 with an exact-version, web-only, 30-day exception for unused NativeTypes /
ParseStructTag functions. The other exception covers unimported OpenPGP in x/crypto
0.57.0. Neither exception overrides a failed exact-binary symbol scan. Fresh source
and image evidence must support these exceptions; any newly affected symbol is a
failure. Revisit CEL with the next compatible Caddy release instead of maintaining
an unreviewed interpreter fork. Excluded forward_auth, templates, tracing and other
unused Caddy modules are not available for configuration without a source change.

The audit command checks immutable release files/image IDs, scans every final
image with pinned Trivy, retains all severities and per-image CycloneDX SBOMs,
extracts Go executables and checks their hashes and vulnerability symbols. New,
expired or mismatched exceptions fail. Public activation requires a passing audit
for that manifest and those binaries, no older than 24 hours; host preflight is
still required. An audit file is local operational evidence, not signed remote
attestation or protection from a dishonest host/operator.

Reproducibility now compares API, Caddy, cloudflared and web assets from two clean
exports, and the release manifest records all three executable hashes. This is
same-host/toolchain evidence; independent reviewer/host checks remain separate.
Source Go locks, patch, source checksums, licenses and build instructions are public
review material. No external deployment or account setup is authorized by this ADR.

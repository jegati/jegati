# 0012 — Production ingress, roles and verifiable releases

Date: 2026-09-14. Status: implemented locally; public edge/host validation pending.

Keep one combined Go process as the default. Optional API-only replicas refresh
destinations, while a dedicated worker runs matching, publication and cleanup.
The existing lease and atomic store transitions retain one matching owner; this
is process separation and API failover, not geographic parallel matching.

Use built browser assets in Caddy and private Compose networks. An exact isolated
connector peer supplies Cloudflare's client address; Caddy freezes it before
stripping forwarding headers. The Go boundary validates the exact Caddy peer and
one canonical client IP. Direct local mode ignores forwarded headers. Credential-
bearing public-route responses are always no-store. Literal Caddy route ordering
enforces ingress checks before all handlers. Loopback health exceptions cannot
access participant routes.

Drop all container capabilities, remove Caddy's upstream file capability, disable
logs/core dumps/container swap and use read-only filesystems with temporary RAM
mounts. Metrics remain a private Unix socket read through docker exec. None of this
proves host administrators cannot inspect memory; public launch needs host review.

Publish immutable source/file/config and image identities plus browser asset hashes.
Keep operational secrets and Valkey configuration at stable mounts outside release
directories so an API release does not unnecessarily restart the memory-only store.
No participant backups. Compatible rollback requires an explicit previous release
and config compatibility, followed by running-container verification. Checksums
and two local clean builds aid inspection; they do not attest remote behavior.

Cloudflare remains a trusted plaintext TLS intermediary with provider-side metadata
processing. The supported topology uses a named outbound tunnel; no deployment
ports are exposed on the host. Real edge rules, provider controls, devices and public
publication require later account access and authorization.

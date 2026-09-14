# jamgati.com alpha deployment package

Canonical origin: **https://jamgati.com**. Initial topology: one combined API/worker,
private Valkey, Caddy and the named Cloudflare tunnel on OVHcloud. Keep functional
settings in config/gati.yaml; this package changes no thresholds, retention or
capacity setting. Public push remains off. Host sizing is still unverified.

## Prepare on the development machine

Commit the reviewed source first, then use a new directory for each artifact:

```sh
source scripts/env.sh
make publication-check test-publication
python3 scripts/release.py prepare --release reports/local/jamgati-alpha --public-host jamgati.com
make audit-images RELEASE=reports/local/jamgati-alpha OUTPUT=reports/local/jamgati-alpha-audit
python3 scripts/rehearse-release.py --release reports/local/jamgati-alpha --output reports/local/jamgati-alpha-rehearsal
```

The rehearsal now reads the hostname from the checksummed deployment.env and uses
it in the test-only connector. Browser assets/config verification still connects to
loopback. It tests the configured origin's local routing without changing public
DNS or contacting jamgati.com. It does not verify Cloudflare, public TLS or OVH.
Retain the release archive, hashes, exact audit/SBOMs and rehearsal evidence together.
Rescan the final artifact within 24 hours of public activation.

## Cloudflare configuration to review and apply after access

Use a named account-managed tunnel. Route jamgati.com to http://web:8080 through
the isolated connector. Use a terminal catch-all 404 for other hostnames. Leave
origin inbound web/API/Valkey ports closed; do not publish the VPS origin address.
Use the exact cache/query/authentication rules in DEPLOYMENT.md. A conservative
initial choice is to cache only immutable /assets/ responses and bypass all /api/
responses at the edge. Application aggregate releases remain precomputed. Respect
origin expiry, and never enable stale serving for API responses.

Enable HTTPS redirection and the available network/edge DDoS controls. Keep precise
application rate limits active; test carrier/NAT effects before tightening edge
rules. Disable injected analytics, beacon scripts, content rewriting, session replay
and exported access logs. Provider-internal processing is outside application TTL.
The available zone rules/features must be verified in the actual account rather
than assumed from a local template or historical plan pricing.

Before activation, test asset MISS/HIT, API/auth/cookie/query bypass, expired/error
responses, spoofed forwarding headers, direct-origin denial, HTTPS location
permission, cold starts, tunnel recovery and served hashes. Do not load-test the
public edge without a separately bounded, authorized test. No account automation
or remote operations are performed by this preparation package.

## Operator monitoring and activation

See MONITORING.md for the prepared systemd collector and optional alert sender.
Choose and privately configure the delivery endpoint. Establish an independently
observed host/tunnel availability check: an alert process on a dead VPS cannot send.
No such external service is configured or claimed here.

Transfer only the verified release and separate private operational credentials.
Use /opt/gati/current as a symlink to the selected immutable release for the service
templates; update it deliberately when changing releases. Templates currently name
the default combined API container. A two-API/dedicated-worker topology needs one
collector per expected process and corresponding alert inputs; do not silently
monitor only one replica. Review actual root/Docker access on the host.

Run host preflight and the activation/verification commands in DEPLOYMENT.md only
after VPS/SSH and Cloudflare credentials are privately available. Confirm no
swap/hibernation/core dumps, participant backups or access/body logs. Keep the
previous compatible release and operational secrets for rollback, never participant
state. The alpha notice/feedback instructions are in the app; GitHub source and
issue template still need publication. There is no separate volunteer stage.

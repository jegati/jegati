# Ephemeral storage schema — willingness (schema 1)

All keys are private Valkey keys; no public participant listing exists. The store
has no RDB/AOF persistence or writable disk volume. Values and command arguments
must not enter logs. Service identities are infrastructure credentials, not users.

| Key | Contents | Lifetime |
| --- | --- | --- |
| `gati:s:<capability-sha256>` | Coarse cell, radius, chosen duration, server creation/deadline in milliseconds, state `gati` | Chosen duration (30..120 minutes); deleted on cancellation |
| `gati:cap:<capability-sha256>` | Constant `used` replay tombstone; no location | Original deadline, including after cancellation; prevents delayed retries resurrecting a cancelled signal |
| `gati:expiry` | Sorted members `<hash>\|<cell>` scored by deadline | Key expires at latest member deadline; expired members pruned in bounded batches |
| `gati:cell:<canonical-cell>` | Sorted capability hashes scored by deadline | Same latest-deadline TTL; cancellation removes membership; cleanup prunes expiry |
| `gati:rate:secret:<epoch>` | Random shared HMAC secret | Remainder of a ten-minute epoch |
| `gati:rate:requests:<epoch>:<hmac>` / `creates:...` | Aggregate network admission counters | Configured fixed window, at most ten minutes; no sliding renewal |
| `gati:rate:writes` | Global write counter | One second |

Every status read checks Valkey server time in addition to native expiry. Create
and cancellation are Lua-atomic; identical create retries return the original
deadline. A used capability cannot mutate its initial area/radius/duration or
resurrect cancellation before its deadline. After expiry, the client must generate
a fresh random capability; no permanent registry tracks reuse across sessions.

The API currently handles index cleanup on a configured reconciliation ticker;
multiple replicas can call the idempotent bounded cleanup script. At most ten
configured batches run per tick. Under backlog, logical expiration remains immediate
but index-removal latency is a benchmark concern. The later dedicated worker will
own matching/publication. Do not claim bounded physical deletion under arbitrary DoS.

Network keys use the socket peer (IPv4 address / IPv6 /56 prefix); untrusted forwarded
headers are ignored. A Vite proxy therefore shares one limiter identity locally.
The short-lived HMAC is sensitive pseudonymous data, and live store compromise can
expose its current secret. No network key is stored in a signal or response.

## Roles and secrets

`scripts/init-secrets.py` creates random infrastructure passwords under ignored
`.runtime/` (host directory mode 0700) and writes a hashed-password Valkey ACL file.
Individual read-only file mounts are readable by container service UIDs; they are
not environment-variable values or command-line arguments. Never commit or publish
`.runtime/`, and recreate services when changing mounted ACLs/credentials.

Default unauthenticated access is disabled. `app` can use only the commands needed
by the store, restricted to `gati:*`; it cannot save/snapshot, monitor, inspect ACLs,
change configuration or access other keys. `health` can ping and read configuration
for runtime checks but cannot read participant keys. This is service-role separation,
not protection against a privileged host operator.

Valkey ACL logs, slow command logging and latency monitoring are disabled to avoid
recording participant command arguments. Core dumps and container swap are disabled
for the store container. Host/provider capture and malicious privileged access remain
outside those controls. Positive memory pressure rejects new writes (`noeviction`).

## Tests

`make test-store` starts and removes a disposable Valkey with random loopback port,
private service ACLs and only synthetic inputs. It never uses the running app store.
It checks native TTL expiry, index cleanup, cancellation/replay, concurrent idempotency,
rate counters, denied privileged commands and the real HTTP willingness contract.
Default unit tests explicitly skip these integrations; they do not fake a passing
store test. The readiness/build/native checks are separate from this target.

Simulation builds compile a separate `gati-sim:*` namespace and require separately
generated ACL credentials. Their frozen clock affects functional deadlines only;
real native TTLs still bound the disposable store. See SIMULATION.md. Production
creation retries now explicitly check stored deadlines as well as native TTL.

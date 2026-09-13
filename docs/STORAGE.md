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

## Matching transaction groundwork (milestone 07, not yet scheduled)

Signals can contain private `_pending`, `_pending_until`, `_gathering`,
`_gathering_until` and `_declined` bookkeeping. The own-session API removes these
fields; the upcoming invitation view will expose only the selected gathering.
`pending:<id>` holds a bounded founding hash list and public intersection while the
stability timer runs; its TTL covers stability plus bounded recovery. The `pending`
sorted index uses ready-at scores and a TTL no longer than its latest reservation.
`destination:<intersection-id>` prevents duplicate use of that exact intersection while
reserved/open, with the corresponding reservation/gathering lifetime.
Activation deletes the pending record/index member and sets `gathering:<id>` plus
a deadline-scored `gatherings` index, with TTL at the fixed end (maximum 60 minutes).
No permanent founder/member history is retained in the gathering record.
Every activation rechecks all surviving founders and the configuration version.
These records are not public APIs. Worker sweeps/recovery/admission still need wiring.

The worker is now scheduled. `dirty` is a coalesced ten-second flag; `worker:lease`
is an expiring random process token, unrelated to participants. Worker snapshot
pages are transient input; the between-run cache contains only gathering
(destinations/deadlines) records. Founding links remain internal. Decline maps and
`_invite_after` have the signal's original TTL and configured cardinality/cooldown
bounds. Going/decline/switch never extend that TTL. Atomic direct join validates a
live reachable destination before creating willingness, with no orphan signal on
rejection. No arrival/member counts exist yet.

## Temporary arrival lifecycle

`arrival-nonce:<capability-hash>` contains nonce hash, gathering ID, expiry and used
flag, at most 120 seconds and never beyond session/gathering expiry. A used nonce
is a short-lived replay tombstone; cancellation removes it. It cannot restore a
retracted arrival. Arrival confirmation stores only an expiry and private member
reference in the session, with no arrival coordinates or coarse-cell history.
`arrivals:<gathering-id>` is an internal sorted set of capability-hash/nonce-hash
members scored by freshness deadline (maximum 15 minutes, bounded by session and
gathering). One credential has at most one live member. Removal on cancel, decline,
switch and retraction is atomic; deadline-aware reads ignore expired scores.

`presence:<gathering-id>` holds only the temporary stable confirmation cohort,
with TTL bounded by stability plus twice the reconciliation interval and by the
gathering end. The successful cohort is deleted once JEMI KËTU forms. A drop below
the configured current-arrival threshold resets the state; subsequent recovery
requires another stability interval. Sorted-set cleanup runs in bounded batches;
its key cannot outlive the gathering. Own-session reads also clear expired arrival
metadata. Physical cleanup delays under overload require the hardening/load tests.

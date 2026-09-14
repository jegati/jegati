# Auditing the Valkey transactions

The store keeps each Lua transaction beside its Go caller. Read the caller's
`KEYS`/`ARGV` construction together with the script: positional arguments are part
of the private storage contract, not HTTP input. [STORAGE](STORAGE.md) defines key
lifetimes and [AUDIT](AUDIT.md) links the surrounding request/worker flows.

| Source / transactions | Checks and changes to inspect | Regression tests |
| --- | --- | --- |
| [store.go](../internal/store/store.go): `create`, `status`, `cancel`, `cleanup` | Immutable coarse claim, capacity, replay tombstone, logical expiry, arrival removal and cell/expiry index cleanup | `TestRealStoreExpiryCancellationAndIdempotency`, `TestRealStoreConcurrentCreationAndACL` |
| [store.go](../internal/store/store.go): `limit`, `networkSecret` | Bounded rate counters and temporary shared network secret; network rate windows use real store time even in simulations | [network tests](../internal/store/network_test.go) |
| [gatherings.go](../internal/store/gatherings.go): `reserve`, `activate` | Cohort claims, deadlines, configuration/ownership checks and stability; activation freezes destination/deadline | [reservation/activation races and configuration rollover](../internal/store/gatherings_test.go) |
| [gatherings.go](../internal/store/gatherings.go): `snapshotPage`, `readGathering`, `readClock`, `lease`, `consumeDirty` | Bounded reads, stale state removal and worker ownership | [snapshot tests](../internal/store/snapshot_test.go), gathering tests |
| [intent.go](../internal/store/intent.go): `intent`, `createAndJoin` | Live claim, reachability inputs, admission cutoff, decline budget/cooldown, retries, removal of an old arrival; join reuses the same creation body | [fixed-clock HTTP journey](../internal/httpapi/gathering_simulation_test.go) |
| [arrivals.go](../internal/store/arrivals.go): `issueNonce`, `confirmArrival`, `retractArrival`, `reconcilePresence` | One-use challenges, current intent/claim, arrival expiry, mode/member-bound early renewal without double counting, bounded presence index, stability and threshold transitions | [concurrent replay and seeded transition model](../internal/store/arrivals_test.go), [renewal races](../internal/store/arrival_renewal_test.go), [renewal HTTP deadlines](../internal/httpapi/arrival_renewal_simulation_test.go) |
| [activity.go](../internal/store/activity.go): `stageActivity`, `latestActivity` | Publication fencing, immutable releases and deadlines; privacy policy is computed in `internal/activity`, not invented by Lua | [activity fencing, expiry and bounded reads](../internal/store/activity_test.go) |
| [push.go](../internal/store/push.go): `registerPush`, `dropPush`, `claimPush`, `finishPush`, `pushStatus` | Expiring binding/revision, capacity, cancellation, worker claim ownership, bounded retries and no resurrection after removal | [push lifecycle/expiry](../internal/store/push_test.go), [delivery tests](../internal/notification/service_test.go) |
| [changes.go](../internal/store/changes.go): `takeCells`, `matchingLeaseState`; [lag.go](../internal/store/lag.go): `oldestLag` | Bounded cell invalidation, owner changes and aggregate lag reads; taking a batch must not erase later writes | [cell/lease races](../internal/store/changes_test.go), [lag tests](../internal/store/lag_test.go) |

Expired gathering links are cleared by [expired_links.go](../internal/store/expired_links.go)
inside status reads and bounded matching snapshot batches. The transaction rereads
current assignments and keeps session TTLs. [Keeper/churn tests](../internal/store/expired_links_test.go)
cover pruning the shared gathering index and preserving a newer assignment.

## Shared fragments and clocks

`newScript` in [store.go](../internal/store/store.go) prepends `removeArrivalLua`
and `dirtyCellLua`, substitutes `__CLOCK__`, then selects the key namespace.
These helpers are part of the same atomic execution, not separate requests.
`createAndJoin` inserts `createLua` inside a local function; keep its early returns
inside that function when editing it.

The normal build uses [clock_production.go](../internal/store/clock_production.go)
and the `gati:` namespace. The simulation build uses
[clock_simulation.go](../internal/store/clock_simulation.go), a bounded explicit
clock and `gati-sim:` keys. Native key TTLs still run on real time. `networkSecret`
intentionally uses `redis.NewScript` directly and a real-time clock. A layout
cleanup must preserve these distinctions, every command's order, and all return
values. A changed source layout changes script SHA hashes; Valkey's normal script
loading handles that, without a schema or data migration.

## Verification

`make test-store` exercises the real disposable Valkey with the race detector.
`make simulate` additionally executes the simulation clock/namespace and full
HTTP matching/admission/arrival journey. Plain `make test` skips environment-gated
store integrations.

The September 14 layout pass used StyLua 2.3.0 (`--syntax lua51 --verify`, 100-column
width, two spaces, `AutoPreferSingle`). The clock and nested creation fragments
were temporarily represented as placeholder calls for parsing, then restored.
All 34 fragments also passed an old/new token comparison, ignoring whitespace,
comments, optional semicolons/trailing table commas and equivalent unescaped quote
styles. This is supplementary regression evidence, not a formal semantic proof.
The formatter was installed locally from its publisher's Linux x86_64 archive,
verified against SHA-256
`4c06b5963b8e832b51ebafc8051bab90ee3322e51d2f5ea59f4eacae78ce8bfc`;
it adds no build/runtime dependency. Future edits should remain readable in place
and run the relevant transaction and journey tests.

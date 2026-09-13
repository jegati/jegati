# 0011 — Incremental matching reads with bounded reconciliation

Date: 2026-09-14. Local implementation; deployment and benchmark results are
recorded separately in PROGRESS.md.

## Decision

Retain the existing private per-cell sorted sets. Relevant creation, cancellation,
intent, reservation and activation transitions atomically mark their coarse cell
in `gati:dirty-cells`. The marker score is the last-change server time. Markers
expire logically after ten seconds; writes prune older markers and set the key's
TTL to ten seconds. The validated finite service grid bounds membership/pruning
work independently of the number of participants. The existing global dirty hint
continues to wake matching; neither queue is a public endpoint.

The matching owner atomically takes and removes a batch BEFORE reading cells.
Changes arriving during processing therefore enter another batch and cannot be
removed by an acknowledgement of older work. This is a simpler alternative to
versioned acknowledgements, suitable while there is one matching owner. Any error
discards its working set; acquiring a new lease also requires a full snapshot.
Full reconciliation runs at least every configured reconciliation interval, which
configuration validation already bounds to ten seconds. This repairs lost hints,
crashes, response loss and changes during a non-atomic population observation.

Only that owner holds private per-cell signal slices. It drops expired entries
and unused slice references, re-evaluates assignment deadlines, and schedules
availability/pending/assignment deadlines independently of incoming requests.
HTTP handlers never inspect the participant cache. On lease loss, error or worker
shutdown it discards the cache. Expiring cache references is not forensic erasure;
live process memory remains visible to a privileged operator.

Full snapshots seek each next page through the previous member's indexed rank in
one atomic script. This avoids growing score-range OFFSET traversal. If deletion
removes the cursor, retry the snapshot at most three times; never silently skip
stable records because a cursor disappeared. Reads use bounded MGET batches,
reuse decoder scratch storage after clearing it, and cache parsed cell IDs within
the observation. Extra service ACL commands remain confined to `gati:*`.

Reachability membership uses binary search over the already sorted immutable
intersection index. No new map provider, participant identifier, exact coordinate
or persistent database is introduced. Public aggregate rules and the atomic
reservation/admission/arrival checks remain authoritative.

## Consequences and limits

- Between full reconciliations, database work follows changed cells. A dense
  cell may still contain a large share of the population; this is not constant
  work per signal. Benchmarks must report both uniform and hotspot cases.
- Candidate planning still examines the in-memory population and periodic full
  reconciliation/publication remain linear. This is not a claim that citywide
  matching is O(1), nor a replacement for future regional worker ownership.
- The cache trades some live memory for fewer database reads/decodes. Measure
  CPU, allocation traffic, process RSS and actual worker delays separately.
- A failed or stale snapshot can delay formation; atomic transitions reject stale
  founders. The cache is not an authoritative membership database.
- Parallel matching across owners remains deferred until measured need and a
  fencing/boundary design. API replicas must not each hold this working set.

## Verification

Real-store tests cover equal deadlines, cancellation during traversal, sparse or
missing records, bounds, mixed decoder state, change-batch handoff, ownership and
incremental cancellation/reconciliation. Worker tests cover expiry/reference
removal and time-based assignment changes. The 3,000-person Tirana response run
and exact before/after benchmark/load evidence are recorded in PROGRESS.md.

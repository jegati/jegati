# 0009: Optional, temporary Web Push

Date: 2026-09-13. Implements the selected optional-notification usability scope.
No participant identity is added. Notifications/follows for users without active
willingness remain separate work. Push is off by default and never a prerequisite
for willingness, matching, admission or arrival.

Use a bounded reconciler over opted-in live sessions, with one encrypted subscription
record doubling as the coalescing outbox. Atomically derive its latest relevant
invitation/presence state, claim delivery and acknowledge with a random claim fence.
This refines the original plan's transition-written outbox: no extra writes are
added to every matching/arrival transaction, and changes between reconciliation
checks may coalesce. This is appropriate for a generic prompt to open current state;
it is not an audit trail or guaranteed delivery of every transition. It does not
change participant state. Decline, cancelled/expired sessions, closed gatherings and
expired presence are rechecked before claiming. In-flight delivery cannot be recalled.

Endpoint/key material is authenticated-encrypted outside Valkey with a key
derived from the service's mounted VAPID private key, bound to the session hash and
random subscription binding. Valkey stores ciphertext, bounded scheduling metadata
and the last relevant state only. No endpoint/credential logs or backup. One record
per signal; its absolute expiry and native TTL cannot exceed willingness or an
earlier browser subscription expiry. Cancellation removes the record and due entry
atomically. Opt-out removes delivery but keeps the minimal rate-gap deadline until
session expiry so repeated opt-in cannot bypass limits. Bounded index reads and
hard key TTLs handle records disappearing before index cleanup. Retry attempts and
queue lifetime are capped. The interval is at least both the configured minimum and
3600 / max-per-hour; retry attempts within one queued notification are separately capped.

Production delivery requires exact configured HTTPS endpoint hosts, verified
public DNS addresses at connection time, TLS hostname verification, no proxy or
redirect forwarding, bounded concurrency, request time and response size. An
upstream Web Push implementation supplies standard encryption and VAPID. Test
transport injection remains a Go test concern, not a production HTTP control route.

Browser opt-in explicitly discloses provider involvement and temporary browser
storage. Only opted-in users may retain an unexpired capability in IndexedDB for
closed/reopened-page delivery; ordinary participation remains sessionStorage-only.
The worker must not cache private requests or responses. Notifications use generic
Albanian text and a fixed first-party opening URL. A binding/deadline and current
server state gate display; clicking never enrolls, admits or confirms arrival.
Opt-out/cancel remove local resume material; closed browsers may physically retain
expired bytes until their next execution. Server deadlines make them unusable. No
forensic browser erasure or provider deletion guarantee is claimed. Browser/OS push
providers can observe delivery metadata and may retain endpoints beyond our TTL.

Actual code and test status belong in PROGRESS. Real provider/device interoperability
is separate from local fake-transport/browser tests; do not claim it from mocks.

Opt-out increments a revision on the existing temporary signal; registration must
echo the current private revision. This rejects delayed pre-opt-out writes without
adding a permanent identifier. Signal cleanup removes push index members before
freeing admission capacity. A changed current state waits for the rate gap before
starting its queue lifetime, so a stable JEMI KËTU update is not lost simply because
the configured gap exceeds queue TTL. Intermediate changes may still coalesce.

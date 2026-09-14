# Initial threat model and evidence index

Update 2026-09-14: matching now retains a private per-cell working set only in the active owner, discards it on error/ownership changes, and rechecks deadlines before use. Ten-second cell change markers contain no participant IDs. [Decision 0011](decisions/0011-incremental-matching-working-set.md) documents the additional live-memory exposure and retention/reconciliation controls; no public count or arrival guarantee changes.

Status: willingness, private activation/arrival, limited network/write admission
and service ACLs are implemented; public aggregate publication remains planned. See STORAGE.md and API.md. See the
[scaffold report](reports/02-scaffold.md) for config validation, API headers and
simulation-profile rejection tests. Keep this document updated as milestones land. The [plan's data inventory](DEVELOPMENT_PLAN.md#6-data-inventory-and-expiry-contract)
defines proposed storage/access/lifetimes; avoid maintaining a conflicting copy.

## Assets and trust boundaries

The proposed deployment boundary and its unimplemented controls are detailed in
[DEPLOYMENT_PREPARATION.md](DEPLOYMENT_PREPARATION.md). In particular, the current
API uses the direct peer IP for network limits: a trusted forwarding boundary must
be implemented and tested before using a shared production proxy. An HTTPS edge
provider can inspect temporary authorization capabilities and coarse-cell request
contents; provider-side retention is outside application TTL guarantees. Tunnel
origin isolation does not remove that provider visibility. These are planning
findings, not evidence that a Cloudflare deployment is configured or audited.

Protect coarse participation, session capabilities, transient admission/arrival
links, notification subscriptions and service integrity. Exact participant
coordinates belong only on the device. Public map features and reviewed released
aggregates are intentionally public.

Boundaries: browser/OS → network/proxy → API → private ephemeral store/worker →
public cache or optional push provider. The build pipeline and host administration
are separate trust boundaries. An app without registration is not automatically
anonymous to network providers or its operator.

## Threats and planned evidence

| Threat / actor | Planned mitigation | Evidence required; current gap |
| --- | --- | --- |
| Public scraper infers a participant | Fixed delayed releases, suppression/buckets, no individual/queryable counts | Map/history/cache composition and withdrawal/differencing tests; inference unresolved |
| Enrolled attacker probes private decisions | Bounded admission/invites, stable threshold, no counts, coarse selection | Colluding enrollment, late-join and intersection-choice probe scenarios; no formal guarantee |
| Sybil attacker fabricates signals/arrivals | Rate/admission budgets, temporary credentials, bounded challenges if useful | Measure false activations and shared-NAT fairness; unique humans unproven |
| Client spoofs presence/replays requests | One-use nonce, eligibility, atomic one-claim rule, freshness | Replay/race/expiry tests; genuine physical presence unproven |
| API client exhausts resources | Schema/body/deadline limits, bounded work/queues, rate limits | Sustained/burst/hotspot tests including accepted and rejected requests |
| Network attacker overwhelms link | Hosting/edge protection plus cached public artifacts | Provider/deployment choice and operational exercise still required |
| Store reader obtains live participation | Coarse minimal records, restricted roles, short retention, no persistence | Store ACL/expiry/index tests and restore inspection; live records remain exposed to privileged access |
| Malicious operator observes/logs individuals | No individual admin UI, auditable configuration, restricted service access | Independent host inspection; privileged covert logging not prevented |
| Provider correlates requests | First-party assets, no app access logs, optional generic push | Browser network capture and provider inventory; network/provider visibility remains |
| Browser compromise steals capabilities | No third-party scripts, CSP, bounded client storage, header-only secrets | Browser security/network/storage tests; compromised device outside guarantee |
| Build/deployment differs from source | Pinned build inputs, signatures/hashes, independent rebuild and host inspection | Artifact comparison and scoped audit; remote version claim is not proof |
| Simulation reaches production | Separate origin/target allowlist, production excludes control routes | Config build guard and scaffold route tests passed (internal/config, internal/httpapi, scripts/smoke.sh); population driver also refuses production profiles/nonloopback/redirect targets before writes |

## Evidence rules

Own-session and private-preview responses use explicit HTTP-layer field allowlists
(`internal/httpapi/responses.go`), not serialized storage records with fields
removed afterward. Response-contract tests reject unexpected fields across the
simulated lifecycle. This prevents accidental expansion through storage schema
changes; it does not reduce the already documented private-invitation inference.

- Test with synthetic inputs only. Do not record real IPs, capabilities, push
  endpoints, request bodies or participant records in reports.
- Separate measured behavior, configuration inspection and architectural limits.
  Do not treat an intended control, green unit test or README claim as deployment
  evidence.
- For each milestone link the relevant code/test/report here or in PROGRESS,
  including failures and skipped checks. If an observable surface is unsafe,
  withhold it and resolve the issue; do not compensate with a policy statement.
- TTL must cover record links, indexes, queues, caches and subscriptions. Logical
  expiry is distinct from removal of bytes or copies made outside the service.
- Runtime endpoint/config hashes are useful comparison aids; independent remote
  backend honesty remains unproven even when builds reproduce.

### Implemented browser boundary (milestone 05)

The client requests location only after an explicit one-shot action, quantizes it
using the published grid, and sends only a cell ID. Manual map choice was removed
in schema 5; the current client requires device location.
No runtime map/font provider, analytics, cookies, localStorage or service worker
is used. Same-origin sessionStorage contains one capability, coarse request,
confirmation flag and deadline; retries reuse it without renewal, cancellation
success/expiry removes it, and reopening rejects expired state before sending it.
An unavailable sessionStorage falls back to tab memory; closing that tab can lose
control while the server signal expires normally. Device backups, restored tabs,
compromised browser code and local clock tampering remain outside deletion claims.
The API independently enforces its own time. Browser tests check exact-coordinate
absence in request URLs/bodies and storage plus failure/reload behavior; this is
not a complete malicious-browser or administrator audit.

### Implemented simulation boundary (milestone 06)

Simulation controls/flags are selected at compile time, require loopback and a
separate run credential, reject browser Origins, and use a disjoint ACL namespace.
The driver refuses production profiles/nonloopback/redirects before writes. Test
clock/source records are never served by the product routes. Synthetic reports
contain no capabilities and live outside production build inputs. Scenarios expose
that one-network limits still admitted more credentials than synthetic humans.
They establish neither independent arrivals nor resistance to distributed Sybils.

### Implemented private gathering flow (milestone 07)

Reservations lock the same bounded founding cohort through stability; atomic
activation rechecks immutable geographic claims, availability and configuration.
Public serialization removes reservation IDs, founder hashes and decline links.
Private invitations expose a collective intersection/time event, not counts. New
recipients can atomically join without an orphan willingness on rejection. Going
and decline revalidate expiry, reachability and state; retries do not renew time.
Cached destination offers are revalidated against the live store before admission.

This does not solve colluding threshold probes or nearest-intersection inference. Full
cross-surface review is still a release gate. The worker's snapshot/selection cost,
lease recovery, fairness, same-intersection conflicts and temporary-index cleanup need
stress/failure testing. Config changes invalidate pending reservations; handling a
removed intersection in an already-open gathering remains hardening work.

### Implemented arrival boundary (milestone 08)

A fresh one-shot browser position is coarsened locally; the request sends only the
coarse cell and authorization headers. The server validates the allowed intersection
cell/rings, current going state, gathering/session deadlines and a one-use nonce.
Raw nonces remain in browser memory only; the store keeps hashes with bounded TTL.
Replays cannot duplicate/refresh confirmation or restore a retracted claim. Stable
cohorts use nonce-specific member references so replacing a retracted arrival does
not conceal a broken stability interval. All loss paths and reads enforce deadlines.
Tests include founding/late equality, wrong cells, replay/races, retraction and
expiry, including joining an already confirmed gathering without automatic arrival.

Browser location can be spoofed and many credentials can represent one adversary.
These controls limit replay and resource consumption, not physical-presence proof
or human independence. No public individual arrival or count endpoint exists.

### Demonstrated inference failure — release gate, milestone 09

The production-threshold probe in reports/09-inference-gate.md reconstructs a
synthetic participant's coarse cell from the ordinary invitation when the observer
controls the other 19 founding signals. It reduces 127 compatible cells to one
while allowing every configured target radius. This is a concrete unmet privacy
requirement, not merely a hypothetical residual risk or an anonymity-proof caveat.
No public aggregate endpoint is involved. Rate limits/TTL/delayed public releases
do not eliminate this counterexample. The user retained nearest-crossroad selection under this bounded claim (decision
0004); local work may continue, with no stronger privacy claim.

### Device-only location boundary (schema 5)

The shipped UI requires device geolocation for willingness and arrival; it offers
no manual participant area, pin or address input. One-shot high-accuracy requests
and public-config age/accuracy checks reject poor fixes locally. Exact coordinates
and reported accuracy are discarded after local coarsening, not uploaded. Required
permission may exclude devices without a usable fix. Browser/OS location services
may use external providers; GATI cannot control their metadata handling.

A modified browser or direct API caller can still fabricate the coarse cell and
reported quality. The server has no authenticated evidence of GPS, a person's
location or unique humanity. Removing manual entry does not repair the collusion
counterexample. W3C describes device position and accuracy estimates, not physical
presence attestation: https://www.w3.org/TR/geolocation/.

### Population simulation evidence and finer cells (schema 6)

The response harness accepts only an isolated simulation profile and declared
loopback peers, drives the real admission/arrival checks and preserves application
rate budgets. Reports distinguish synthetic people, accepted credentials, unique
actors and repeated response/claim events. Injected gathering discovery does not
prove unfinished public-map or notification behavior. Functional clock advances
cannot bypass real abuse windows. A crowded-network control makes their exclusion
cost visible; distributed Sybil resistance remains unresolved.

Fractional radii and optional 100 m cells are supported. Smaller cells reveal more
precise participant areas, and thresholding still provides no formal anonymity.
The former 1,000 m baseline excluded 0.1/0.5 km radii under conservative whole-
cell reachability. The user selected 100 m cells / 50 m reported device accuracy
as the default after that comparison (decision 0006). This increases spatial
exposure and index cost; historical inference evidence retains its original config.

The dense experiment exposed a real ten-minute rate-secret rollover race. Secret
creation/read is now atomic using Valkey time. Real-store tests cover epoch-boundary
expiry and concurrent API replicas, without storing raw addresses or adding logs.
Secrets retain the existing bounded ten-minute lifecycle. These local controls do
not establish protection from an upstream DoS or privileged operator observation.

### Accepted public inference and cell-only gathering statistics (decision 0007)

The user authorizes statistics and map implementation despite inferential leakage.
Buckets, suppression and delay do not prevent colluding threshold probes, temporal
comparisons, nested-count or origin/destination correlation. These remain documented
limitations; they are not blocking gates for this authorized local feature work.
Public entries must omit intersection IDs, names and coordinates and share the
willingness public grid. An eligible observer can still learn destinations through
private joining. Public snapshots can be archived indefinitely by others.
Push is explicitly opt-in and off by default; provider correlation and possible
longer-lived endpoint metadata remain. Opted-in IndexedDB resume retains a bounded
credential whose server expiry invalidates access; closed clients cannot guarantee
physical deletion on schedule. No opt-in means no added background resume storage.
Publication correctness, suppression, coordinate omission, expiry, replay protection,
SSRF and queue bounds remain required implementation tests.

### Implemented activity surface evidence

The pure builder rejects duplicate contributions, bounds capture/output, and emits
only cell IDs plus published buckets/event metadata. The real publisher uses a
separate expiring lease and immutable epoch keys with delayed visibility and logical/
native expiry; public reads never scan participant records. HTTP tests cover cache
expiry, auth no-store, query rejection and concurrent identical reads. Browser
fixtures cover cell layers, neutral suppressed data, identical own/gathering counts,
expiry, explicit joining and capability-preserving retry. Real fixed-clock API tests
cover private JEMI KËTU with a still-suppressed public arrival value. Neither these
checks nor the accepted nested/threshold-inference fixtures establish anonymity.
Capture is a bounded interval, so cancellation after observation may remain in that
release; no per-participant capture history is persisted. Push-related controls in
the plan remain unimplemented, and no subscriptions/provider calls are active.

### Explicit destination preview (decision 0008)

`POST /api/gathering-preview` can disclose a published gathering's exact mapped
crossroad to an eligible coarse location claim before attendance commitment. This
expands the accepted destination-inference surface: willingness is not required for
this query and a scripted client can invent claims. The header capability is random;
new preview capabilities stay in page memory and no participant/preview record is
created. Existing signals must use their immutable claim. Strict body/origin/query
checks, shared admission/write rate limits, current released-ID gating, live
reachability/cutoff checks and no-store bound the endpoint. Response lifetime uses
`geography.location_fix_max_age_seconds` and earlier public/live deadlines. Final
admission revalidates eligibility; a preview neither reserves a place nor guarantees
later admission. Public maps/snapshots remain cell-only. Fixed-clock store/API and
browser tests cover no enrollment, unchanged public release and preview expiry.

### Optional push transport (decision 0009, schema 8)

Opt-in Web Push requires a provider endpoint and encryption keys. Transport material
is AES-GCM encrypted under a domain-separated key derived from the mounted VAPID
service key, with signal hash + random binding as associated data. Store compromise
alone cannot decrypt or transplant endpoint material; compromise of the API host/key
still can. The store retains only one current-state/outbox record per live signal
and a minimal rate-gap deadline. It adds no personal history. Sender claims and
acknowledgements are fenced; retries/queues/index work are bounded. Cancellation and
read-time deadlines prevent later claims, but cannot recall a request already in flight.

Provider URLs require exact configured HTTPS hosts. DNS is checked at connection
time; all answers must be public addresses and the checked IP is dialed directly.
TLS verifies the original hostname; proxies and redirects are disabled. Tests cover
private/metadata/transition addresses, mixed DNS answers and non-forwarded redirect
authorization. Work/concurrency/time/response sizes are bounded. Payloads are padded,
encrypted metadata (binding/deadline), without location, participant/gathering ID or
message text. Browser display fetches current authorized state and uses generic
Albanian text. Providers can still associate endpoints, sender, recipient and timing;
application TTL is not provider deletion. Bad/stale provider responses stop or bound
retries. Delivery is best-effort, can coalesce states, and is not exactly-once.

The reviewed pinned transport and actual restricted-store fake-provider tests check
endpoint encryption, immutable registration, cancellation before send, padded/encrypted
requests, expiry and claim races. They do not establish Chrome/Firefox/Safari provider
interoperability or OS notification behavior; that needs real-device evidence.


Browser opt-in persists only a bounded capability/binding/revision/deadline/status,
without location/history. Reopening fetches existing state without re-enrollment;
uncertain recovery offers explicit retry/cancel. Opt-out removes local display/resume
before provider cleanup. Server opt-out revisions fence delayed registration writes;
claim fences prevent old delivery acknowledgements from overwriting newer work.
Expiry cleanup removes due-index entries before freeing admission capacity. Tests
exercise both cancellation during registration and retained-index churn.

The actual first-party worker has no fetch interception or response caches. Browser
tests inject provider delivery via CDP with the app closed and check native generic
notification display, stale/forged binding rejection and cancellation. Permission
revocation is checked on resume and worker execution. Local opt-out cannot recall an
already-displayed OS notification or provider request. Suspended browsers can retain
expired bytes until running again, and browser-managed push subscriptions may outlive
app records; no forensic erasure or guaranteed background cleanup is claimed.

## Local automated hardening and monitoring (2026-09-13)

The optional operational socket is owner-only, with no public TCP listener or
participant labels. Completed request windows are suppressed/bucketed, collector
fields are allowlisted, history is bounded, and output is private from creation.
Worker errors are fixed states, not raw exception text. Correlation from process
resource use and task timing remains possible for a privileged observer; these
metrics are not a formal anonymity mechanism. See decision 0010 and MONITORING.md.

Additional evidence now covers seeded nonce/arrival/cancellation sequences,
configuration rollover, competing push workers against a 100-subscription fake
provider outage, native parser/capability/geography fuzzing, and private-monitor
schema rejection. Owned multi-replica tests inject process/store failures and check
cancellation/replay and loss of nonpersistent sessions. Load reports contain exact
synthetic traffic statistics only; those fields are not new production endpoints.
The actual normal-delay publisher browser journey passed through public preview
and late arrival. Browser emulation adapters are test-only and leave production
location checks unchanged. See AUTOMATED_TESTING.md and PROGRESS.md for results,
failed harness assumptions corrected, and untested release gates. These checks do
not eliminate location spoofing, Sybil, inference or malicious-operator risks.

## Production ingress and release boundary — 2026-09-14

[Decision 0012](decisions/0012-production-boundaries.md) adds exact-peer proxy trust,
canonical client IP validation, forwarding-header removal, credential-cache bypass,
private container networks and built browser CSP. The local production rehearsal
tests actual runtime persistence/log/swap/core settings, forged requests, real
replica switching and private monitoring. Caddy file capabilities are removed so
ALL can remain dropped. Host swap is enabled on this developer PC: container
validation does not certify host privacy. Public activation checks host prerequisites
and still requires provider/edge inspection. No participant backups or public
monitoring are introduced.

Cloudflare/tunnel is a trusted ingress and plaintext TLS intermediary; compromise
can reveal request IPs/coarse cells/capabilities or defeat network budgets. It cannot
be claimed blind or governed by application TTL. Host administrators can inspect
live Valkey/API memory. Immutable releases, independently compared browser assets
and privileged image/config checks improve verifiability, not remote attestation.
See [the runbook and residual gates](DEPLOYMENT.md).

## Final review findings — 2026-09-14

Application dependency audits and behavioral tests do not cover the whole runtime.
The [image inventory and security review](reports/security-2026-09-14.md) found
outdated Caddy Go/OS dependencies and additional tunnel package advisories. Public
launch is on hold pending remediation/triage and fresh artifact verification.
The Go application and npm dependency scans found no affected application code;
the module-only OpenPGP advisory concerns a package GATI does not import.

Release verification now rejects launch-command/environment/user, mount, resource,
network-attachment and reserved-proxy-address drift. It also refuses an unintended
live tunnel/worker omitted by deployment flags. Tests include an actual additional
network attachment rejected before accepting the restored topology. These controls
address accidental configuration drift, not malicious administrators or a dishonest
Docker API. Account-managed tunnel rules and host network definitions still need
separate review. CI now runs both Go and npm dependency audits; full production-image
scanning remains a separate release gate.


## Runtime remediation evidence — 2026-09-14

[Runtime remediation](reports/runtime-remediation-2026-09-14.md) supersedes the
unresolved-runtime status above for source 06fa798. Caddy and cloudflared are now
locked source builds using patched Go dependencies in scratch images. The tunnel's
four upstream Sentry initialization sites are disabled, with tests establishing no
client or captured event even when a DSN is configured. A build-time source check
rejects changed initializer counts/locations. This patch closes an outbound crash
reporting path that discarding container logs alone did not close.

All three exact Go executables retain symbols and pass binary vulnerability scans.
The remaining OpenPGP and CEL package findings are unused-code exceptions, tied to
specific versions/services and expiring 2026-10-14. Their existence is visible in
raw scans and the audit summary. Failed binary scans are never waived. Public
activation/rollback requires a passing audit for the exact release, no older than
24 hours. This guard checks local evidence; it cannot prevent a dishonest operator
from forging files or modifying the host.

Browser, two-API failover, arrival replay, network budgets, private monitoring and
release verification are rehearsed locally. Two clean builds match all executable
and browser bytes in the new images. Cloudflare/account rules, provider visibility,
real tunnel connectivity, real device/push behavior and independent review remain
separate gates. No new participant data, logging, tracking or public endpoint was
introduced. See decision 0013 and RUNTIME_SECURITY.md for source and audit details.

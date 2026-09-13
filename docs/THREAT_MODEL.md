# Initial threat model and evidence index

Status: initial threat design plus milestone 02 scaffold evidence. Willingness lifecycle, limited network/write admission and service ACLs are implemented;
activation/arrival/publication controls remain planned. See STORAGE.md and API.md. See the
[scaffold report](reports/02-scaffold.md) for config validation, API headers and
simulation-profile rejection tests. Keep this document updated as milestones land. The [plan's data inventory](DEVELOPMENT_PLAN.md#6-data-inventory-and-expiry-contract)
defines proposed storage/access/lifetimes; avoid maintaining a conflicting copy.

## Assets and trust boundaries

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
| Enrolled attacker probes private decisions | Bounded admission/invites, stable threshold, no counts, coarse selection | Colluding enrollment, late-join and crossing-choice probe scenarios; no formal guarantee |
| Sybil attacker fabricates signals/arrivals | Rate/admission budgets, temporary credentials, bounded challenges if useful | Measure false activations and shared-NAT fairness; unique humans unproven |
| Client spoofs presence/replays requests | One-use nonce, eligibility, atomic one-claim rule, freshness | Replay/race/expiry tests; genuine physical presence unproven |
| API client exhausts resources | Schema/body/deadline limits, bounded work/queues, rate limits | Sustained/burst/hotspot tests including accepted and rejected requests |
| Network attacker overwhelms link | Hosting/edge protection plus cached public artifacts | Provider/deployment choice and operational exercise still required |
| Store reader obtains live participation | Coarse minimal records, restricted roles, short retention, no persistence | Store ACL/expiry/index tests and restore inspection; live records remain exposed to privileged access |
| Malicious operator observes/logs individuals | No individual admin UI, auditable configuration, restricted service access | Independent host inspection; privileged covert logging not prevented |
| Provider correlates requests | First-party assets, no app access logs, optional generic push | Browser network capture and provider inventory; network/provider visibility remains |
| Browser compromise steals capabilities | No third-party scripts, CSP, bounded client storage, header-only secrets | Browser security/network/storage tests; compromised device outside guarantee |
| Build/deployment differs from source | Pinned build inputs, signatures/hashes, independent rebuild and host inspection | Artifact comparison and scoped audit; remote version claim is not proof |
| Simulation reaches production | Separate origin/target allowlist, production excludes control routes | Config build guard and scaffold route tests passed (internal/config, internal/httpapi, scripts/smoke.sh); population-tool destination checks remain future work |

## Evidence rules

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
using the published grid, and sends only a cell ID. Manual map choice is supported.
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
Private invitations expose a collective crossing/time event, not counts. New
recipients can atomically join without an orphan willingness on rejection. Going
and decline revalidate expiry, reachability and state; retries do not renew time.
Cached destination offers are revalidated against the live store before admission.

This does not solve colluding threshold probes or nearest-crossing inference. Full
cross-surface review is still a release gate. The worker's snapshot/selection cost,
lease recovery, fairness, same-crossing conflicts and temporary-index cleanup need
stress/failure testing. Config changes invalidate pending reservations; handling a
removed crossing in an already-open gathering remains hardening work.

### Implemented arrival boundary (milestone 08)

A fresh one-shot browser position is coarsened locally; the request sends only the
coarse cell and authorization headers. The server validates the allowed crossing
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

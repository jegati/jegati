# Current HTTP API

User-facing errors are generic Albanian JSON. No cookies or participant identifiers
are issued in URLs. Authenticated responses are `Cache-Control: no-store`.

| Route | Behavior |
| --- | --- |
| `GET /healthz` | API process health (does not prove every dependency or privacy control) |
| `GET /api/config` | Nonsecret functional config, schema version 6, canonical SHA-256 |
| `GET /api/geography` | Public fixed grid bounds/steps for client-side coarsening |
| `GET /api/map/roads` | Cacheable first-party public OSM road geometry, with ETag |
| `POST /api/signals` | Create or identically retry this capability's willingness |
| `GET /api/signal` | Read only the caller's current signal |
| `DELETE /api/signal` | Cancel caller's signal; idempotent |
| `POST /api/going` | Admit existing willingness to a reachable live gathering |
| `POST /api/decline` | Decline the current gathering while retaining willingness |
| `POST /api/join` | Atomically create/reuse willingness and admit a recipient |
| `POST /api/arrival-nonce` | Issue an expiring arrival challenge for a going session |
| `POST /api/arrival` | Confirm a fresh coarse arrival; idempotent nonce replay |
| `DELETE /api/arrival` | Retract arrival, retaining going intent and willingness |

Participant routes require `Authorization: Bearer <base64url-encoded 32 random bytes>`.
The client generates the capability; only its SHA-256 hash is stored. Create accepts
exactly `cell`, `radius_km`, `availability_minutes` in an application/json object.
Unknown/duplicate/null/coerced fields, coordinates, unsupported grid/radii/durations,
extra JSON documents and oversized bodies are rejected. No raw token/hash is echoed.

Successful create returns status 200 with coarse cell, radius, duration, state and
server-authoritative `created_at`/`expires_at` Unix milliseconds. Identical retries do
not renew it; changed content is 409. Missing/expired/cancelled status is 410;
cancellation returns 204. Invalid credentials: 401. Cross-origin browser requests:
403. Limits: 429 with Retry-After. Missing/unavailable store or admission capacity:
503. There is no GET collection endpoint, anonymous count query or admin route.

Forwarded network headers are not trusted. Vite proxies use the socket peer's shared
limit; configured proxy trust is a later deployment concern. The JSON media type,
explicit bearer capability and same-host Origin check prevent ambient-cookie CSRF.
This is not strong human/Sybil verification.

Own-session responses can include an `invitation` with a random gathering ID,
public intersection, activation/end times, state and configuration hash. They never
include counts, founder/member lists, pending reservations or a group center.

Going/decline accept exactly `gathering_id`; join accepts the three willingness
fields plus `gathering_id`. Reachability, cutoffs and state are revalidated. A stale
join returns 410 without creating unintended willingness; successful retries never
renew the gathering/session. Going to another valid gathering switches intent.
Declines suppress that gathering for this session, with configurable cooldown and
bounded temporary decline links. They do not cancel willingness. Notification/follow endpoints are not implemented yet; invitation delivery currently
uses the foreground own-session poll. Direct recipient join is implemented at the
API; public map/notification cards will connect the client to it in their milestones.

Arrival issuance and confirmation also require `X-Gati-Arrival-Nonce`, an
authorization header containing a client-generated random 32-byte base64url nonce.
Issuance has an empty body and returns only `expires_at`; repeated unused issuance
preserves the first deadline. Confirmation accepts exactly `cell`, checked against
the current intersection's allowed coarse cell/rings. One credential contributes once.
The own-session view includes `state: here` and `arrival_until` while fresh. A
gathering becomes `jemi_ketu` after configured stable presence; it resets if current
accepted arrivals fall below threshold. Internal nonce/member/cohort data is never
returned. Wrong-area claims are 409; unavailable/expired/used-retracted claims 410.

## Public activity (schema 7)

`GET /api/activity/latest` returns a canonical delayed Tirana release, or 204 when
none is available. Query parameters are rejected. No capability or location is
required. Response fields: version, id, config_sha256, map_version, observed_from,
observed_until, release_at, expires_at, grid, areas and gatherings. Times are epoch
milliseconds. Areas have cell/willing lower-bound buckets. Gatherings have opaque
id, public cell, ends_at, sampled state and optional going/here lower-bound buckets.
No intersection identifier/label/coordinates or individual record is present.
Missing buckets are unavailable/suppressed, never zero. Going includes arrivals.
The fixed public grid is common to both layers; private admission remains unchanged.

ETag supports 304. Shared max-age is at most 30 seconds and never beyond expiry;
authenticated requests are always no-store. Public reads inspect only aggregate
keys. Origin retention includes pending release and cannot revoke external copies.
Known inference limitations are accepted in decision 0007.

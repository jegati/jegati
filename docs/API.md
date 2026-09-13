# Current HTTP API

User-facing errors are generic Albanian JSON. No cookies or participant identifiers
are issued in URLs. Authenticated responses are `Cache-Control: no-store`.

| Route | Behavior |
| --- | --- |
| `GET /healthz` | API process health (does not prove every dependency or privacy control) |
| `GET /api/config` | Nonsecret functional config, schema version 2, canonical SHA-256 |
| `GET /api/geography` | Public fixed grid bounds/steps for client-side coarsening |
| `GET /api/map/roads` | Cacheable first-party public OSM road geometry, with ETag |
| `POST /api/signals` | Create or identically retry this capability's willingness |
| `GET /api/signal` | Read only the caller's current signal |
| `DELETE /api/signal` | Cancel caller's signal; idempotent |

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

Admission, activation, arrival and notification routes are not implemented yet.

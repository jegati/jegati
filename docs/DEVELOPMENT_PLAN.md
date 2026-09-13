# GATI 🦩 — development and verification plan

Status: architecture and milestone plan. Implementation has started; consult PROGRESS.md for completed functionality and validation. This document describes intended behavior, not evidence that every protection already exists.

Read [AGENTS.md](../AGENTS.md) for development instructions and
[PROGRESS.md](PROGRESS.md) for actual implementation and validation status. The
[requirements summary](REQUIREMENTS.md), [initial decisions](decisions/0001-mvp-boundaries.md)
and [threat model](THREAT_MODEL.md) preserve the working context for future sessions.

## 1. Product contract

**Bëje vullnetin tënd të dukshëm.**

Build a small, anonymous, Albanian web app that works on this computer and can later be deployed from the same repository:

**A JE GATI? → JAM GATI → JEMI GATI → PO, PO SHKOJ → JAM KËTU → JEMI KËTU**

The first screen offers availability duration, travel radius, a coarse area selection or optional one-time location permission, and **JAM GATI**. No account, onboarding sequence, profile, feed, chat, organizer role, user-generated event descriptions, attendance pressure, or participation history. Display “Gatishmëria nuk është premtim për të marrë pjesë.” Cancellation is immediate and neutral. Use 🦩 as the simple brand symbol; all user-facing copy, accessibility labels, errors, notifications, and installation instructions are Albanian.

The initial activity is gathering near an automatically selected mapped pedestrian crosswalk. There is no manually predefined meeting-area catalog. Continuous matching finds a compatible group and selects the nearest eligible crosswalk to its coarse group center, subject to every counted participant's travel radius. The destination refers to pedestrian space beside the crossing, not standing in the roadway. A notification names the crossing and gathering deadline. Newly notified people can join an active gathering and confirm arrival without having participated in its original activation. Availability starts at 30 minutes; gathering formation is not tied to scheduled time slots.

MVP includes willingness, activation, going/declining, arrival, aggregate map, foreground notifications, optional short-lived background push, expiring area follows, simple protected statistics, local simulation, deployment and audit tooling. Future availability is deferred. Native mobile apps, custom activities, route planning, persistent follows, federation, and advanced cryptographic presence proofs are deferred.

## 2. What we can honestly promise

1. No identity registration or permanent application participant identifier. A random, expiring capability is necessary to update one willingness signal, prevent replay, and associate an arrival with that signal. It links actions within that temporary session; it must not be described as unlinkable during the session.
2. Exact coordinates are processed on the device, converted to a coarse cell, and never sent to GATI. The server necessarily sees coarse cells and temporary session state. Request timing, IP addresses at the network boundary, and optional push endpoints remain privacy risks.
3. One accepted arrival per eligible temporary credential is achievable. Proving that credentials represent independent humans, or that browser-reported positions are genuine, is not achievable with this MVP. **JEMI KËTU** means a threshold of accepted arrival claims, not independently authenticated people. If verified human independence is mandatory, that requirement remains unmet and blocks a claim of full compliance.
4. Fixed, delayed, bucketed releases reduce public inference. They do not mathematically guarantee anonymity against collusion, observation at a meeting, or an adversary controlling most inputs. Thresholding alone is not differential privacy.
5. A memory-only store, bounded lifetimes, no participant backups, restricted service credentials, and no participant admin UI reduce compromise exposure. They cannot prevent a privileged malicious host operator from reading live memory or modifying code. Encryption whose keys live on the same host does not solve that problem.
6. Public source, independent rebuilds, artifact hashes, and deployment inspections provide evidence. A remote server can lie about its version or secretly log requests. An ordinary web deployment cannot cryptographically prove the absence of hidden backend behavior. Hardware attestation or independently operated relays would introduce complexity and other trust assumptions; they are not MVP promises.

These limits must appear in the public privacy/threat documentation and be reflected accurately in concise Albanian product explanations. A public release is conditional on accepting these bounded guarantees; stronger absolute interpretations of the requirements need a different research scope.

## 3. Small architecture

| Component | Proposed choice | Reason |
| --- | --- | --- |
| Client | TypeScript, Vite, small PWA; initially no UI framework | One browser implementation for desktop and mobile, modest dependency surface |
| Map | MapLibre GL JS, first-party local basemap assets | Collective polygons and labels, no external tile requests revealing viewed areas |
| API and workers | Go, one repository and binary with separate API/worker modes | Stateless API replicas; bounded background aggregation; straightforward tests and builds |
| Temporary state | Valkey with RDB/AOF persistence explicitly disabled | TTLs, atomic state transitions, counters and expiring indexes without a participant disk database |
| Durable content | Reviewed files and static artifacts | Imported crossing dataset, schemas, public configuration and protected aggregate summaries need no SQL service initially |
| Deployment | Docker Compose and Caddy on one Linux host | One documented deployment path, local parity, TLS at the public boundary |
| CI | Repository workflow: tests, builds, dependency/license checks, artifacts | Auditable release evidence; no runtime dependence on CI |

Choose and pin supported versions at implementation time. Do not add PostgreSQL, Kubernetes, Kafka, a hosted analytics service, or separate business microservices initially. Valkey is not memory-only by default merely because it is an in-memory database: explicitly configure and verify persistence settings [1].

Data path:

```text
Browser: coordinates → coarse cell; expiring capability
   │ first-party HTTPS, no coordinates in requests
   ▼
Caddy / optional public edge protection
   ├── static client + basemap + canonical public snapshots
   └── Go API → private memory-only Valkey
                         ▲
                    Go worker
                         ├── activation and arrival transitions
                         ├── immutable, privacy-filtered snapshots
                         └── expiring notification delivery queue
```

The API accepts a documented coarse-cell format, never participant coordinates. Start with a fixed geographic grid and a spatial index of pedestrian crossings imported from a versioned OpenStreetMap extract for Tirana. Precompute static cell/radius-to-crossing reachability; choose destinations dynamically from that map data. Use a reviewed geographic library for distance calculations and test cell boundaries. Publish the grid, import recipe, source checksum/date, attribution, filtering rules and resulting crossing dataset. Crossing coordinates are public map features, not participant locations. Basemap requests come from the same origin and do not use an external geocoder, font CDN, or map SDK telemetry.

## 4. Matching and state transitions

### Willingness

- Defaults to evaluate in simulation: duration choices 30/60/90/120 minutes, with a minimum of 30 minutes, travel radius 1/3/5 km, approximately 1 km cells. These are tunable hypotheses, not privacy-certified numbers.
- The client generates a cryptographically random capability of at least 256 bits. Only its hash is stored; send the capability in an authorization header, never a URL. Client credentials must be cleared on cancellation/expiry while the client runs and rechecked immediately on reopening; a closed browser cannot execute a cleanup timer. No automatic session renewal or permanent browser ID. Explain that browser backups and a compromised device are outside server deletion guarantees.
- Store coarse cell, radius bucket, server timestamps, deadline, temporary state and optional current gathering reference. Derive everything else where possible. One active signal per capability; renewing creates a fresh session after the old session ends. Multiple devices or browser resets can still create multiple credentials.
- Validate supported geography, enum choices, schema sizes, expiry and replay atomically. A server-generated deadline is authoritative.

### Continuous matching and automatic crosswalk selection

- Match on every accepted creation, change, cancellation, admission and expiry. A dirty-cell queue coalesces changes with a configurable maximum debounce of 2 seconds; a reconciliation sweep every 10 seconds recovers missed work. Deadline timers recheck stability and eligibility even without a new request. These are processing bounds, not scheduled gathering slots. Public map releases remain separately delayed for privacy.
- At evaluation time `t`, count a signal only if its availability deadline is at least `t + minimum_remaining_minutes` (initially 15 minutes). All counted signals therefore share a usable upcoming interval. A 30-minute willingness submitted at any wall-clock minute can qualify immediately; no rounding to quarter-hours or waiting for a full aligned window.
- For each affected neighborhood, use the static reachability index to find crossings that are within every proposed member's chosen radius. Conservatively calculate the greatest geographic distance from the reported cell to a crossing. Distance is geographic distance, not walking time; coarsening can exclude some boundary matches.
- Generate candidate groups from eligible unassigned signals sharing at least one reachable crossing. Rank candidates by eligible count descending, oldest signal first, then stable map-feature ID. Resolve overlapping memberships atomically so a signal supports at most one new gathering. This bounded heuristic can miss alternatives; report missed groups in simulation. Dense indexes are processed in resumable bounded batches, not silently truncated to a small handpicked list of meeting places.
- Bound each founding transaction to exactly the configured activation threshold, selecting the oldest compatible signals. Other compatible people use the late invitation/admission flow. Reserve the same founding cohort throughout stability; loss of a founder invalidates the reservation rather than treating a replacement as uninterrupted eligibility. See decision 0003.
- For each proposed group, calculate the mean of its members' coarse-cell centers. Define **closest crosswalk** as the eligible mapped crossing with the shortest geographic distance to that coarse group center, restricted to crossings reachable by every counted member. Break ties by stable map-feature ID. Never use exact device locations for selection or publish the computed center. A selected crossing still reveals information about the coarse group location: include nearest-crossing selection and repeated candidate probes in the inference review. Revalidate membership, destination and deadlines before activation.
- Import pedestrian crossing nodes and crossing ways, normalize duplicates, and retain their public source IDs and geometry. OpenStreetMap documents `highway=crossing` and `highway=footway` + `footway=crossing` [8]. Exclude explicitly prohibited, private, disused or construction crossings and non-pedestrian features. Use mapped pedestrian-side geometry when available; otherwise describe the crossing as a landmark without inventing an exact sidewalk point. Document incomplete map coverage; map tags do not certify space or crowd capacity.
- If no eligible shared crossing is mapped, keep willingness active and display an Albanian explanation that no suitable mapped meeting point was found. Do not invent a destination or silently fall back to a predefined plaza. Crossings may be excluded through published map corrections, but normal formation must not depend on an operator approving each location.
- Map imports are offline administrative jobs, never per-participant external queries. Pin an import version during matching; updates apply to new gatherings. Once activated, the crossing and deadline stay fixed, including when late participants join. If a crossing is later marked unavailable, close the gathering with a neutral update rather than silently moving people.

### JEMI GATI and joining an ongoing gathering

- Suggested activation threshold: 20 eligible live signals, continuously eligible for 10 seconds. A timer triggers validation at the stability deadline; falling below threshold or losing a common destination resets the timer. On activation, store the triggering decision and configuration/map versions, not a permanent membership history. Under healthy load the processing target is within 5 seconds after stability completes; notification transport adds its own delay.
- Set `gathering.ends_at = min(activation_time + maximum_lifetime_minutes, earliest founding signal deadline)`, initially capped at 60 minutes. All founding signals have at least the configured 15 minutes remaining at activation. Late arrivals never extend this deadline. Show the destination, remaining time and **PO, PO SHKOJ** / **JO TANI**. Activation records a threshold reached, not a promise of continuing attendance.
- Give an existing compatible open gathering priority over forming a competing one. Offer it to newly GATI users and eligible nearby/area followers. A late participant needs sufficient remaining availability and a reachable destination; they need not have been present at activation. Gatherings already in **JEMI KËTU** remain joinable until their admission cutoff.
- A notification or public aggregate gathering card can open the join flow. Someone without an active willingness session chooses duration (minimum 30 minutes), radius and coarse area, then presses **PO, PO SHKOJ**; atomically create their expiring capability/session and admit it. Someone with an active session reuses it. Viewing a card or receiving a notification does not count as willingness, going or arrival.
- Revalidate admission atomically against live session state, destination reachability, gathering deadline and rate limits. Require at least `late_join_min_remaining_minutes` (initially 5) in both the session and gathering. Return a neutral expired/unavailable response when a stale notification no longer permits joining. Admit one capability to at most one gathering at a time; retries cannot duplicate it. A switch retracts the previous live contribution before admission to another gathering.
- Founding and later participants use the same going and arrival flow. Late joins do not rewrite the original activation count, create another activation event or automatically count as arrivals. Admission responses and status endpoints never reveal individual membership or exact counts.
- **PO, PO SHKOJ** changes intent only; **JO TANI** declines that gathering invitation while keeping the user's willingness active until its original deadline. Suppress repeat invitations to the declined gathering for that session; allow another compatible gathering or explicit cancellation. No response implies no attendance. Cancellation removes live contributions immediately without rewriting already released snapshots.
- Participant invitations can reveal a threshold event. Do not expose exact counts or distance to threshold. Apply bounded enrollment and invite cooldowns, and test probe accounts and late-join inference. Continuous matching does not authorize continuous publication of raw activity.

### JAM KËTU / JEMI KËTU

- A going user can request a one-use, short-lived arrival nonce for their currently admitted gathering, whether admission occurred before or after activation. Require a fresh one-shot location fix, convert it locally to the allowed coarse arrival cell, and send the cell and nonce. Never use background tracking.
- Check gathering deadline, prior going state, nonce lifetime/replay, admitted membership and allowed arrival cells. Reject insufficient location accuracy locally; do not upload accuracy or exact coordinates. Manual area selection works for willingness but cannot substitute for the arrival location check.
- Atomically accept at most one arrival per admitted capability. Issue **JEMI KËTU** after a proposed 10 fresh accepted arrival claims stable for a configurable 10 seconds, evaluated continuously with deadline timers. Public map publication still needs the stricter public threshold. Repeat arrivals, duplicate requests and worker retries must not inflate counts.
- A location-spoofing client or Sybil attacker can pass these checks. Nonces prevent replay, not spoofing. Name this limitation in the acceptance report.
- An arrival claim is fresh for 15 minutes or until the gathering ends, whichever comes first. Do not silently refresh it. Public labels describe a recent confirmation window, not continuous occupancy. Allow retraction; it affects future releases only.

### Central functional configuration

Maintain one annotated `config/gati.yaml` with a typed schema and a read-only, nonsecret effective-configuration endpoint consumed by the client. Operators change counts, radius choices and timing in this file rather than editing code. Keep credentials in a separate secret source. The configuration file and validation/inspection commands are implemented in milestone 02. Matching and notification behavior consuming these settings remains future work.

```yaml
profile: production
availability:
  minimum_minutes: 30
  choices_minutes: [30, 60, 90, 120]
  maximum_minutes: 120
geography:
  cell_size_meters: 1000
  travel_radius_choices_km: [1, 3, 5]
  crossing_dataset: tirana-crossings-v1
matching:
  activation_count: 20
  activation_stability_seconds: 10
  maximum_debounce_seconds: 2
  reconciliation_seconds: 10
  minimum_remaining_minutes: 15
  crossing_index_batch_size: 500
  candidate_batch_size: 100
  destination_rule: nearest_eligible_crosswalk_to_coarse_group_center
  prefer_open_gatherings: true
  maximum_gathering_minutes: 60
  late_join_min_remaining_minutes: 5
arrivals:
  confirmation_count: 10
  confirmation_stability_seconds: 10
  freshness_minutes: 15
  nonce_seconds: 120
  allowed_cell_neighbor_rings: 0
public_activity:
  minimum_count: 20
  count_buckets: [20, 50, 100, 250, 500, 1000]
  release_seconds: 300
  delay_epochs: 1
  snapshot_retention_minutes: 15
  daily_summary_retention_days: 30
notifications:
  nearby_gati_count: 50
  nearby_arrival_count: 50
  nearby_radius_km: 5
  foreground_poll_seconds: 30
  push_min_interval_seconds: 300
  push_max_per_hour: 6
  queue_ttl_seconds: 300
  area_follow_max_hours: 24
```

- `activation_count` controls **JEMI GATI**; `confirmation_count` controls **JEMI KËTU**. Counts refer to accepted temporary signals/arrival claims, not verified unique humans. The two `nearby_*_count` values separately control broad “enough activity near you” alerts for followers/nonparticipants. Public alert counts must be bucket boundaries at or above `public_activity.minimum_count`; never consult suppressed raw counts for those alerts. A private invitation uses the user's chosen travel radius; a broad nearby alert uses the configured notification radius around the followed/coarse area, then checks actual join eligibility when opened.
- These are starting values, not fixed constants. Add documented settings in the same schema for body/rate/admission limits, worker budgets, other retention bounds and supported geography as those features land. No hidden behavioral defaults in code. Derive allowed arrival cells from the fixed crossing cell and configured neighbor rings; disclose the resulting coarse presence area, and review expansions for spoofing/privacy implications.
- Validate at startup and in CI: reject unknown keys, nonpositive limits, unsorted buckets, choices below the 30-minute minimum, inconsistent deadlines, alert thresholds outside buckets, or a stability period incompatible with the available overlap. Production must not permit a minimum availability below 30 minutes. Reject a configuration that cannot accommodate the minimum availability choice after normal processing/stability delays.
- Version the supported privacy bounds in the schema: initially activation and arrival thresholds at least 10, public count at least 20, public release interval at least 300 seconds with at least one delayed epoch, and retention no longer than this plan's inventory. These are review guardrails, not proven anonymity guarantees. Operators can tune counts upward and other settings within those documented bounds without code changes. Changing the bounds themselves requires a visible schema/privacy review; no private environment override may bypass them.
- A separate loopback-only simulation profile can use small thresholds and accelerated clocks for quick tests. Production images reject that profile. Publish effective nonsecret settings, schema version and hash with deployment evidence.
- Provide `make config-check CONFIG=config/gati.yaml` and `make config-show CONFIG=config/gati.yaml`. Changes take effect through a coordinated validated restart/deployment, not an unaudited admin panel. Each gathering pins its functional configuration/map version until expiry; new candidates use the new version. Privacy restrictions and lower retention caps apply immediately; discard incompatible unpublished state rather than retaining it under old rules. Never retroactively lower a live gathering's threshold or move its destination.
- Include tests that changing each event threshold produces the expected behavior, that a higher threshold prevents early notification, that unrelated thresholds do not change, and that invalid production/simulation overrides fail before serving requests.

## 5. Aggregate publication and notification privacy

Publish precomputed snapshots, never participant markers or arbitrary query results.

- Fixed public areas, one published geographic resolution, and globally aligned five-minute releases with a one-epoch delay. No user-selected aggregate query radius, time range, availability filter, or exact count endpoint. The chosen travel radius filters eligibility privately; it does not produce a custom public count. Published gathering cards may identify an activated crossing once public release rules permit it; that marker represents a collective destination, never an individual. Include a snapshot timestamp and live admission revalidation when opening its join action.
- Suggested public minimum: 20 contributions. Labels: `20+ GATI`, `50+ GATI`, `100+ GATI`, etc. Arrival layers below the public minimum are suppressed even if a private **JEMI KËTU** notification was emitted.
- Use one canonical immutable public result for an area/epoch across map, statistics, area-follow/large-nearby notifications and exports. Authenticated participant invitations and arrival updates use separate continuous event decisions and never include counts; their lower thresholds must receive a separate inference review. Pan/zoom selects existing polygons; it does not recompute groups. Do not publish overlapping parent totals or complementary state counts that reveal suppressed cells by subtraction.
- Bound one session's contribution to one area and gathering per epoch. Delay publication, use stable release rules, and test temporal differencing, joins, withdrawals, bucket crossings, colluding probes and neighboring-area inference. Never claim these rules eliminate all inference.
- Initially show recent protected releases and a delayed daily city-wide activity summary derived only from released data. Label measurements as temporary signals, not unique people. Do not sum snapshots into a misleading unique-participant figure or persist raw inputs to compute history.
- Keep live snapshots for 15 minutes at origin/CDN; retain only approved daily broad summaries for 30 days. Public information can be copied indefinitely by others; origin expiry cannot revoke it.
- If the privacy review finds the release rules inadequate, coarsen geography/time, increase thresholds, or withhold the affected numeric surface. Do not silently ship it. Formal differential privacy with an explicit contribution model and privacy budget is a later specialist option, not a home-grown noise function.

Foreground notifications use a short authenticated status poll with jitter/backoff, initially every 30 seconds while visible. Session creation, updates and join responses also return any already available invitation immediately; continuous matching itself does not wait for a client poll. Polling is sufficient at MVP scale if measured; persistent sockets are unnecessary initially.

Optional Web Push requests permission only after an explicit user action. Store its endpoint and keys separately under an expiring notification handle, limited to the session lifetime or a maximum 24-hour area follow. Push subscriptions are identifying/correlatable delivery capabilities, and involve browser push providers [3]. Background push therefore has a clear privacy tradeoff and is never required to participate. Expire/unsubscribe server-side and clean up the browser subscription when possible; providers may retain their own metadata.

Lock-screen payloads are generic, such as “Ka një përditësim në GATI.” Fetch the relevant **JEMI GATI**, meeting information or **JEMI KËTU** after opening. Deduplicate by notification handle/event/epoch, cap delivery (initially one update per five minutes and six per hour), coalesce events, and expire failed deliveries. The cap applies to push delivery, not in-app state updates or joining; always fetch the latest state on opening. Area follows store one coarse area or a bounded list, no travel history; large-nearby alerts use these same public snapshots and coarse areas, with configurable bucket-aligned counts and nearby distance. Include a gathering reference only when its publication rules allow it, so the recipient can open the current gathering and join. Follows alone do not register willingness. Foreground-only operation remains fully functional when push is unavailable. Local development uses a fake delivery sink; live push interoperability is a separate opt-in check.

## 6. Data inventory and expiry contract

Initial maximum lifetimes below must be enforced in code, indexes, queues, caches, browser storage and deployment configuration. Changes require a privacy review.

| Data | Who can access it | Maximum retention / deletion |
| --- | --- | --- |
| Exact location | Browser/OS location provider | No app persistence or network transmission; discard after cell conversion |
| Willingness + hashed capability | Restricted API/worker store roles | 120 minutes; immediate logical removal on cancellation |
| Candidate memberships / time indexes | API/worker | No longer than their underlying signals; cleanup within 60 seconds |
| Gathering destination and aggregate event state | API/worker | Gathering deadline, at most 60 minutes by default; no retained participant history |
| Founding/late admission, going, decline and arrival links | API/worker | Earlier of session expiry and gathering end; arrival freshness max 15 minutes; denial/cooldown state no longer than session |
| Arrival nonce and replay state | API | Nonce max 2 minutes; spent state until its relevant session deadline |
| Push endpoint, delivery keys, follow areas | Notification worker, restricted registration API | Session deadline, or explicit area follow ≤24 hours; delete on unsubscribe |
| Notification queue / retry entries | Notification worker | ≤5 minutes, and never beyond destination handle expiry |
| IP address | Network boundary in transit | No request/access logs; no durable storage |
| Abuse counters | Edge/API limiter | Epoch-keyed HMAC of address/prefix, ≤10 minutes; destroy rotation secrets, no joins to signals |
| Public snapshots / daily summaries | Everyone | 15 minutes / 30 days at origin; external copies cannot be recalled |
| Operational metrics | Operator | ≤7 days; broad counters/latencies only, no tokens, IPs, cells or per-event labels |
| Map extract, crossing index, schema, build/config artifacts | Everyone | Durable, contains no participant activity |

Treat address-derived rate-limit keys as sensitive temporary pseudonyms, not anonymized data. Never put capability hashes, request bodies, coordinates, cells, endpoint URLs, headers or personal state in logs, trace spans, error reporting, CI artifacts or metrics. Disable access logs at every controlled proxy and app layer. Document the DNS, host and optional edge providers' independent visibility.

Valkey keys have explicit TTLs, but set/sorted-set members require separate expiry discipline. Every read checks the logical deadline; worker cleanup bounds index retention. Native expiry is not a promise of immediate physical byte erasure [2]. Disable persistence, swap and core dumps for the runtime environment; deny snapshot/debug/replication commands to service roles and keep Valkey off public interfaces. Reject new writes under memory pressure rather than evicting arbitrary live state. Store live state on no backed-up volume. Restarts may lose active sessions: explain this neutrally in the UI and do not invent recovered attendance.

Back up only reviewed configuration, map extracts and crossing indexes and approved public summaries, with secrets handled separately and encrypted. Never back up participant state. Test this by inspecting an actual restore. TTL cannot erase copies already stolen from live memory or retained by an uncontrolled provider.

## 7. Abuse controls and scaling

Start with schema/body limits, request deadlines, bounded work per request, capability-scoped limits, short-lived network-prefix limits, global and regional admission budgets, one-use nonces and atomic idempotency. Same-origin restrictions, strict CORS, CSP, safe output encoding, and no third-party executable scripts protect browser capabilities. Network limits must tolerate shared NATs; do not require a phone, CAPTCHA tracking vendor, fingerprint or persistent device identifier.

An optional short-lived computational challenge can raise automation cost under load; it must be accessible and measured on low-end phones. It does not establish unique people. Use anonymous aggregate burst/replay/error metrics to detect anomalies, and temporarily slow or disable activation in an attacked region. Fail closed on unverifiable counts and drop stale notifications; publish a generic service status without participant details. Do not blacklist Tor users by default or advertise network anonymity merely because Tor works.

For 100,000 live signals, precomputed compatibility and incremental cell/time/radius counters avoid all-pairs matching. Re-evaluate affected crossing candidates with bounded batches and stability/expiry timers, then atomically validate before activation. Set explicit limits on crossing-index batch size per cell, membership allocation, worker batch size, notification fanout and outstanding requests. Backpressure must preserve expiry and prevent unbounded queues.

API replicas are stateless against shared Valkey. A worker lease with fencing prevents duplicate publishers; transition/delivery idempotency remains required even with leases. One host/store is the low-cost starting deployment and a known availability limit. Add API replicas first; introduce regional store/worker partitions only when benchmarks justify them. Assign each signal one owning region and define boundary matching before partitioning. Do not imply that multi-key scripts transparently work across Valkey Cluster slots [6].

Public traffic uses static assets and cached canonical snapshots. Authenticated state, tokens and push registration must never enter shared caches. An optional CDN may cache only public content; TLS termination still gives it network visibility. Volumetric DoS needs upstream network capacity/protection, not just application rate limiting. Select and document the host/edge provider before public deployment; the local stack cannot demonstrate protection from a saturated internet link.

Provisional benchmark targets, to be tested rather than advertised as achieved:

- 100,000 active signals on a documented 4-vCPU/8-GB Linux reference host, with client load generated elsewhere. Also report this computer's actual hardware and results.
- 300 successful new/updated signals per second sustained; 3,000 attempted writes per second for a 60-second burst with bounded, explicit admission rejection.
- 100,000 foreground clients polling every 30 seconds: approximately 3,334 status requests/second, plus public traffic. Test 10,000 cached public reads/second separately.
- Sustained-load p95 accepted-write and status latency <300 ms; continuous-matching processing lag <5 seconds after debounce/stability deadlines at sustained load; expiry reconciliation <10 seconds. Public-map freshness is intentionally several minutes, not real time.
- No double activation/counting, expired state accepted, unbounded growth or private cache responses. Benchmark a city-wide hotspot as well as uniform traffic. Record accepted/rejected traffic and end-to-end latencies so rate limiting cannot conceal poor capacity.

Measure CPU, memory per signal/index/subscription, bandwidth, cache hit rate and notification fanout before estimating deployment cost. Set a deployment budget from those measured requirements; do not promise a cheap VPS can handle arbitrary spikes.

## 8. Local development and Tirana simulation

The following are target commands to implement, not commands available today:

```sh
make dev
make test
make simulate SCENARIO=tirana-evening SEED=42
make simulate SCENARIO=tirana-sybil SEED=42
make load PROFILE=100k
make verify-local
make down
```

`make dev` starts client, API, worker, isolated Valkey and fake notifications on loopback using Compose. A small first-party Tirana map fixture works without runtime internet access after dependencies/images are installed. Publish its provenance/license and an attributed reproducible process for obtaining a larger basemap. Do not bulk-download the standard OSM tile service; its policy prohibits offline/prefetch use [4]. MapLibre renders the local assets [5]. No live location permission is necessary for simulated users.

Use one deterministic Go simulation CLI that drives the real API and matching code. A controllable clock supports fast functional scenarios; performance tests use real wall-clock time. Explicitly separate simulation ground truth from production API responses. The simulator dashboard may show synthetic individuals only on a separate development origin with an unmistakable Albanian simulation banner. Compile simulation routes/clock overrides out of production, bind them to loopback, and test their absence in production artifacts. Never allow a deployment environment flag alone to turn them on in a public binary.

Version-controlled scenario parameters:

```yaml
city: tirana
seed: 42
population: 1000
duration_minutes: 180
distribution: clustered # also uniform, sparse, single_hotspot
availability_minutes: [30, 60, 90, 120]
radius_km: [1, 3, 5]
going_probability: 0.55
arrival_probability_given_going: 0.65
cancellation_probability: 0.15
arrival_delay_minutes: [2, 15]
location_error_meters: 300
sybil_fraction: 0.0
duplicate_request_fraction: 0.02
offline_fraction: 0.10
clock_speed: 20 # functional simulation only
# Explicit weights, grid, crossing map version, thresholds, timing and
# boundary policies are inherited from config/gati.yaml (planned).
```

Include city center clusters, Lake Park, sparse outskirts, radius/grid boundaries, incompatible availability, willingness without going, going without arrival, late joins before/after JEMI KËTU, follower-to-participant admission, duplicate and expired join links, decline while remaining GATI, arbitrary-minute 30-minute availability, threshold stability/reset, nearest-crossing ties, unreachable nearest crossing, missing crossings, map updates, destination stability under late joins, expiry before threshold, replayed/spoofed arrivals, NAT-shared users, Sybil bursts, notification failure, worker restart, cache staleness and complete store loss. Synthetic traffic must never contact a production endpoint: enforce a loopback/explicit test-target allowlist and separate development credentials.

Reports include configuration/hash/seed, synthetic ground-truth people versus accepted credentials, activations, late admissions/rejections, crossing choice and selection latency, missed compatible groups, false activations under attack, suppressions, delivery delay, load/error rates, peak memory and expiry behavior. This makes the limits of “independent arrivals” observable. Publish synthetic reports freely; real production sessions must not become fixtures.

## 9. Ordered, reviewable commits

Each numbered item is an intended commit-sized milestone. Split one if review becomes unwieldy; keep code, relevant tests and documentation together. Implementation follows this order, and each commit must leave existing local startup/checks usable. Consult PROGRESS.md and git history for completed work; this table describes intended deliverables, not completion evidence.

| # | Suggested commit | Deliverable and acceptance evidence |
| --- | --- | --- |
| 01 | `docs: define GATI product and threat boundaries` | Preserve requirements, this plan, continuous matching/late-join/crossing ADRs, data inventory and claim/evidence matrix. Record unresolved guarantees and selected MVP semantics. |
| 02 | `chore: scaffold local app and open-source project` | Go/client skeleton, Compose, Make targets, pinned dependencies, contributing and security-reporting docs. Proposed AGPL-3.0-or-later for app code, explicit docs/data licenses and notices; review ownership/license compatibility before release. Fresh clone starts on loopback. Add the central typed configuration schema, annotated defaults, validation/inspection commands and separate simulation overrides. |
| 03 | `feat: define coarse geography and Tirana fixtures` | Versioned grid, reproducible OSM pedestrian-crossing import/index, conservative reachability table and local attributed basemap. Test crossing normalization, closest eligible selection, missing map data and boundaries; no handpicked meeting catalog. |
| 04 | `feat: add expiring anonymous willingness` | Capability-authenticated create/status/cancel, one active signal per capability, no coordinates accepted, TTL schema/index cleanup and baseline rate/body limits. Test replay, expiry, cancellation and store loss. |
| 05 | `feat: build Albanian willingness flow` | Minimal responsive accessible screen, availability ≥30 minutes/radius, coarse/manual location, JAM GATI and cancellation. First useful action in seconds; inspect network/storage for location or persistent IDs. |
| 06 | `test: add deterministic local population simulator` | Seeded Tirana API-driving CLI, fake clock build and separate synthetic dashboard. Deterministic fixtures; production binary lacks simulation controls. |
| 07 | `feat: activate compatible gatherings` | Continuous matching/timers, nearest eligible mapped crossing, single allocation, stable configurable threshold and fenced/idempotent worker. JEMI GATI, late admission, going and decline-without-withdrawal. Test arbitrary-minute availability, ties, missing crossings, destination stability, concurrent joins, expiry and worker retries. |
| 08 | `feat: confirm temporary arrival claims` | Arrival nonce, local coarse position check, single accepted claim, freshness, retraction and JEMI KËTU. Test founding and late participants equally; distinguish replay prevention from unverifiable GPS/human claims. |
| 09 | `feat: publish protected collective map` | Fixed delayed snapshots, suppression/buckets, no markers, safe history summary and shared public cache. Tests for complementary totals, temporal differencing, private activation probes and cache leakage; withhold unsafe outputs. |
| 10 | `feat: add expiring notifications and area follows` | Foreground updates, fake sink, optional push, expiring subscriptions/follows, configurable nearby-event counts/radius, actionable current-gathering links, generic payloads, deduplication, caps and retry expiry. Foreground works without external services; optional real-device push matrix documented. |
| 11 | `security: harden abuse controls and data lifecycle` | Distributed rate/admission budgets, anomaly behavior, restricted Valkey roles, no persistence/logging/swap/dumps, browser headers and retention checks. Attack scenarios include Sybils, spoofing, NAT fairness and scraping; publish residual risks. |
| 12 | `perf: validate 100k signals and traffic bursts` | Real-time load harness and reproducible reports; optimize only measured bottlenecks. Include dense crossing datasets, continuous update storms and mass late admission; meet agreed reference-host targets or revise deployment sizing before claiming scale. |
| 13 | `ops: add one reproducible deployment path` | Production Compose/Caddy, pinned image digests, host preflight, automated local/CI checks, bounded monitoring, public-only backup/restore, rollback and secret rotation. Deploy/recover on a clean disposable Linux VM. |
| 14 | `build: publish verifiable release artifacts` | Locked build environment, SBOM, signed digest/provenance manifest, reproducible client/backend build checks, public build/config/schema metadata, external artifact comparison and operator host audit scripts. Explain what each check can and cannot prove. |
| 15 | `docs: complete independent audit and pilot checklist` | Ten-requirement evidence matrix, Albanian UX walkthrough, network/data-flow captures using synthetic inputs, hostile-query results, restore and expiry reports. Independent reviewer rebuilds and follows the deployment audit; resolve material findings before public pilot. |

Milestones: 01–06 local anonymous willingness and simulation; 07–10 full local product flow; 11–12 security/performance evidence; 13–15 deployable, reviewable pilot. No calendar estimate until the first slice establishes development pace.

## 10. Deployment verification and operation

One documented route: provision a supported Linux VM, run a checked-in host preflight/setup script, supply a domain and locally generated secrets, verify the signed release manifest, and run production Compose using image digests. Expose only HTTPS/HTTP for TLS and restricted operator access; never expose Valkey. Use the same immutable images in staging and production. Deployment should be an explicit operator action, with automated validation around it.

Publish a signed release manifest containing source revision, tree status, build recipe/toolchain, dependency lock hashes, image digests, client asset hashes, SBOM, schema version and public privacy configuration hash. Thresholds, grid/crossing-dataset version, retention limits, logging policy and enabled external integrations belong in public configuration; secrets do not. A self-reported `/.well-known/gati-build.json` is a comparison aid, not attestation.

Provide two distinct verification paths:

- **Any outside reviewer:** independently rebuild the documented artifacts, compare unsigned artifact bytes/digests, download served client/service-worker assets and compare them, inspect browser requests and caches, run black-box privacy/abuse probes and compare published configuration. Reproducible builds require a recorded environment and independent comparison [7]. Rebuild signatures/provenance separately from reproducible payloads.
- **Reviewer with operator-granted host access:** verify running image digests, mounts, commands, runtime configuration, firewall, store ACLs and persistence settings, logging agents, backups, swap/core-dump settings and outbound integrations. Run synthetic canaries through expiry and inspect store/index/queue cleanup. Record signed findings with a date and scope.

Neither path proves that no parallel service, covert logging, future code change, provider snapshot or privileged observer exists. Publish that gap rather than implying a badge or version endpoint solves it. Browser-delivered code can also be targeted per visitor; a separately distributed reproducible client would improve that trust boundary but is later scope.

Monitor broad service health: request/error rates, latency, worker delay, memory, queue depth, cleanup lag and cache age. No geography/session labels. Alert on unexpected persistence, old live records, growing queues and missing releases. Rollback uses a previously verified digest; incompatible ephemeral schema changes can intentionally reset sessions with a clear service notice. Restore only nonparticipant backups and test that no sessions reappear.

## 11. Requirement verification before implementation

| Required property | Architecture response | Evidence / qualification |
| --- | --- | --- |
| 1. Anonymous willingness | No registration; expiring random capability; no permanent ID | Network/storage inspection; session correlation and IP boundary documented |
| 2. Private geographic matching | On-device coarsening, fixed grid, nearest shared reachable mapped crossing | Coordinate-rejection and cell-boundary tests; live store still contains coarse area |
| 3. Automatic expiration | Server deadlines + TTL + index/queue cleanup; no live backups | Fake-clock tests, real-time canary deletion, restore inspection; no forensic erasure promise |
| 4. JEMI GATI | Continuous configurable threshold for a shared mapped crossing and overlapping remaining availability; late joining; one allocation per signal | Deterministic, race and adversarial enrollment scenarios |
| 5. JAM KËTU / JEMI KËTU | Founding/late participant one-use arrival claims and configurable freshness threshold | Replay/race tests; independent humans and true physical presence are unproven |
| 6. Aggregate maps | Canonical delayed areas/buckets; suppression; no individual interface | Differencing/composition tests and independent privacy review; no formal anonymity guarantee |
| 7. Fake signals / DoS | Layered limits, admission budgets, replay checks and upstream protection | Attack/load reports; determined Sybils and volumetric attacks remain residual risks |
| 8. 100k+ signals | Bounded incremental matching, static public reads, stateless API replicas | Reference-host benchmark required; scale is a target until measured |
| 9. Cheap maintenance | One codebase, Compose, one ephemeral store, static config | Clean-machine startup/deployment, restore/rollback drill and measured cost estimate |
| 10. Independent verification | Open source/schema/infra/tests, reproducible artifacts and two audit paths | Independent rebuild/review; remote backend honesty is not fully provable |

Public pilot gate: working Albanian flow with late joins, automatic crosswalk destinations and continuous matching of ≥30-minute availability, validated central configuration, accepted privacy limitations, no raw location transmission, bounded retention throughout the stack, no individual public information, adversarial review of private and public inference surfaces, tested 100k sizing, a clean deployment/rollback exercise, and accurate public claims. If those are not met, continue local simulation rather than labeling the product production-ready.

## 12. Primary technical references

References inform the design; defaults and exact APIs must be rechecked against the versions pinned during implementation.

1. Valkey persistence: https://valkey.io/topics/persistence/
2. Valkey expiry semantics: https://valkey.io/commands/expire/
3. W3C Push API and security/privacy considerations: https://www.w3.org/TR/push-api/
4. OpenStreetMap standard tile usage policy: https://operations.osmfoundation.org/policies/tiles/
5. MapLibre GL JS documentation: https://maplibre.org/maplibre-gl-js/docs/
6. Valkey Cluster specification and hash-slot constraints: https://valkey.io/topics/cluster-spec/
7. Reproducible Builds planning and environment recording: https://reproducible-builds.org/docs/plans/
8. OpenStreetMap pedestrian crossing tags: https://wiki.openstreetmap.org/wiki/Tag:highway%3Dcrossing and https://wiki.openstreetmap.org/wiki/Tag:footway%3Dcrossing

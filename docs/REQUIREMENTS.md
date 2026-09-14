# GATI — agreed requirements

This is a repository summary of the user's requirements and subsequent corrections.
It is not a verbatim copy of the original recap. The latest explicit user direction
takes precedence. Implementation defaults and their rationale are in the
[development plan](DEVELOPMENT_PLAN.md).

## Purpose and experience

GATI makes temporary willingness to join collective real-world activity visible.
One person stays anonymous; the collective becomes visible. It is not a social
network, chat, discussion forum, organizer platform or permanent membership system.

The core flow is:

**A JE GATI? → JAM GATI → JEMI GATI → PO, PO SHKOJ → JAM KËTU → JEMI KËTU**

The first meaningful action takes seconds: choose availability, travel radius and
press JAM GATI to obtain a fresh one-time device-location fix and submit. There
is no registration, identity form or complex onboarding. Willingness expires
automatically and is not a commitment to attend. No response means no attendance.
Declining/cancelling is neutral, easy and never guilt-inducing.

## About and public purpose

The Albanian “Rreth nesh” section explicitly supports the peaceful Flamingo
protests in Albania. It values continuing daily at the usual hour in front of the
Prime Minister's office and aims to complement that gathering with continuous,
decentralized participation across the country and at different times. Do not
invent an exact scheduled hour or imply nationwide coverage: this version covers
Tirana. After meeting, participants choose how to continue peacefully themselves.

Explain the willingness → invitation → voluntary commitment → arrival flow with
the existing slogan. Reading About or returning to participation must not enroll,
request device location, change an active session or pressure attendance. Keep it
visually clear and accessible without adding onboarding or external embeds.

Privacy/security claims must match the code: no registration or identity fields,
locally coarsened device location, temporary coordination data with automatic
expiry, suppressed/delayed public aggregates, rate/replay controls and open source
with verification tooling. Do not say “zero data retention,” absolute anonymity,
verified humans/location or independently audited deployment. Provide the source
link and concise limits for inference, network/operator exposure and optional push.

## Matching, destination and joining

- Participant location comes only from the device location service. No manual
  map pin, address or area input is offered. Require a fresh, sufficiently precise
  reported fix, coarsen locally and discard exact coordinates. Device estimates
  can be inaccurate or spoofed; this is not a guarantee of genuine location.
- Default geographic cells are **100 m (0.1 km)**, explicitly selected by the user
  after reviewing the simulation comparison. The default maximum reported device
  error is 50 m, required by the whole-cell accuracy bound.
- Minimum offered availability is **30 minutes**.
- Matching is **continuous**, based on geographic proximity, compatible remaining
  availability, user-selected travel radius and configurable thresholds.
- Destinations are selected automatically from mapped **road intersections (crossroads)**,
  not a handpicked catalog of meeting places. The plan defines closest as nearest
  to the group's coarse center among intersections reachable by every counted member.
- JEMI GATI gives relevant users a destination and remaining time with
  **PO, PO SHKOJ** and **JO TANI**. No in-app communication is required.
- Newly notified people can join **current gatherings**, including those already
  marked JEMI KËTU. They do not need to have been counted in the initial activation.
- Arrival uses a fresh coarse location claim and **JAM KËTU**. Enough accepted
  arrivals produce **JEMI KËTU**. Individual arrivals are never public.
- A currently confirmed participant may explicitly press **JAM ENDE KËTU** shortly
  before presence expiry, with a fresh device fix and challenge. This refreshes one
  contribution within the original session/gathering deadlines; it never extends
  willingness, changes destination or automatically keeps a gathering alive.
- Ended public gatherings are labeled from their released deadline and removed
  from the live gathering map layer; their historical buckets expire with the
  snapshot. Expired gathering metadata is pruned in bounded cleanup/reconciliation.
- The requested independence of arrivals is a security goal, not a solved property:
  the current design establishes distinct credentials, not distinct humans.

## Notifications, map and statistics

Support JEMI GATI, relevant gathering information and JEMI KËTU notifications.
Background push is off by default and explicitly chosen by each user; participation
and in-app updates work without it.
Optional area follows and sufficiently large nearby activity alerts are included.
They must reveal only necessary information and offer an actionable join flow.

Show significant GATI areas, bucketed activity such as **50+ GATI**, activated
gatherings, sufficiently confirmed gathering areas and simple aggregate statistics
over time. The public statistics map represents collective cells, never individuals
or exact gathering destinations. Gatherings use the same public cells as willingness;
exact crossroads remain available through private invitations, explicit eligible
preview and admission. Previewing never enrolls or commits the user.
After JAM GATI, show published approximate nearby willingness. Gathering cards must
show approximate going and fresh-arrival statistics before and after JAM KËTU.
Keep the Tirana collective activity map available throughout these states. These
statistics use the same privacy-preserving release rules across cards and maps.

Suppress or coarsen small groups. Mark deliberately delayed data with its time;
do not describe historical releases as exact current occupancy.

Counts for activation, arrival confirmation and large-nearby alerts, plus radii,
timing and other functional settings, must be easy to configure and validate.
Simulation settings may differ from production under an explicit isolated profile.

## Privacy and hostile use

Privacy is architectural. Collect no name, email, phone, profile, social graph,
permanent participant identifier, unnecessary device identifier or continuous
location history. Location is coarse wherever possible, temporary, purpose-limited
and discarded when no longer needed. No tracking, analytics/advertising SDKs,
secret participant database or hidden individual admin interface.

Reduce exposure to location inference, repeated-query/differencing and correlation,
and document database-compromise and administrator risks. The user explicitly
accepted inferred information for implementation (decision 0007), including the
public statistics surfaces; do not block development on eliminating these attacks.
Keep suppression, fixed delayed buckets, temporary credentials and cell-only public
gathering geography. No raw counts, individual records or exact participant positions
may be published. The original absolute inference guarantee remains unmet; describe
counterexamples honestly. Device-only UI does not fix spoofing or establish unique
humans. This acceptance is not authorization for public deployment.

Assume fake mass willingness, fake arrivals, notification flooding, API abuse,
scraping, DoS, identification attempts and manipulated statistics. Use appropriate
rate/admission limits, temporary credentials, geographic/query bounds, caches,
edge protection, anomaly detection and privacy-preserving abuse resistance.
Requiring identity is not an acceptable solution.

## Scale, maintenance and verification

Design and benchmark for **100,000+ active signals** and large traffic spikes.
Prefer stateless API replicas, TTL-based data, efficient geographic aggregation,
precomputed public releases, caching and bounded work. Avoid needless microservices.

Maintainability for one or a few developers matters: low cost, few dependencies,
simple local development, automated deployment/cleanup, bounded monitoring,
appropriate nonparticipant backups and one documented deployment process.

Publish the client/backend source, infrastructure configuration, storage schema,
tests, data-flow documentation and threat model. An independent reviewer should
be able to determine what enters the service, storage/lifetimes/access, compromise
exposure and identification risks. Use reproducible/verifiable builds where
practical, and document what actual-deployment checks can and cannot establish.

## Local development and identity

Develop and test on this PC before deployment. Include reproducible, parameterized
Tirana map/population simulations with deterministic seeds, configurable behavior
and hostile-use cases. Keep synthetic ground truth out of production interfaces.

The user-facing MVP is entirely **Albanian**. Use 🦩 and
**Bëje vullnetin tënd të dukshëm.** The design should be minimal, contemporary, warm
and human. No profiles, feeds, unnecessary forms, gamification, rankings, streaks,
badges or participation history.

Advance availability is deferred and excluded from the current MVP review. Native
mobile apps are not required for the initial browser/PWA implementation.

## Before claiming completion

Launch sequencing: the first public version is a direct public alpha, without a
separate volunteer stage (decision 0015). Improve it from optional sanitized reports
and bounded private operational monitoring, without participant tracking. Device
emulation is not real-phone, GPS, installed-PWA or external-push evidence. Actual
deployment controls and honest disclosure of untested behavior remain necessary.

Explain and provide evidence for anonymous willingness, coarse matching, expiration,
JEMI GATI, arrival confirmation, aggregate privacy, hostile-use controls, 100k scale,
maintainability and independent verification. Prefer a simpler implementation when
an added feature makes the product less anonymous or harder to maintain.

## Selected usability implementation (decision 0008)

Implement optional background notifications, one-action willingness, private
location-eligible destination preview before PO, PO SHKOJ, and clearer waiting,
connection recovery and arrival-expiry guidance. Do not add sharing links/QR or
new purpose/activity copy in that earlier work. The subsequently approved About
section above supersedes that copy exclusion. See USABILITY_IMPLEMENTATION_PLAN.md.

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
provide a coarse area or optional one-time location fix, then press JAM GATI. There
is no registration, identity form or complex onboarding. Willingness expires
automatically and is not a commitment to attend. No response means no attendance.
Declining/cancelling is neutral, easy and never guilt-inducing.

## Matching, destination and joining

- Minimum offered availability is **30 minutes**.
- Matching is **continuous**, based on geographic proximity, compatible remaining
  availability, user-selected travel radius and configurable thresholds.
- Destinations are selected automatically from mapped **pedestrian crosswalks**,
  not a handpicked catalog of meeting places. The plan defines closest as nearest
  to the group's coarse center among crossings reachable by every counted member.
- JEMI GATI gives relevant users a destination and remaining time with
  **PO, PO SHKOJ** and **JO TANI**. No in-app communication is required.
- Newly notified people can join **current gatherings**, including those already
  marked JEMI KËTU. They do not need to have been counted in the initial activation.
- Arrival uses a fresh coarse location claim and **JAM KËTU**. Enough accepted
  arrivals produce **JEMI KËTU**. Individual arrivals are never public.
- The requested independence of arrivals is a security goal, not a solved property:
  the current design establishes distinct credentials, not distinct humans.

## Notifications, map and statistics

Support JEMI GATI, relevant gathering information and JEMI KËTU notifications.
Optional area follows and sufficiently large nearby activity alerts are included.
They must reveal only necessary information and offer an actionable join flow.

Show significant GATI areas, bucketed activity such as **50+ GATI**, activated
gatherings, sufficiently confirmed gathering areas and simple aggregate statistics
over time. A map represents collective areas/destinations, never individuals.
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

Protect against location inference, repeated-query/differencing attacks, temporary
signal correlation, database compromise and administrator abuse. Public interfaces
must not allow individual participation reconstruction. Do not claim that listing
residual risks satisfies an absolute guarantee; track unresolved properties visibly.

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

Explain and provide evidence for anonymous willingness, coarse matching, expiration,
JEMI GATI, arrival confirmation, aggregate privacy, hostile-use controls, 100k scale,
maintainability and independent verification. Prefer a simpler implementation when
an added feature makes the product less anonymous or harder to maintain.

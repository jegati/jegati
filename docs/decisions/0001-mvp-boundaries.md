# 0001 — Local-first anonymous MVP boundaries

Date: 2026-09-13. Status: working implementation direction from the agreed plan;
not evidence that any privacy control or feature is implemented.

## Context

GATI needs an actionable collective gathering flow without identity, organizers,
chat or permanent history. It must be practical to develop locally, audit and
deploy as a small project. User corrections require continuous matching, a
30-minute minimum availability, automatic crosswalk destinations, late joining
and easy functional configuration.

## Decisions

1. Use a TypeScript/Vite PWA, MapLibre, one Go API/worker codebase, temporary Valkey
   and Compose/Caddy. Pin versions during setup. Add services only for a measured
   need. Live participant state has no persistence or backups.
2. Convert coordinates to coarse cells on the device. Use temporary random
   capabilities to control sessions. Source data must describe their bounded
   linkability, not promise that actions within a session cannot be linked.
3. Continuously evaluate shared crossing reachability and overlapping remaining
   availability. Choose the eligible crossing nearest the coarse group center;
   break ties deterministically. Import crossings from versioned map data. Keep
   activated destinations/deadlines fixed; missing map coverage does not authorize
   invented destinations. Coarse matching may miss boundary opportunities.
4. Admit newly notified users to still-open gatherings after revalidation. Going
   does not count as arrival. Founding and late participants use the same one-use
   arrival claims. JO TANI declines one gathering without ending willingness.
5. Separate continuous private decisions from delayed canonical public releases.
   Publishing a precise collective destination still needs an inference review.
   Public and private thresholds are distinct, with observable surfaces reviewed
   together. No public member lists or individual locations.
6. Use one typed functional config with visible defaults, validation, version/hash
   and documented privacy bounds. Production rejects simulation settings. No admin
   dashboard is necessary to change these settings.
7. Use deterministic, isolated local simulations and fake push delivery before
   testing external push. Background push is optional and introduces provider
   metadata exposure. Real participant data must never become test fixtures.
8. Reproducible builds and privileged host audits provide different evidence.
   Publish their scope and limitations; a version endpoint is not attestation.

## Consequences and unresolved properties

- Distinct credentials and browser location do not establish independent humans
  or genuine physical presence. Sybil resistance remains bounded.
- A privileged host operator can inspect live memory or add logging. Memory-only
  storage reduces retained exposure but does not solve malicious administration.
- Suppression, bucketing and delay are not a formal anonymity proof. Test private
  probes, temporal differencing and crossing-selection inference before publication.
- Restarting the ephemeral store can lose sessions; do not reconstruct attendance.
- Crosswalk map tags are not a guarantee of pedestrian space or crowd capacity.
- The 100k capacity, operating cost and reproducible deployment remain targets
  until their checks run. Public release depends on the plan's evidence gates.

## Superseded assumptions

Do not reintroduce the original 15-minute availability option, fixed gathering
time slots, handpicked meeting-area catalog, or admission restricted to founding
participants. Those were corrected before implementation.

## Updating decisions

For a material change, add a numbered Markdown record with context, decision,
consequences, validation and any superseded record. User-directed changes should
be recorded directly; routine internal choices do not require extra approval.

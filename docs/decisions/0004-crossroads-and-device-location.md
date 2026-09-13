# 0004 — Crossroads, device-only input and bounded inference claim

Date: 2026-09-13. Status: accepted product correction; validation in PROGRESS.md.

The user answered the nearest-destination/privacy decision with “yes”, corrected
“crosswalk” to “crossroad”, and required device-service location with no manual
input. This supersedes the pedestrian-crosswalk and optional-location assumptions
in decisions 0001/0002 and the earlier plan. Retain the exact nearest eligible
landmark rule and continue local development with the disclosed collusion risk.
This does not resolve the original absolute inference requirement or authorize
public deployment, remote pushes, identity collection or device fingerprinting.

A crossroad is a mapped street intersection, including T-junctions. Derive it from
shared OSM node IDs with at least three distinct adjacent street nodes. Geometry
alone cannot establish connectivity. Use the conservative street/access filters
in geography/import.go and data/tirana/README.md. Keep the activated landmark and
deadline fixed; revalidate all counted radii against the same coarse cells. Refer
to pedestrian space beside the junction; do not direct users into traffic or infer
a safe sidewalk point from a road center. No handpicked meeting catalog is added.

Use one-shot browser geolocation for both willingness and arrival. No map-click,
map-center, address, coordinate or manual-area fallback may set participant location.
Inspect reported age/accuracy on device under visible config bounds; never upload
exact coordinates or the accuracy estimate. Ask explicitly when the user acts;
do not track in the background. If location is denied/unavailable/poor/outside the
service area, explain in Albanian and allow a retry. Map browsing remains possible.

Device-service input reduces accidental/manual misplacement but cannot guarantee
accuracy or prevent spoofing. The browser can be modified, and an anonymous API
accepting coarse cells cannot attest their origin. Do not claim GPS, verified
presence or independent humans. The colluding-input probe remains meaningful even
without a manual location UI. Retain its historical crosswalk evidence and rerun
it against the new intersection dataset; do not relabel a known failure as a pass.

Schema 5 explicitly changes map/config/private destination fields. This is a local
prototype migration: recreate the disposable store and discard the old client
session format. Do not interpret old crosswalk records as intersection gatherings.
No participant state is migrated or backed up. Future deployed incompatible changes
need a documented drain/closure process, tracked in deployment hardening.

References: [OSM junction representation](https://wiki.openstreetmap.org/wiki/Junctions),
[W3C device geolocation](https://www.w3.org/TR/geolocation/),
[original inference evidence](../reports/09-inference-gate.md).

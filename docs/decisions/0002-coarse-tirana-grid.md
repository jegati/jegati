# 0002 — Fixed coarse Tirana grid and imported crossing index

Date: 2026-09-13. Status: implemented geographic foundation; matching engine pending.

Use a published Tirana bounding box (41.28..41.38 N, 19.75..19.90 E), fixed origin,
reference latitude and latitude/longitude steps derived from configured nominal
cell size. Canonical cell IDs encode grid version, size, column and row. A changed
region/grid definition needs a new version. Boundary cells retain their whole
rectangle rather than clipping to a tiny identifying sliver. Coordinates outside
the supported box cannot create a signal; map coverage remains explicitly bounded.

The browser will use the server-published grid steps to quantize locally. Only
canonical cell IDs enter participant APIs. Public map geometry is separate from
private participant locations. Rounded coordinates masquerading as cells are not
an accepted request format.

Use Orb's reviewed spherical distance routine and a conservative 6,400 km curvature
bound to cover geographic model error. Add a cell-radius bound obtained by
integrating a straight path in latitude/longitude under maximum local metric
coefficients. This is more conservative than checking only a cell center. Unit tests
sample cell interiors/boundaries and verify rejection of center-only radius matches.
It is a geographic distance rule, not a pedestrian-route or travel-time guarantee.

Build a finite cell/radius-to-crossing index at startup/offline, sorted by map ID.
Closest shared selection uses the mean of member cell centers and only crossings
reachable for every member. No crossings produces no destination. The higher-level
continuous allocation/stability algorithm remains milestone 07.

Import actual OSM nodes/ways and archive provenance. Do not manufacture missing map
features or expose individual willingness to an external map provider. Automated
filtering can miss valid options and cannot verify physical space or crowd capacity.
The fixture is a local-development basemap; public deployment still needs review.

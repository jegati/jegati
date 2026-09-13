# Tirana map fixture

© OpenStreetMap contributors. Source and derived geographic data in this directory
are available under the Open Database License (ODbL) 1.0:
https://www.openstreetmap.org/copyright and https://opendatacommons.org/licenses/odbl/1-0/.
This is third-party map data, not AGPL application code or participant activity.
Retain attribution when displaying/exporting it. Individual OSM objects identify
public map features; they are not GATI participants or arrivals.

`source/query.overpass` describes the public extract (41.28..41.38 N,
19.75..19.90 E). `source/manifest.json` records its provider, base timestamp and
SHA-256 digests. `source/overpass.json.gz` contains the exact response, compressed
with mtime=0. No user-specific API requests were involved. A first provider request
timed out; the successful complete response is retained. Source snapshots may be
incomplete or outdated and do not establish real-world intersection suitability.

Rebuild the derived assets without network access:

```sh
make map-import
make map-check
```

`intersections.json` (tirana-intersections-v1) derives crossroads from shared OSM
street node IDs with at least three distinct neighboring nodes. T-junctions and
four-way intersections qualify; bends, duplicate/split ways and geometric overlaps
without a shared node do not. Source IDs identify the junction node and its ways.
This follows [OSM's road-junction representation](https://wiki.openstreetmap.org/wiki/Junctions).

Eligible classes are primary, secondary, tertiary, unclassified, residential,
living_street and pedestrian streets. Paths, service driveways, ramps, motorway/
trunk roads, area outlines, explicitly restricted/private/disused/construction,
bridge/tunnel and motorroad ways do not contribute. Junctions touching explicitly
prohibited or grade-separated ways are conservatively omitted. A mapped roundabout
entry may qualify where three street legs connect; its central island is never
invented as a meeting point. Nearby nodes of complex junctions remain distinct.
No proximity merge invents geometry. The UI describes pedestrian space beside the
junction landmark; this import does not certify access, a sidewalk or crowd capacity.

The original archived extract includes highway ways with full node/geometry arrays
and selected pedestrian-crossing nodes. It lacks other node-only restrictions;
missing map metadata remains a limitation. The original query/checksums are preserved
for reproducibility. The obsolete crosswalk-derived artifact has been replaced;
original probe evidence still identifies its original map version.

`roads.geojson` is a minimal basemap of in-bounds road/path geometry and names.
It needs no external tiles/fonts. Features are sorted deterministically; road
segments outside the service bounding box are omitted. It does not supply routing,
transport-time estimates or certified pedestrian accessibility.

To update the source deliberately, submit the checked-in query to a public Overpass
instance, reject partial/error responses, archive the exact result with a new
manifest/hash, regenerate assets and run the tests. Increment the dataset version
when deploying a changed dataset. An independent rebuild from the archived input
must produce identical derived bytes; live queries are not reproducible over time.

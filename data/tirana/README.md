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
incomplete or outdated and do not establish real-world crossing suitability.

Rebuild the derived assets without network access:

```sh
make map-import
make map-check
```

`crossings.json` keeps eligible pedestrian crossing nodes/ways and source IDs,
normalizes crossing ways already represented by nodes, and records a shared map
reference point plus available geometry. Explicit access/foot prohibitions,
private/disused/construction features and crossings on forbidden mapped ways are
excluded conservatively. No proximity-only merge combines distinct nearby crossings.
Way-only crossings use the mean of public mapped vertices as the reference point;
the UI must describe the pedestrian space beside that crossing, not the roadway.

`roads.geojson` is a minimal basemap of in-bounds road/path geometry and names.
It needs no external tiles/fonts. Features are sorted deterministically; road
segments outside the service bounding box are omitted. It does not supply routing,
transport-time estimates or certified pedestrian accessibility.

To update the source deliberately, submit the checked-in query to a public Overpass
instance, reject partial/error responses, archive the exact result with a new
manifest/hash, regenerate assets and run the tests. Increment the dataset version
when deploying a changed dataset. An independent rebuild from the archived input
must produce identical derived bytes; live queries are not reproducible over time.

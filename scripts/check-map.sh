#!/usr/bin/env bash
set -euo pipefail
repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_dir"
source scripts/env.sh
staging=$(mktemp -d)
trap 'rm -rf -- "$staging"' EXIT
go run ./cmd/map-import -out "$staging"
cmp data/tirana/intersections.json "$staging/intersections.json"
cmp data/tirana/roads.geojson "$staging/roads.geojson"
node --input-type=module <<'JS'
import { readFileSync } from 'node:fs';
import { gunzipSync } from 'node:zlib';
import { createHash } from 'node:crypto';
import assert from 'node:assert/strict';
const manifest = JSON.parse(readFileSync('data/tirana/source/manifest.json'));
const archive = readFileSync('data/tirana/source/overpass.json.gz');
const raw = gunzipSync(archive, { maxOutputLength: 64 * 1024 * 1024 });
const hash = (bytes) => createHash('sha256').update(bytes).digest('hex');
assert.equal(hash(archive), manifest.gzip_sha256);
assert.equal(hash(raw), manifest.raw_sha256);
assert.equal(JSON.parse(raw).osm3s.timestamp_osm_base, manifest.osm_base_timestamp);
assert.equal(JSON.parse(readFileSync('data/tirana/intersections.json')).source_sha256, manifest.raw_sha256);
console.log('Map source checksums and byte-identical offline rebuild passed.');
JS

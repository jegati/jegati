// Compare functional evidence, retaining all scenario/config/input/map data and
// frames. Wall runtime, sampled heap and source build metadata are not outcomes.
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { createHash } from 'node:crypto';

const paths = process.argv.slice(2);
if (paths.length !== 2) throw Error('Provide two completed synthetic report.json paths');
const omitted = ['wall_seconds', 'max_sampled_driver_heap_bytes', 'source_revision', 'source_dirty'];
const canonical = value => {
  if (Array.isArray(value)) return value.map(canonical);
  if (value && typeof value === 'object') return Object.fromEntries(Object.keys(value).sort().map(k => [k, canonical(value[k])]));
  return value;
};
const reports = paths.map(path => {
  const r = JSON.parse(readFileSync(path));
  assert.equal(r.status, 'completed'); assert.equal(r.simulation, true);
  for (const key of omitted) delete r[key];
  return canonical(r);
});
const differences = [...new Set([...Object.keys(reports[0]), ...Object.keys(reports[1])])]
  .filter(key => JSON.stringify(reports[0][key]) !== JSON.stringify(reports[1][key]));
assert.deepEqual(differences, [], 'Functional fields differ');
console.log('Identical functional reports: ' + createHash('sha256').update(JSON.stringify(reports[0])).digest('hex'));
console.log('Excluded only: ' + omitted.join(', '));

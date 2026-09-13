import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
const path = process.argv[2];
if (!path) throw Error('Provide a synthetic report.json');
const r = JSON.parse(readFileSync(path));
const c = r.counts;
assert.equal(r.simulation, true);
assert.equal(r.status, 'completed');
assert.equal(r.effective_config.profile, 'simulation');
assert.equal(r.config_sha256, createHash('sha256').update(JSON.stringify(r.effective_config)).digest('hex'));
assert.equal(c.credentials_accepted + (c.credentials_rejected || 0), c.generated_credentials);
assert.equal((c.cancelled || 0) + (c.expired_verified || 0), c.credentials_accepted);
assert.equal((c.arrivals_accepted || 0) + (c.arrivals_rejected || 0), c.arrival_attempts || 0);
assert.equal((c.direct_joins_accepted || 0) + (c.direct_joins_rejected || 0), c.direct_join_attempts || 0);
assert.ok((c.credentials_invited || 0) <= c.credentials_accepted);
assert.ok((c.gatherings_confirmed_observed || 0) <= (c.gatherings_observed || 0));
assert.equal(r.gatherings.length, c.gatherings_observed || 0);
for (const g of r.gatherings) { assert.ok(g.ends_at > g.activated_at); assert.ok(g.first_seen >= g.activated_at); }
for (const f of r.frames) { const active = Object.values(f.cells).reduce((a,b)=>a+b,0); assert.ok(f.here <= f.going && f.going <= active); }
assert.equal(Object.keys(r.frames.at(-1).cells).length, 0);
assert.equal(r.frames.at(-1).here, 0);
for (const values of Object.values(r.by_radius)) assert.ok((values.invited || 0) <= values.accepted);
if (r.grid.size_meters === 1000) for (const radius of ['0.1','0.5']) assert.equal(r.by_radius[radius]?.invited || 0, 0);
if (process.argv.includes('--success')) { assert.equal(c.credentials_invited,c.generated_credentials); assert.equal(c.arrivals_accepted,c.generated_credentials); assert.ok(c.gatherings_confirmed_observed>0); }
if (!process.argv.includes('--allow-rate-limits')) assert.equal(c.rate_limited_requests || 0,0);
console.log(`Synthetic report checks passed: ${c.synthetic_people} people, ${c.credentials_accepted} credentials, ${c.gatherings_observed || 0} observed gatherings, expiry/funnel/config invariants.`);

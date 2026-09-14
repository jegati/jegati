# What the Tirana study actually found

Completed 2026-09-14. **33 modeled days**: nine main cases × three seeds plus two
10%-reaction cases × three seeds. Three extra 15-second-step runs check timing
sensitivity. The nine exploratory first-pass runs are not the final evidence.
These are synthetic behavioral illustrations, not predictions or observed turnout.

## Useful message

The experiments support a conditional story about **local density, visibility and
follow-through**. They do not show that any small audience automatically creates
citywide protest. Use the modest 30,000 / 3% example in the pitch, and keep the
zero-gathering result beside it. An audience of 30,000 means hypothetical app users,
not followers guaranteed to adopt it or the city's population.

All main results below use current config: 30 willing founders, 20 fresh arrivals,
100m private / 1km public cells; 60-minute assumed willingness; 80% acceptance and
85% attempted arrival; hourly map viewing. Unless stated otherwise, 25% of eligible
idle viewers decide they have time and want to join. Three-seed ranges are not
confidence intervals. [Full assumptions and reproduction](METHOD.md).

| Audience / baseline willingness | Placement / radius | Hours with ≥1 confirmed gathering, median (range) | Confirmed cells, median (range) | Mean total active, median |
|---|---|---:|---:|---:|
| 3,000 / 5% | Uniform / 1km | 0 | 0 | 151 |
| 10,000 / 3% | Uniform / 1km | 0 | 0 | 305 |
| 30,000 / 3%, no reactive joining | Uniform / 1km | 0 (0–2 minutes) | 0 (0–1) | 906 |
| 30,000 / 3% | Uniform / 1km | 4.19 (3.88–4.68) | 4 (3–8) | 1,034 |
| 30,000 / 5% | Uniform / 1km | 11.61 (11.48–11.63) | 109 (109–111) | 4,574 |
| 10,000 / 3% | Uniform / 3km | 0.27 (0.24–0.58) | 12 (11–18) | 2,035 |
| 10,000 / 3% | Three synthetic clusters / 1km | 11.60 (11.56–11.82) | 18 (16–20) | 2,035 |

“Mean total active” includes reactive sessions. In the stronger uniform case this
is about **15% of the audience**, and in the clustered case about **20%**. It would
be misleading to describe those outcomes as sustained by only 5% or 3% participation.
The 3% uniform case rises more modestly, to about 3.45% mean total active.
Uniform coverage uses the full ~140km² service rectangle, not demographic density.

## Joining-rate and timing sensitivity

Reducing reactive joining to 10%, still assuming hourly checks:

- 30,000 / 5% uniform: **11.08h** median (11.06–11.36), **94 cells** (91–101),
  mean active **2,473** instead of 4,574.
- 10,000 / 3% clustered: **10.88h** median (10.81–11.38), **14 cells** (13–14),
  mean active **881** instead of 2,035.

At 15-second rather than 30-second steps, seed 42's 30,000 / 3% case changes from
4.19h / 4 cells to **3.85h / 5 cells**. The stronger uniform case changes from
11.48h / 109 cells to **11.55h / 106 cells**; the clustered case from 11.60h / 16
cells to **11.61h / 16 cells**. Individual confirmed-gathering counts change by
14–28% across these three cases. Do not advertise precise event counts as robust
predictions. Finer timing also changes discretized random schedules; this is a
sensitivity check, not a proof of numerical convergence.

## Duration finding: short invitations, repeated throughout the day

Median invitation lifetime was approximately **15 minutes** in the reactive 1km
cases. The oldest eligible founders often have only the required remaining overlap
left. Their expiry freezes the gathering deadline; joining does not extend it.
Among gatherings that reached confirmation, median accumulated confirmed time was
**4.5 minutes** in the uniform 1km cases and **7.5 minutes** in the clustered case.
These are accumulated fresh-confirmation minutes, not measured physical crowd dwell.

This produces many successive IDs: median 73 activations / 71 ever-confirmed in the
30,000 / 3% case, and 1,878 / 1,812 in the 30,000 / 5% case. They are **not** thousands
of distinct independently organized protests. Changing the lifetime rule would be
a product decision; no matching, admission, location or expiry behavior was changed.

Wider radius is not automatically better: the 3km case produced a median 473
invitations but only 15 ever confirmed, with 16 minutes of citywide confirmed time.
Walking and short remaining lifetimes limit follow-through even when geometric
matching succeeds. Map crossroads are not a guarantee of safe pedestrian space.

## Evidence and checks

Source model commit `7dfeba7`; no model changes after that commit for these runs.
Reports mark the working tree dirty because it also contained the user's About
edit and reporting/sensitivity files. The model never reads `web/index.html`.
Config SHA256: `e1c698921caf6151fe6f6f2ad245a8529caa8829b88176848d18656b94ce1ff7`.
Full map/study hashes and per-case metrics: [summary](tirana-day-summary.json).
Finer-step provenance: [timing comparison](timing-sensitivity.json).

Raw local reports: `reports/local/tirana-day-final.json`, `tirana-day-15s.json`,
`tirana-reaction-sensitivity.json`. Raw reports contain synthetic event lifecycles
and frames, no exported participant identities or trajectories. They are generated
artifacts, not tracked participant records. The standalone report and shareable
PNG/PDF exports are at `reports/local/gati-pitch/index.html`, with PNG/PDF exports in its `media/` directory.
Main matrix runtime: 208.07s; lower-reaction matrix: 72.06s; finer-step runs: 77.26s,
excluding each process's geographic-index startup. These concurrent local model
measurements are not server performance benchmarks.

Focused Go race tests cover deterministic replay, stationary availability,
no discovery without published events, no arrival inferred from willingness,
finite audience/funnel, frozen deadlines, public count suppression and invalid
assumptions. Matching/activity/geography regression packages and Go vet passed.
Python checks preserve zero cases, seed ranges, and reject incomplete or incompatible
summaries. Chromium report checks cover desktop/mobile layouts, controls, all
available seeds, empty outcomes and absence of external requests, with PNG/PDF export.
Browser emulation is not human usability evidence.

Independent API/Valkey control (`outreach-api-control-fixed`): 80 accepted/invited/
arrived synthetic credentials, 39 late admissions, five verified arrival retries,
80 verified expiries and one confirmed gathering. Associated real-store clock,
continuous activation/admission, renewal and known inference checks passed. The
initial control failed because newly enabled outbound push is forbidden for a
simulation store; the isolated setup now explicitly disables push in its generated
copy while retaining the source config. No real push delivery was tested.

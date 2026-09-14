# Isolated Tirana simulations

[Completed 3,000-person experiment findings](reports/population-3000.md) include
the 18-case matrix, finer-grid comparison, failures found and repeatability evidence.

The main population scenario now uses **3,000 synthetic people** and the current
`config/gati.yaml` settings. Each run owns a fresh loopback API/Valkey, drives the
actual willingness/admission/arrival endpoints, and removes its processes/store.
No real device location, external map request or production participant is used.

```sh
make simulate-population
make simulate-population SEED=43
make simulate-population SCENARIO=tirana-success
make simulate-suite
# Explicit 100 m profile, now equivalent to the default:
make simulate-population SIM_CONFIG=config/simulation-population-100m.yaml
```

Defaults create timestamped directories under `reports/local/`. Use `OUTPUT=...`
for a specific new directory; existing population reports are not overwritten.
Open `index.html` for the standalone Albanian map replay or suite comparison table.
Each run includes `report.json`, `timeline.csv`, effective/source config snapshots,
binary/map hashes and local platform information. `make check-population
REPORT=path/to/report.json` checks the funnel, expiry, geometry and config invariants.

The script validates the selected application config, converts its profile to
`simulation`, and disables outbound push in the run directory. Production-enabled
push cannot run against a simulation store. All matching/publication parameters
are retained. It never edits your input file.
The API remains a simulation-only build in the `gati-sim:*` namespace. Current
baseline: activation 30, arrivals 20, radii 0.1/0.5/1/3 km and 100 m cells.
`config/simulation-population.yaml` is an explicit example of these settings;
the default command reads your current gati.yaml instead of that static example.

`config/simulation-population-100m.yaml` retains the previously labeled comparison using 100 m cells
and at most 50 m reported device error; it now matches the default. The historical
1,000 m experiment baseline could not match either
0.1 or 0.5 km radii under the conservative whole-cell rule. Finer cells expose a
more precise area and still exclude some boundary matches. See decisions 0005/0006. Historical report snapshots retain their original settings.

## Deterministic refactor comparisons

Run `make test-planner-replay` to compare eight fixed-time planner snapshots from
3,000 seeded participants using the full 6,491-crossroad dataset and the current
`config/gati.yaml` (100 m cells). It checks proposals, founder selection/deadlines,
offers and exclusion after declining, then repeats each snapshot with shuffled
input order. A reviewed JSON fixture pins configuration/map/input hashes, visible
test summary counts and a digest of every decision. An intended configuration,
map or behavior change requires inspecting the resulting differences before
updating `internal/simulation/testdata/tirana-planner-replay.json`; the test never
automatically overwrites that fixture. CI includes this target.

This test performs no HTTP requests and issues no credentials. Its fixed labels
exist only in `planner_replay_test.go`, which is excluded from application builds.
The three-proposal bound per snapshot and synthetic halfway cancellation rule are
fixture choices, not application settings or predictions of a complete gathering
funnel. Normal `make test` skips this slower full-map fixture; use its explicit target.

Keep running `make simulate` and the API population scenarios for storage,
arrival, response and notification behavior. Those end-to-end runs deliberately
retain cryptographic credentials and gathering IDs: the same population seed can
still produce small differences in cohort tie-breaking and observed outcomes.
The new fixture makes planner comparisons reproducible without weakening runtime
randomness or changing the simulation's default behavior.

## Scenario inputs

Edit the explicit YAML files in `simulation/scenarios/`. Population is bounded at
10,000; behavior runs cap worst-case generated credentials at 30,000. Main case:
`tirana-population.yaml` (3,000), plus dense/low-followthrough/churn (3,000 each),
sparse (100), boundaries (200), and a guaranteed-success control (80). These are
synthetic distributions, not Tirana demographic/attendance estimates.

| Input | Meaning |
| --- | --- |
| `seed`, `population` | Repeatable random input and synthetic people, distinct from credentials |
| `distribution` | `clustered`, `uniform`, `sparse` edges or `single_hotspot` |
| `availability_minutes`, `radius_km` | Choices sampled for actors; must also be accepted by selected app config |
| `cancellation_probability` | Probability an actor cancels during its availability |
| `duplicate_request_fraction` | Fraction exercising idempotent creation and arrival retries |
| `sybil_fraction`, `sybil_credentials` | Synthetic people generating multiple credentials (1–10 copies) |
| `behavior.enrollment_minutes` | Period over which people become available at unaligned times |
| `behavior.duration_minutes` | Observation duration; covers enrollment plus longest availability, ≤240 minutes |
| `behavior.network_groups` | Actual separate loopback source addresses (1–512), not fake forwarding headers |
| `behavior.writes_per_second` | Real write pacing (1–250), at most half selected global write budget |
| `behavior.snapshot_seconds` | Report sampling, 30–300 virtual seconds; does not schedule matching |
| `behavior.accept_probability`, `decline_probability` | Invitation decisions; remaining probability ignores the invitation |
| `behavior.arrival_probability_given_going` | Fraction of accepted journeys attempting arrival; others are planned no-shows |
| `behavior.late_join_probability` | Fraction considering synthetic observer-discovered gatherings at enrollment |
| `behavior.retract_probability` | Fraction withdrawing a successful arrival claim after one virtual minute |
| `behavior.offline_probability`, `offline_seconds` | Paused status polling and its duration range |
| `behavior.response_seconds` | Inclusive range for response delays |
| `behavior.speed_meters_per_second`, `detour_multiplier` | Declared journey approximations; not pedestrian routing |
| `behavior.availability_weights`, `radius_weights` | Probabilities aligned to choices; each list sums to one |
| `behavior.clusters` | Public synthetic centers, spread in km and weights summing to one; used only for clustered placement |

The simulator checks the actual API responses; willingness/notifications do not
count as going or arrival. Declines preserve willingness, admissions revalidate
reachability/deadlines, nonce replays cannot extend freshness and final reads verify
expiry. An actor's synthetic position moves to its destination when its journey
arrives. Journeys interrupted by changed invitations and those exceeding session
lifetimes are separately counted. One actor can accept multiple invitations over
its temporary session: reports distinguish unique actors from response/claim events.

## Reproducibility and limits

A stable event queue orders enrollment, polls, responses, journeys, snapshots and
worker deadlines. Enrollment times are distinct to avoid random capability hashes
choosing different founders on ties. Simulation builds have no wall-time background
worker; authenticated clock advances explicitly drive the same real matching worker.
Normal builds retain their real timer. Clock `run_worker:false` exists only in the
simulation wrapper; ordinary APIs and production binaries do not expose it.

Functional time can advance quickly, but native TTLs and abuse windows use real
time. Main case uses 256 declared source networks with original budgets, and 200
writes/second pacing. Rate rejections are counted and are not hidden as unmatched
people. Runs time out after 20 real minutes, below the shortest offered 30-minute
native TTL. This tests logical expiry, not forensic erasure or 100k throughput.
A suite freezes its selected config and runs up to three isolated cases concurrently.
Wall times include concurrency effects and must not be compared as load benchmarks.

The replay shows synthetic active credentials and the last observed gathering state.
Brief collective transitions between participant responses may not be seen; the
report does not pretend to have continuous privileged observation. Separate API
integration tests verify exact stability/expiry boundaries. Per-radius geometry
statistics compare exact synthetic starts and whole-cell reachability, not whether
a compatible group actually existed. A small exhaustive planner oracle tests
bounded fixtures; city-wide optimal allocation is not claimed.

Bulk actors use coarse API inputs. The real client is tested separately through
mocked-device browser journeys on its normal clock, including denied/stale/poor
fixes and the full gathering flow. No manual participant-location UI is added.
Legacy synthetic late discovery is injected from already observed gatherings; it
does not exercise the implemented public map or push delivery. Area follows remain
unimplemented.
No secrets are exported in reports. Source/map/config hashes identify the experiment;
known location spoofing and colluding-input inference limitations remain.

## Legacy checks

`make simulate` and `make simulate SCENARIO=tirana-sybil SEED=42` retain the smaller
willingness-only population driver with `config/simulation.yaml` (3/2 thresholds),
then run separate real-store activation/arrival integration checks. They are not
city response-population reports. Legacy map counts are cumulative accepted inputs,
not current occupancy. See reports/06-simulation.md for historical evidence.

`make privacy-probe` remains a deliberately failing known-inference diagnostic.
See [SIMULATION_PLAN.md](SIMULATION_PLAN.md) for the original experiment plan and
[MATCHING_PARAMETERS.md](MATCHING_PARAMETERS.md) for application controls.

## Reviewing and repeating experiments

`SEED=...` overrides the scenario's `seed`; omitting it now respects the YAML.
A suite uses seeds 42/43/44 explicitly. After correcting a failed check or code path:

```sh
python3 scripts/run-simulation-suite.py --output reports/local/my-suite --resume
```

Resume uses the suite's saved config, rechecks completed reports, and archives
incomplete/failed runs under `failed-attempts/` before rerunning them. It preserves
the previous summary. Reports record each run's source revision and binary hashes;
a resumed suite can contain evidence from more than one revision. Never run two
writers against the same output directory.

To check exact functional repeatability (after `source scripts/env.sh`):

```sh
node scripts/compare-population.mjs path/to/first/report.json path/to/repeat/report.json
```

This compares full frames and metrics, excluding only wall runtime, sampled driver
heap and source build metadata. Scenario/config/map/input differences still fail.
The 120-person `tirana-shared-network` scenario deliberately exercises the unchanged
creation budget. Check it with `node scripts/check-population.mjs path/to/report.json
--allow-rate-limits`; rejected credentials must remain visible in its results.
The small success control and standalone replay check are included in the CI
workflow; a local pass does not assert that remote CI has executed.

## Day-long outreach study

The separate [offline day study](../docs/outreach/METHOD.md) reuses the matching
planner and public release builder for a finite audience becoming willing across
12 hours. It models reactive joining from released public events and compares
multiple seeds. It is a behavioral illustration, not the API/Valkey simulator,
a turnout prediction or human usability evidence.

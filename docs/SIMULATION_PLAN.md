# Tirana population and interaction simulation plan

Status: implementation authorized and underway, 2026-09-13. The user increased
the main population from 300 to 3,000 and selected their current configuration.
See SIMULATION.md and PROGRESS.md for implemented commands and validation.
The design below incorporates that correction; original effort estimates remain
planning estimates, not a measurement of coding-agent elapsed time.
This takes priority over continuing milestone 09. Existing commands and evidence
are in SIMULATION.md; product settings are explained in MATCHING_PARAMETERS.md.

## Implemented scope and remaining limits

The response runner now drives willingness, decisions, journeys, arrival/retraction,
late admission, cancellation and expiry through the actual Tirana API. It uses the
archived 6,491-junction map, isolated memory-only Valkey, scheduled virtual worker,
seeded weighted population, real network budgets and standalone timeline replay.
See SIMULATION.md for runnable commands and PROGRESS.md for validation state.

Separate API integration tests cover continuous activation, going/decline, late
admission, arrival replay, JEMI KËTU and expiry. Thirteen existing browser checks
cover the device-only client and a scripted gathering flow. Those are regression
evidence, not observations of realistic city-wide behavior. Public aggregate maps,
area-follow alerts and push delivery are not implemented; do not report them as
simulated product features.

The proposed suite tests system behavior under explicit human-behavior assumptions.
It cannot measure whether real Albanian users understand the UI, trust the privacy
claims or actually attend. That later requires a small consent-based human test.
All locations, decisions and behavior probabilities below are synthetic assumptions,
not demographic or attendance estimates for Tirana.

## First experiment and scenario matrix

Start with 3,000 synthetic people over three simulated hours, joining at unaligned
moments during the first hour. An illustrative distribution is 40% central Tirana,
25% near the lake, 20% northwest and 15% elsewhere within the supported rectangle.
Reuse existing cluster centers, make their centers/spread/weights explicit, and
validate all generated points against the service boundary. Match only with the
imported crossroads and actual conservative coarse-cell radius calculation.

Availability distribution: 30/60/90/120 minutes with weights 40/35/15/10%.
Travel radii: 0.1/0.5/1/3 km with equal 25% weights, matching the user’s config. Initial invitation outcomes:
60% accept, 25% decline, 15% ignore; 80% of accepters attempt arrival. Response delay
is sampled between 5 and 90 seconds. Conditional no-shows and cancellations are
separate from declining. Begin with scripted guaranteed-success cases before
interpreting stochastic results. Compare seeds 42, 43 and 44 for each city scenario.

| Scenario | Population/timing | What to observe |
| --- | --- | --- |
| Typical mixed evening | 3,000, weighted clusters, arrivals spread over 60 minutes | Time to invitation, selected crossroads, unmatched areas, going-to-arrival conversion |
| Sparse city | 100 on the outskirts, same radius distribution | Honest no-match outcomes, radius exclusions, expiry without pressure |
| Dense hotspot | 3,000, strong central cluster with a 10-minute enrollment burst | Stable founding, late admission into existing gatherings, duplicate/competing destinations, admission rejection |
| Low follow-through | Same 3,000-person input distribution as evening, 25% accept and 40% of accepters arrive | JEMI GATI without JEMI KËTU; willingness never treated as attendance |
| Churn and late responses | 3,000 with cancellations and delayed newcomers | Stability resets; late admission before/after JEMI KËTU; cutoff/expiry and JO TANI retention |
| Geographic/time boundaries | Small constructed cohorts, then mixed 200-person case | 30-minute minimum, cell edges, unreachable crossroads, near-cutoff arrival and location-quality failures |

Add targeted hostile/network cases: one crowded NAT, distributed networks, replayed
arrival, one actor with multiple credentials, lost API responses, disconnection and
reconnection. Reuse the existing privacy counterexample as a separate known-failure
diagnostic. Successful simulated gatherings must not be presented as privacy proof.

## Actor model and execution

An actor has a seeded synthetic start point, enrollment time, availability, radius,
response policy and journey parameters. Use the actual API for every state change:

1. At enrollment, quantize synthetic location to the same coarse grid, then POST
   willingness. Count rate-limited/rejected credentials separately from accepted ones.
2. Poll own status with configured visibility/poll/jitter rules. Record when the
   backend activated and when the actor first observed the invitation separately.
3. After the sampled response delay, send going or decline, or deliberately ignore.
   JO TANI keeps willingness; do not turn an ignored invitation into a no-show.
4. For actors going, estimate travel duration from geographic distance, a declared
   detour multiplier (initial range 1.2–1.8) and speed (0.8–1.6 m/s). This is a journey
   approximation, not OSM pedestrian routing. Actors can cancel, fail to arrive or
   reach the destination after their deadlines; let the real API decide acceptance.
5. Request a fresh arrival nonce and submit a coarse arrival claim only on attempted
   arrival. Model fresh/stale or noisy device fixes in the browser sample. Observe
   confirmation stability, expiry and withdrawal without secretly refreshing claims.
6. Offer late actors a gathering reference from the simulator's synthetic observer,
   explicitly labeled as injected discovery. Use actual join validation. This tests
   admission, not unimplemented area notifications or public-map discovery.
7. Continue through final session/gathering expiry, then verify expected 410 responses,
   bounded links and no surviving logical participation. Remove the disposable store.

The bulk simulation drives API actors; it does not validate device-service origin.
The existing 13 Playwright checks exercise the real client with mocked device
geolocation, including denial/poor accuracy and the full gathering flow. They use
the normal real-time API, separately from accelerated bulk actors. A standalone
replay browser test verifies the exported map/slider without external requests.
No real GPS fix or manual-location product option is used.

Advance virtual time through scheduled actions AND worker/poll/expiry deadlines;
do not jump over ten-second stability intervals. Give simultaneous enrollments a
stable sub-second ordering to avoid random capability-hash tie-breaking changing
founding cohorts. Keep seed streams stable per actor; capabilities remain random,
unpublished and temporary. Normalize random gathering IDs to report-only aliases.
Control the scheduler so background worker timing does not secretly alter fast-run
results; preserve the real worker code and separately test normal timer execution.
Record source/map/config hashes, seed, clock mode and input/output hashes.

## Configuration and network controls

Application decisions and human assumptions are separate inputs. The population
runner validates and snapshots current config/gati.yaml, changing only the profile
to simulation. `SIM_CONFIG` selects an explicit alternative. Baseline thresholds:
30 willing / 20 arrivals, radii 0.1/0.5/1/3 km, cells 1,000 m. Retain the existing
3/2 profile for small deterministic regressions. The explicit 100 m comparison
also lowers maximum reported device error to 50 m; it does not change the baseline.

The actor limit is 10,000, with a 30,000-credential worst-case bound for response
simulations. Main behavior runs use 256 actual loopback peers and pace writes at
200/second, preserving application budgets. No forwarding headers bypass limits.
Shared-network/abuse runs use one peer and report rejections separately. The
virtual clock never advances real rate windows. A crowded-NAT control therefore
measures the creation budget, not inability to form geographically compatible groups.

Do not disable production controls or add production test-network overrides. If a
fast functional run needs larger simulation-only budgets, expose the full diff and
repeat the relevant abuse cases with unmodified budgets and real time. Put a wall
time limit on fast runs so native TTL expiration cannot silently distort fake time.

## Measurements and output

Produce a standalone first-party Tirana HTML replay, JSON/CSV summaries and a short
Markdown findings report for each batch. The replay has play/pause, time slider,
scenario/seed/config labels, collective area counts, crossroad destinations and
state transitions. Optional actor inspection exists only in the isolated synthetic
report, never a production participant endpoint. No capabilities or nonce secrets
are exported. Main reporting remains Albanian and clearly marked SIMULIM.

Measure:

- Funnel: generated people, credentials attempted/accepted/rejected, invitations
  observed, accept/decline/ignore, arrival attempted/accepted/rejected and expiries.
- Backend activation time versus participant-observed wait (median/p95, plus counts
  still unmatched at the end; no percentile calculated from an empty sample).
- Gathering count, observed time to JEMI KËTU, planned no-show events, late-join success,
  active/fresh arrival trajectories and cancellation effects.
- Distance/travel-time estimates, selected intersection/source IDs, unreachable or
  late attempts, coarse-grid false negatives relative to synthetic exact geometry.
- Matched fraction by scenario area/radius (availability-stratified reporting is a follow-up); repeated invitations and
  ignored people, reporting small synthetic groups only in the isolated report.
- Missed compatible groups using a small independent exhaustive oracle on bounded
  fixtures. City-wide heuristics get a clearly labeled diagnostic, not a claim of
  globally optimal allocation based on the same planner that is being tested.
- Wall runtime and broad memory/resource measurements, distinct from simulated
  duration. This 100–3,000-person suite cannot establish 100k production capacity.

Follow-up experiments, after reviewing the requested current-config baseline:
use paired inputs/seeds to compare activation thresholds 10/20/30 and arrival
thresholds 10/15/20 in separate one-factor sweeps. Do not run every cross-product
initially. Follow with targeted radius choices and 30/60-second polling comparisons.
Do not quietly adopt whichever settings yield the most gatherings: report waiting,
travel, false activation, privacy and no-show tradeoffs together.

## Acceptance and incremental commits

| Proposed commit | Deliverable / gate | Focused developer effort |
| --- | --- | --- |
| 1. Scenario schema and event scheduler | Weighted spatial/temporal inputs, behavior fields, config selection, deterministic ordering, safe peers/pacing; strict validation/isolation tests | 4–6 h |
| 2. Responses and journeys | Actual API going/decline/ignore/late join/nonces/arrival/retraction; deadline-aware journey scheduler and focused state assertions | 4–7 h |
| 3. Replay and metrics | Timeline/first-party map, funnels, matched/unmatched metrics, input hashes and isolated synthetic report exports | 3–5 h |
| 4. Browser sample and experiment report | Mocked-device journeys, targeted regressions, initial 18 runs, comparisons, documented findings and limitations | 3–6 h |
| Contingency | Timing, limiter or report inconsistencies found during the runs | 2–4 h |

Total planning estimate: **16–28 focused developer hours**, roughly **2–4 working
days** for one developer. This is an effort estimate, not a promise about elapsed
coding-agent time. First useful response-enabled run: about **8–13 hours** into
implementation. Existing willingness-only scenarios need no new feature work and
can be rerun within a short local session (budget roughly 5–15 minutes for both,
including checks/report inspection).

For the new suite, initially budget **30–90 minutes** of machine execution and
inspection for 18 runs (six city scenarios × three seeds), after implementation.
Calibrate this with the success control and 3,000-person runs; report actual times before
expanding sweeps. Real-time validation takes its actual scheduled duration; fast
clock results are not load benchmarks. Broad performance regressions or structural
matching fixes would be estimated separately if the experiments expose them.

No new paid service, map API, hosting account or credentials are planned: use the
installed local Go/Node/Docker/Playwright toolchain and archived map. Electricity
and existing coding-assistant usage still have their usual cost; no monetary quote
is inferred from this estimate. Public notification delivery, production aggregate
maps, full route planning, human recruitment and the 100k benchmark are outside this
estimate.

Before calling the suite complete: verify deterministic replay where intended,
founding/destination/deadline/radius invariants, equal late/founding arrival checks,
no false attendance from willingness, unmodified client device-only input, expiry,
production test-route absence, and report mismatches as findings rather than hiding
failed runs. Preserve each run under a distinct ignored local output directory and
commit only reviewed synthetic summaries.

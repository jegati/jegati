# Tirana response simulations — 3,000-person baseline

Executed locally on 2026-09-13. All **18 final scenario checks passed** after the
fixes described below. This is synthetic functional evidence, not a human usability
study, anonymity proof or 100k capacity benchmark.

## Inputs and reproducibility

The baseline uses the user's current `config/gati.yaml`: **30** willing credentials
to activate, **20** fresh arrivals to confirm presence, radius choices
**0.1/0.5/1/3 km**, **1,000 m cells**, minimum availability **30 minutes**. Only the
run profile changes to simulation. No application rate budget was raised.

Main population: 3,000 synthetic people enrolling over 60 virtual minutes, observed
for 180 minutes and final expiry. Equal radius weights; availability weights
40/35/15/10% for 30/60/90/120 minutes. Four declared spatial clusters, 60% accept,
25% decline, 15% ignore, 80% conditional arrival attempts, 15% cancellation.
Responses take 5–90 seconds; journeys use 0.8–1.6 m/s and a 1.2–1.8 detour factor.
These probabilities are assumptions, not measured Tirana behavior. See
[scenario files](../../simulation/scenarios/) and [parameter guide](../MATCHING_PARAMETERS.md).

The driver uses real API transitions, 256 loopback source addresses, 200 paced
writes/second and scheduled virtual matching. Native TTLs/rate windows remain real.
The first suite ran three cases concurrently; an additional finer-grid comparison
and some controls overlapped. Host: AMD Ryzen 7 7435HS, 16 logical CPUs, about
23.2 GiB RAM, Linux 6.17. Go/Node/Compose versions are pinned in
[scripts/toolchain/versions.env](../../scripts/toolchain/versions.env).

The initial 18-case batch took **929.22 seconds (15.5 minutes)** including the failed
attempt and checks. The baseline seed-42 driver took 246.0 seconds during that batch;
the finer-grid driver took 270.9 seconds, excluding its roughly 37-second index
startup. A post-fix baseline repeat took 202.8 seconds under different concurrent
load. These timings are not comparable throughput benchmarks.

## Outcomes

Invited/arrived columns count credentials with at least one such event; these runs
use one credential per synthetic person. Gathering counts are distinct gatherings
observed through participant replies. “Ever JEMI KËTU” is not final occupancy.

| Scenario | Seed | Accepted | Invited | Arrived actors | Gatherings | Ever JEMI KËTU |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| boundaries | 42 | 200 | 0 | 0 | 0 | 0 |
| boundaries | 43 | 200 | 0 | 0 | 0 | 0 |
| boundaries | 44 | 200 | 0 | 0 | 0 | 0 |
| churn | 42 | 3000 | 960 | 196 | 14 | 2 |
| churn | 43 | 3000 | 1194 | 314 | 26 | 3 |
| churn | 44 | 3000 | 1073 | 261 | 22 | 5 |
| dense | 42 | 3000 | 1507 | 810 | 1 | 1 |
| dense | 43 | 3000 | 1495 | 808 | 1 | 1 |
| dense | 44 | 3000 | 1451 | 796 | 1 | 1 |
| low-followthrough | 42 | 3000 | 986 | 123 | 17 | 0 |
| low-followthrough | 43 | 3000 | 1255 | 151 | 27 | 0 |
| low-followthrough | 44 | 3000 | 1119 | 133 | 25 | 0 |
| population | 42 | 3000 | 1107 | 489 | 27 | 15 |
| population | 43 | 3000 | 1091 | 368 | 19 | 8 |
| population | 44 | 3000 | 1076 | 444 | 21 | 14 |
| sparse | 42 | 100 | 0 | 0 | 0 | 0 |
| sparse | 43 | 100 | 0 | 0 | 0 | 0 |
| sparse | 44 | 100 | 0 | 0 | 0 | 0 |

All final suite runs accepted their generated population with **zero HTTP 429s**.
All accepted credentials were either cancelled or verified expired; the final
sampled frame has no active signals or fresh arrivals. Sparse and boundary mixes
formed no group. Separate bounded geometry/API tests exercise the precise boundary
cases; an empty mixed scenario is not evidence that every boundary rule was tested.

The dense scenario concentrated eligible actors into one gathering, including late
admissions. The low-follow-through cases formed 17–27 gatherings but none reached
JEMI KËTU. Churn cases reached it only 2–5 times. This distinguishes willingness
from observed arrival under the declared behavior model.

For baseline seed 42: 1,107 people were invited (36.9%), 1,003 selected PO, PO SHKOJ at least once,
and 489 confirmed arrival at least once (16.3% of the population). There were
**746 arrival confirmations**, including subsequent gatherings; those are not
746 different people. The driver observed 132 successful direct joins, 449 other
late going responses and 11 joins after JEMI KËTU. These are event counts and must
not be added as disjoint people. Discovery for direct joins is injected from
synthetic observers; public notifications/map discovery are still unimplemented.

Median first-invitation wait among invited people was 0 minutes (many immediately
received an invitation to an existing open gathering); p95 was 20.26 minutes. The **1,893 never invited**
are excluded from that percentile, not treated as satisfied users. Median modeled
journey was 17.23 minutes. Planned no-shows (459), journeys outliving the session
(521), and journeys superseded by another invitation (736) are distinct event
categories, not a unique-person attendance rate. This suggests reviewing journey
length, admission deadlines and repeated invitations before a real pilot.

## Explicit 100 m-cell comparison

The alternative config changes cell size to 100 m and maximum reported device
error to 50 m. Thresholds, radii, behavioral parameters and seed remain the same.
The user's baseline file stays at 1,000 m. Bulk API actors do not exercise the
browser accuracy check; that setting is required for a valid fine-grid config.

| Seed 42 result | 1,000 m cells | 100 m cells |
| --- | ---: | ---: |
| Accepted people | 3,000 | 3,000 |
| Invited people | 1,107 | 1,891 |
| People confirming arrival | 489 | 1,067 |
| Arrival confirmation events | 746 | 1,498 |
| Observed gatherings | 27 | 41 |
| Ever JEMI KËTU | 15 | 30 |
| Invited with 0.1 km radius / 759 actors | 0 | 11 |
| Invited with 0.5 km radius / 738 actors | 0 | 503 |
| Invited with 1 km radius / 755 actors | 399 | 655 |
| Invited with 3 km radius / 748 actors | 708 | 722 |

With 1,000 m cells, the conservative maximum distance from the whole cell already
exceeds 100/500 m, so those choices cannot match. Even with 100 m cells, only
163/759 actors choosing 0.1 km have any reachable junction under the coarse rule;
572/759 have one using their synthetic exact start. Coarsening alone excludes all
otherwise reachable destinations for 409 of them. This is a geometric diagnostic,
not a count of missed compatible groups. Only 11 received an actual invitation.

Finer cells disclose a smaller participant area. Sampled API resident memory was
roughly **53 MB** for 1,000 m versus **430 MB** for 100 m, reflecting the full static
cell/radius/junction index. These were process samples, not measured peaks. The
report's heap metric measures only the driver. Neither setup certifies anonymity,
real device accuracy, safe pedestrian access or 100k-user capacity.

## Failures found and validation

- The initial three sparse checks failed because a zero-length Go gathering slice
  encoded as null. The checker now accepts that valid outcome while checking its
  reported count. Rechecking the existing completed reports passed; they were not
  relabeled as rerun simulations.
- Dense seed 42 failed with HTTP 503 at **16:20:00 UTC**, exactly when a network
  rotation secret expired between separate SET NX and GET calls. Commit `6f46384`
  makes selection/read atomic with Valkey time. A real-store regression checks the
  final millisecond before rotation, expired old secrets, and concurrent replicas.
  The dense rerun completed and passed. Its failed report remains archived locally.
- The shared-network control accepted **55 of 120** people; all 55 arrived and one
  gathering confirmed. The existing creation limit counts retries too. Its rate
  rejections are expected and explicitly allowed by that diagnostic's checker,
  not by the geographic baseline checks. Shared-NAT exclusion remains a product
  tradeoff; a credential does not establish a distinct human.
- The 80-person success control completed the entire flow with all 80 invited and
  arriving. A separate seed-43 control verified that YAML seed selection works.
- `make verify-local`, `make test-store`, legacy simulation/clock/arrival integration,
  all **13 Chromium app tests** against rebuilt Compose, and the standalone replay
  browser test passed. The checker rejects deliberately corrupted expiry evidence.
  CI now includes the small success/replay checks; remote CI has not been run.

Baseline seed 42 was repeated before and after the limiter fix. **All functional
JSON fields, including the entire timeline, match exactly**. The comparison excludes
only wall runtime, sampled driver heap and source revision/dirty metadata. Canonical
functional SHA-256: `f9141abd66bc40039ba48131ddab20ac98404b28da460182312a69295c69c13a`.
This is deterministic functional evidence, not a reproducible release-build claim.

## Artifacts and next experiments

- [Open the local 18-case comparison/replays](../../reports/local/tirana-suite-3000-current/index.html).
- [Open the 100 m-cell replay](../../reports/local/tirana-population-100m-42/index.html).
- [Reviewed JSON summaries and hashes](population-3000-summary.json): 18 suite cases,
  fine-grid comparison, two repeats and three controls. Full local reports contain
  frames, CSV, configuration snapshots and binary/map hashes, with no capability or
  nonce secrets. Large synthetic replay files are ignored by Git.
- [Run/configuration guide](../SIMULATION.md), including resume/recheck commands.

The suite preserves its initial summary and failed attempt. Most evidence uses Go
source at `29dd406`; the corrected dense rerun and post-fix repeat use `efb3f8c`.
Builds recorded dirty working trees, including unrelated untracked material and
report/documentation edits. Go source was held fixed throughout the initial batch;
binary hashes identify each experiment. Do not treat these as clean release builds
or evidence of a remote deployment's honesty.

Follow-ups are explicit: paired threshold/polling sweeps, availability-stratified
outcomes, route-aware journey research, and consent-based human usability testing.
Public aggregate maps, follows/push, production privacy hardening, 100k load tests
and deployment verification remain the subsequent development milestones. The
known nearest-junction collusion inference issue is still disclosed in decision
0004; successful simulations do not resolve it.

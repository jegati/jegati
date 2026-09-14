# Tirana day study: read before sharing the charts

This is an **offline synthetic behavioral model**, not a prediction, real participant
observation, human usability study or deployment/load benchmark. It changes no app
matching settings. The shareable Albanian pitch is in `PITCH.sq.md`.

## Reproduce

From the repository root, with the pinned toolchain (`source scripts/env.sh`):

```sh
go test -race ./internal/simulation -run TestDay -count=1
go run ./cmd/daystudy -output reports/local/tirana-day.json
go run ./cmd/daystudy -study simulation/studies/tirana-reaction-sensitivity.json -output reports/local/tirana-reaction.json
go run ./cmd/daystudy -study simulation/studies/tirana-day-15s.json -output reports/local/tirana-fine.json
python3 scripts/render-daystudy.py reports/local/tirana-day.json reports/local/tirana-day --supplement reports/local/tirana-reaction.json
node scripts/check-daystudy.mjs reports/local/tirana-day/index.html reports/local/tirana-day-media
```

Choose a new output filename for every run; the CLI refuses to overwrite evidence.
Edit `simulation/studies/tirana-day.json` for behavioral assumptions and seeds.
The two additional study files specify lower reactive joining and finer time
resolution; they do not override application settings. Application rules still come from `config/gati.yaml`. No credentials, network
endpoint, store, device location or third-party map service is used by this model.
An independent short API/Valkey control uses `make simulate-population
SCENARIO=tirana-success`; it is separate evidence, not an end-to-end day study.

## Meaning of x% and y%

- N means a hypothetical audience who know/have the app, **not Tirana's population**.
  No census or adoption data is used. One synthetic actor has at most one active
  session. Synthetic labels never enter production or exported actor records.
- x is the stationary baseline probability of willingness at any instant. Each
  actor has fixed 60-minute willingness periods separated by random idle periods
  with mean `60*(1-x)/x` minutes. Initial remaining willingness is uniformly sampled
  to avoid an artificial synchronized start. The measured mean is reported, not
  assumed to equal x*N exactly. The app starts the experiment with no gatherings;
  existing willingness is a cold-start modeling choice.
- Everyone is assumed to check the map once per hour at a staggered, fixed phase.
  **y is the chance an idle, geographically eligible viewer of a released gathering
  decides they have time and want to join**. It includes willingness and availability,
  and is sampled only once per hourly opportunity. It is not y% of city residents,
  not a push-delivery rate and not repeated trials every simulation tick. Reactive
  joiners have 60 minutes of new willingness; they add to baseline activity.
- Baseline schedules use an independent random stream, unchanged in paired y=0 and
  y=25% cases. A baseline start overlapping an existing reactive session does not
  create a second session or extend it. Locations are fixed synthetic origins;
  this is not a model of home/work commuting or journeys back between sessions.
- Baseline invitees accept with probability 80%; others decline that invitation.
  Reactive joiners already decided to go. 85% of going decisions attempt arrival.
  Walking time is coarse-origin straight-line distance * 1.3 / 1.2 m/s, rounded up.
  It is **not** a pedestrian route, street safety or crowd-capacity assessment.
- Arrivals are assumed to submit valid location claims at the destination. Planned
  dwell is 30 minutes, clipped to session/gathering deadlines; explicit departure
  removes presence. Each eligible presence renewal succeeds behaviorally with
  probability 80%. Missing it loses fresh presence; physical presence and fresh
  app confirmation are therefore distinct.

## What is reused, what is approximated

The real `matching.Planner` and full imported Tirana intersection index select
cohorts/destinations, apply whole-100m-cell reachability, prefer open gatherings,
exclude declined invitations and freeze deadlines. The real `activity.Build`
applies public member suppression, count buckets and 1km aggregate cells. Reactive
actors only discover events that this builder released; private/raw events cannot
trigger reactive enrollment. Model admission rechecks the remaining-time cutoff.

The behavioral scheduler is independent of the API/store/worker: 30-second steps
round response/travel/stability timing upwards. Logical activation and presence
transitions are modeled, not actual transactions, push, geolocation, retries, abuse
limits or worker races. Snapshot capture is instantaneous; publication is observed
at the next model step. The tool explicitly rejects slower-than-step publication
or extra delayed epochs rather than pretending to simulate them. Map report frames
are saved every 5 minutes; duration metrics integrate every model step, not frames.
A finer-step sensitivity run accompanies headline results. The 20:00 boundary
censors still-open gatherings; time totals count only the observation window. Production matching
remains continuous; it is not changed to half-minute slots.

Current settings require 30 willing founders and 20 fresh arrival confirmations.
A gathering expires at the **earliest founder's willingness deadline**, capped at
60 minutes after activation. With a stationary population, older founders may have
only a little more than the required 15 minutes left. New arrivals do not extend
that deadline. A day containing many successive gathering IDs is not one protest
lasting all day, and a JEMI GATI invitation is not a gathering with people present.

## Geography and outcome definitions

Uniform placement covers the entire fixed service rectangle (19.75–19.90 E,
41.28–41.38 N), roughly 140 km², including areas unlike Tirana's inhabited core.
It is deliberately **not** population-weighted. The clustered sensitivity uses
three explicitly synthetic centers and uniform 0.025° × 0.02° rectangles around
(19.818,41.327), (19.822,41.307), (19.799,41.335). It cannot support claims about
coverage of the entire city. All road geometry is the repository's OSM-derived data.

- Activated gatherings: distinct fixed-deadline invitations during the window.
- Ever confirmed: invitations reaching 20 fresh claims for the stability interval.
- Hours with any confirmed gathering: union of time across the city, maximum 12h.
- Confirmed gathering-hours: sum over simultaneous gatherings; may exceed 12h.
- Confirmed cells: distinct 1km cells with a confirmed gathering at some time,
  **not** total city area served, simultaneously active cells or named neighborhoods.
- Peak simultaneous confirmed: maximum concurrent confirmed gatherings.
- Public map: suppressed/bucketed releases only, shown as cells, never exact
  destinations or people. Analytical duration overlays are synthetic aggregate
  study results, not a new public history feature of the application.

All matrix seeds (42/43/44), including failures to gather, must remain in the
summary. Median and min–max describe just these runs, not confidence intervals or
empirically calibrated probabilities. Display a representative seed explicitly;
never select a flattering seed and call it typical without checking all results.
No political turnout, attention, attendance, renewal or safety assumption here has
been measured. No claim of causal real-world protest impact follows from the model.

## Final pitch PDF

The five-page Albanian handout adds the product pitch to three sustained-activity
scenarios: clustered 10k/3% with 10% reaction, uniform 30k/5% with 10% reaction, and
uniform 30k/5% with 25% reaction. All three maps/headlines use seed42; the intervals
include seeds42/43/44. Shared color bins show accumulated confirmed gathering-minutes,
not people or uninterrupted occupancy. Negative cases and limitations remain on
page5. It is a prelaunch pitch, not a claim that jamgati.com is already serving users.

After generating the reports above:

```sh
python3 scripts/build-pitch.py reports/local/final-pitch --main reports/local/tirana-day.json --reaction reports/local/tirana-reaction.json
node scripts/export-pitch.mjs reports/local/final-pitch
```

Output: `GATI-Pitch-dhe-Simulime.pdf`, HTML, five page previews and a provenance/hash
manifest. The builder rejects changed featured assumptions pending copy review;
the exporter refuses to overwrite a PDF and checks page count, margins/footer
collisions and external requests. Browser/fonts/creation metadata can affect PDF
bytes; its checksum identifies the artifact, not a reproducible-build guarantee.
The current reviewed final copy is under `reports/local/gati-pitch-final/`.

# Selected usability work: implementation and evidence

Date: 2026-09-13. Plan: ../USABILITY_IMPLEMENTATION_PLAN.md. Selected suggestions
2/3/5 are implemented. Suggestion 1 has server-side offer discovery; subscriptions,
Web Push transport and browser opt-in/resume remain pending. Sharing links/QR (4)
and additional purpose copy (6) remain excluded.

## User-visible checks

JAM GATI acquires fresh device location and submits in one action. Acquisition can
be cancelled; late device callbacks do not submit. A public-map join offers an
eligible private crossroad preview before commitment. New preview capabilities stay
in memory and existing participation is unchanged. JO TANI/escape closes the preview
without enrolling. Expired previews cannot confirm; offline/online controls recover.
Waiting, retry, server expiry/restart and arrival freshness have Albanian guidance.
Expired arrival returns JAM KËTU without automatic renewal.

19 Chromium tests passed: real API willingness/invitation/arrival tests plus public
activity/preview fixtures. Real-store fixed-clock tests separately exercise preview
published-ID/reachability/immutable-claim gating, strict inputs and no enrollment or
public-release changes. The worker offers invitations without an HTTP status poll
or push subscription. `make verify-local`, `make test-store`, `make simulate`, focused
Go race tests and container checks passed. UI tests use synthetic device fixes;
there was no real-device location verification or human usability study.

## Population scenario

```sh
make simulate-population SCENARIO=tirana-population SEED=42 OUTPUT=reports/local/tirana-usability-3000-42
make check-population REPORT=reports/local/tirana-usability-3000-42/report.json
```

The runner snapshots current functional config and changes only the profile for
isolated simulation. Private cells remain 100 m, public cells 1 km, activation 30,
arrival threshold 20. Seed 42 uses 3,000 synthetic Tirana participants with configured
radius/availability mixtures, response delays, trips, declines, offline intervals,
late joining and cancellations. The report validator passed all count, configuration,
expiry and funnel invariants.

| Observed metric | Result |
| --- | ---: |
| Accepted credentials | 3,000 |
| Distinct invited credentials | 1,921 |
| Distinct credentials going | 1,716 |
| Distinct arrived credentials | 1,112 |
| Accepted arrival confirmations, including renewed claims | 1,589 |
| Gatherings observed / ever confirmed | 46 / 30 |
| Joins after JEMI KËTU | 22 |
| Cancelled / expiry verified | 456 / 2,544 |
| HTTP requests | 262,875 |
| Wall time | 222.49 seconds |

Offers now reach background sessions before their next foreground poll; this can
change later synthetic choices and cohort formation. Older seed-42 count baselines
are historical. This run is not a controlled performance comparison: functional
time is accelerated and other development checks were running on the PC.

The report records source d564fce and its source-dirty flag; artifact/dataset/config
hashes and hardware metadata are in [the synthetic summary](usability-3000-summary.json).
A subsequent small correction bounds the worker's in-memory cursor by its signal's
expiry even after leadership loss; Go race and fixed-clock real-store checks passed
again after it. The population results describe the recorded build, not an independently
reproduced final release. Population discovery still uses the synthetic scheduler;
this run does not measure the browser preview journey or push/provider delivery.
It is not a 100k capacity benchmark or deployment certification.

[Local synthetic replay](../../reports/local/tirana-usability-3000-42/index.html).

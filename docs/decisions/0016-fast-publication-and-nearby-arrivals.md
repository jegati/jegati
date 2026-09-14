# 0016 — Fast public activity, nearby arrival and available opt-in push

Date: 2026-09-14. Status: user-directed implementation; supersedes the old public
publication floor and server push default in decisions 0001/0007 and the plan.

The user requested the reviewed usability changes, push active by default, nearby
arrival acceptance, and public counts published after seconds instead of minutes.
"You can also be close" is interpreted as widening the arrival area. Per-person
push remains optional and requires an explicit action and browser permission.

- Production transport enabled, public VAPID contact https://jamgati.com. Service
  keys are private operational secrets. No automatic permission prompt, enrolment,
  subscription or provider request occurs just from opening the app or JAM GATI.
- Arrival accepts the destination's 100 m cell and its eight neighbours, clipped to
  the service grid. This approximately 300 m square is not a 100 m circle or proof
  of GPS/human presence. Farther cells still fail. Fresh device fixes, one-use
  nonces, explicit going, replays, renewal eligibility and deadlines remain checked.
- Default public epoch 10 seconds, extra delay 0, capture budget 5 seconds, snapshot
  retention 1 minute. Validator supports intervals >=5 seconds and delay >=0.
  Public coordinates remain 1 km cells; threshold 20 and buckets remain unchanged.
  No per-user counts or queries. Public cache age is capped by the release interval
  and expiry; browser refresh derives from the interval (bounded 5–30 seconds).
  Normal button-to-screen visibility may take multiple intervals due to capture,
  cache and polling. No promise of instantaneous or exact counts.
- The publisher checks every second with current settings and only starts a
  zero-delay capture when its whole budget fits before the release boundary.
  Incomplete/late captures still fail closed. Fallback releases expire normally.
  Capture work is now more frequent; new scale/monitoring evidence is necessary.
- Initial form selects 30 minutes and 1 km when offered, otherwise the first
  configured radius. All existing availability/radius choices remain. Copy explains
  travel-time allowance, optional background push/tab recovery, pedestrian space
  without an invented corner, and that an app deadline does not imply people left.

## Privacy and operational consequences

Removing the multi-minute buffer materially weakens temporal privacy. A coalition
with 19 credentials can correlate an added participant with a 20+ release within
seconds. Threshold/bucket crossings and disappearance can expose timing or changes
through repeated observations; suppression is not formal anonymity. Existing
inference acceptance permits this change; it does not waive direct disclosure,
coarse geography, no tracking, TTL, private credentials or rate limits.

Nearby acceptance reduces false rejection from coarse boundaries but also broadens
false-presence claims. Browser/device GPS can be spoofed. No exact participant
coordinates or accuracy measurements are sent to the server to compensate.

Changing the functional config changes its public hash. Older releases/benchmarks
and the existing overnight run retain their recorded configuration, and cannot
certify the new cadence or arrival policy. Restart a new lifecycle run from the
committed implementation. Actual OS/provider push and host capacity remain separate
checks; do not silently claim them from mocked transport/browser tests.

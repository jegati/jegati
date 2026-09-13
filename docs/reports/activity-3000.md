# Activity publication implementation: local evidence

Date: 2026-09-13. Scope: activity policy/publisher/API/browser slice, not background
push, a human usability study, a 100k benchmark or public deployment certification.
Decision 0007 accepts documented inference and requires cell-only public gatherings.

## Implemented and exercised

- Shared 1 km public cells, separate from 100 m private matching; no intersection
  IDs/coordinates/labels in published gathering entries.
- Current willingness, going and fresh-arrival lower-bound buckets, minimum
  suppression, fixed delayed releases, bounded capture and expiry.
- Private JEMI KËTU may coexist with a suppressed public arrival figure. Expired
  arrival claims cannot preserve a public confirmed-state label below the private
  confirmation threshold, even before worker reconciliation.
- Immutable, fenced aggregate keys and cached API; no per-request participant scan,
  no caller-specific count queries or authenticated shared caching.
- Albanian nearby/gathering cards, always-visible first-party city map, cell layers
  and explicit joining with fresh device input and capability-preserving retries.
- Push remains unimplemented/off. The plan requires explicit user opt-in; neither
  provider calls nor service-worker/IndexedDB resume storage were added here.

## Checks

`make verify-local`, `make test-store`, `make simulate SCENARIO=tirana-evening` and
`make check-containers` passed. The full Chromium suite passed 17 tests. Four activity
browser tests were rerun after final display corrections. Browser statistics use
fixed public response fixtures; real-store and fixed-clock HTTP tests independently
exercise the real publisher, private arrival transitions and suppression. The full
browser-through-real-publisher journey at normal five-minute epochs is not claimed.

The inspected screenshot is [the synthetic arrival/statistics view](../../reports/local/activity-here.png).
It deliberately shows the separate private destination map and public cell map.

## 3,000-person run with the publisher enabled

Command:

```sh
make simulate-population SCENARIO=tirana-population SEED=42 OUTPUT=reports/local/tirana-activity-3000-42
make check-population REPORT=reports/local/tirana-activity-3000-42/report.json
```

The runner used the current functional config, changing only the profile for the
isolated store/clock build. The compiled run started from commit 284c0e3 with a dirty
working tree during UI work; this is disclosed in the manifest, not a reproducible
clean-release claim. The final configurable-confirmation expiry guard was added
subsequently and covered by a focused regression; default 20/20 thresholds in this
run are unaffected by that correction.

Results: 3,000 synthetic people/accepted credentials; 1,891 distinct credentials
invited; 1,067 distinct credentials arrived; 1,498 accepted arrival confirmations
including renewed claims; 41 gatherings observed, 30 ever confirmed; 25 joins after
JEMI KËTU; 456 cancellations and 2,544 expiry checks. Report validation passed.
All count metrics match the previous 100 m seed-42 baseline.

Wall time: 221.33 seconds on this PC, while other development/browser checks ran.
Functional time is accelerated. This is not a controlled performance comparison.
The publisher executes during clock steps, but population discovery still uses the
existing synthetic scheduler; it does not yet drive choices from the public map or
an optional push sink. That planned integration remains incomplete.

[Reviewed synthetic summary and hashes](activity-3000-summary.json).
[Local replay](../../reports/local/tirana-activity-3000-42/index.html).
No credentials, notification endpoints or real participant input are in this report.

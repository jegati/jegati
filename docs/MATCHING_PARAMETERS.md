# Matching parameters: current implementation reference

Checked against schema 7 and the implementation on 2026-09-13. Values below are
the current operating settings, not recommendations inferred from the experiments. Bounds
are validation choices, not privacy guarantees. Cross-field checks can reject
combinations even when each value is individually within its range.

## Where to configure

- `config/gati.yaml`: normal local Compose/application functional settings.
- `config/simulation.yaml`: small isolated regression settings (activation 3,
  arrival 2), retained for legacy checks.
- `config/simulation-population-100m.yaml`: explicit 100 m-cell / 50 m accuracy
  profile, now matching the default and retained for prior experiment commands.
- `simulation/scenarios/tirana-population.yaml`: 3,000-person behavior inputs;
  see SIMULATION.md for the other scenarios. These are not application policy.
- `internal/config/types.go` and `config.go`: authoritative field names/validation.
- `GET /api/config`: running service's effective nonsecret config and SHA-256.

Validate/inspect the normal file with `make config-check` and `make config-show`.
Use `make config-check CONFIG=path/to/production-profile.yaml` for another normal
file. `make config-check-simulation` checks the current simulation file. Settings
are loaded at startup: edit the normal file, validate it, then recreate/rebuild the
local API (e.g. `make dev` starts/rebuilds Compose). Confirm the running config/hash
at `/api/config`; editing a file is not evidence a running process reloaded it.
Gathering destinations/deadlines are frozen; use fresh isolated stores for each
comparison. Grid/map/schema incompatibilities require the documented restart/drain
policy, not reinterpretation of active records.

`make simulate-population` validates and snapshots **current config/gati.yaml**,
changing only the profile to simulation. Use `SIM_CONFIG=path/to/config.yaml`
for an explicit comparison. `make simulate-suite CONFIG=path/to/config.yaml`
runs six scenarios across three seeds. `make simulate` retains the small legacy
profile. See SIMULATION.md for implemented commands and behavior fields.

## All fields under `matching`

| Parameter | Normal / small regression setting | Actual effect | Accepted choices / constraints |
| --- | --- | --- | --- |
| `activation_count` | 30 / 3 | Size of stable founding cohort needed for JEMI GATI. Counts willing credentials, not promises or verified humans. Late joiners do not require another full cohort. | Normal 10–500; simulation 1–500; no greater than `limits.max_active_signals`. |
| `activation_stability_seconds` | 10 / 10 | Same eligible founding cohort must remain valid this long before activation. Cancellation/missing founder breaks it. | Positive integer; must fit remaining-availability constraint below. |
| `maximum_debounce_seconds` | 2 / 2 | Current worker timer interval for checking matching/deadlines. A scheduling target, not a guaranteed maximum latency under load. | 1–5 seconds, no greater than reconciliation interval. |
| `reconciliation_seconds` | 10 / 10 | Full eligibility reconciliation interval without a dirty event; also used for worker lease/recovery timing. Changes trigger earlier work. | 1–10 seconds, at least debounce interval. |
| `minimum_remaining_minutes` | 15 / 15 | Minimum availability left to count in a new founding cohort, rechecked at activation. This is not the user's minimum offered availability. | Positive; no greater than maximum gathering duration; must leave room for stability/reconciliation within minimum offered availability. |
| `maximum_gathering_minutes` | 60 / 60 | Caps duration starting at activation. Actual end is earlier if a founder's willingness expires earlier; fixed after activation. | 1–60 minutes, at least minimum remaining time. |
| `late_join_min_remaining_minutes` | 5 / 5 | Admission requires both gathering and participant to have at least this much time left. Applies even after JEMI KËTU. | Positive; no greater than founding minimum; arrival stability must fit inside it. |
| `invitation_cooldown_seconds` | 60 / 60 | Wait after declining before another automatic invitation can be offered. The declined gathering stays declined for the session. | Normal 30–300; simulation 1–300 seconds. |
| `prefer_open_gatherings` | true / true | Prefer compatible existing/open gatherings rather than founding competing ones; planner also reserves prospective destinations during stability. | Only `true` is supported. `false` is rejected. |
| `destination_rule` | `nearest_eligible_crossroad_to_coarse_group_center` / same | Choose nearest eligible imported road intersection to mean coarse founder-cell centers, reachable by all counted founders. Geographic distance; no routing. Stable map-ID tie-break. | Only this literal rule is supported; no alternate/random/manual setting. |
| `candidate_batch_size` | 100 / 100 | Maximum due reservations, background offer candidate checks, and new reservation attempts per reconciliation step (separate loops). A rotating cursor spreads offer checks across sessions; it does not require an open page or push subscription. | Currently 1–1,000,000 under generic schema bound; very large values are not performance validated. It does not change activation count. |
| `intersection_index_batch_size` | 500 / 500 | Intended future budget for incremental intersection-index processing. **Currently stored/validated but not consumed by the matcher. Changing it has no processing effect.** | Currently generic positive-integer bound 1–1,000,000. Treat as inactive until implemented. |

The exact timing constraint currently validated is:
`minimum_remaining_minutes * 60 + activation_stability_seconds + reconciliation_seconds < availability.minimum_minutes * 60`.
Matching is continuous; none of these settings creates appointment slots.
The bounded planner chooses oldest compatible founders and ranks candidate junctions
by compatible count, oldest signal, then stable map ID. There is no configurable
fairness algorithm. One credential is not one person.

## Availability and geography

| Parameter | Current default | Effect and choices |
| --- | --- | --- |
| `availability.minimum_minutes` | 30 | Lowest user choice; cannot be below 30. Must equal first choices entry. |
| `availability.choices_minutes` | `[30,60,90,120]` | Strictly increasing positive integer minute choices, 1–32 entries, within min/max. Actual app permits alternatives such as `[30,45,60]` when min/max agree. Simulation choices must be accepted by the selected application config. |
| `availability.maximum_minutes` | 120 | Highest user choice/session lifetime, at most 120; must equal last choice. |
| `geography.travel_radius_choices_km` | `[0.1,0.5,1,3]` | Ascending finite values from 0.1 to 20 km in 0.1 km increments, at most 32 entries. Simulation choices must be present in the application config. Uses conservative distance from the whole coarse cell, not walking distance or time. |
| `geography.cell_size_meters` | 100 | Nominal private geographic grid size, integer 100–5000. Larger cells conceal more precision but can exclude more radius matches. This changes IDs and the reachability index. |
| `geography.intersection_dataset` | `tirana-intersections-v1` | Installed road-junction dataset version; startup requires matching file version. Other versions require importing/installing reviewed data, not just changing a label. |
| `geography.location_max_accuracy_meters` | 50 | Client rejects larger reported device errors; positive integer, at most half the configured nominal cell size. Accuracy estimate stays on-device. |
| `geography.location_fix_max_age_seconds` | 60 | Client rejects stale fixes and expires an unused willingness fix before submission; allowed 5–120 seconds. No manual fallback. |

The default is now 100 m cells / 50 m reported device error (decision 0006).
With the previous 1,000 m cells the whole-cell reachability rule excludes every 0.1/0.5 km
choice: the cell's uncertainty already exceeds that radius. Setting the cell size
to 100 m also requires `location_max_accuracy_meters` to be at most 50. Finer cells
reveal a more precise area, increase index cost, and still exclude some short-radius
matches. Radius and cell size are separate controls. See decision 0005 and the
population report; do not interpret device accuracy as proof of presence.

Changing age/accuracy limits does not prevent spoofed device/API claims. Production
UI always obtains location from the device. Synthetic API actors bypass that UI;
mocked-browser tests separately exercise the actual client quality checks.

## Arrival confirmation (`arrivals`)

| Parameter | Normal / small regression setting | Effect and choices |
| --- | --- | --- |
| `confirmation_count` | 20 / 2 | Accepted fresh arrival credentials required for JEMI KËTU; normal 10–500, simulation 1–500. Independent of willingness activation count. |
| `confirmation_stability_seconds` | 10 / 10 | Same valid arrival cohort must survive this interval; positive and strictly shorter than freshness and late-join remaining time. |
| `freshness_minutes` | 15 / 15 | Maximum life of an arrival claim, further capped by session/gathering end; 1–15 minutes. No automatic claim renewal. |
| `nonce_seconds` | 120 / 120 | Maximum arrival challenge lifetime; 1–120 seconds and strictly less than freshness. Replays cannot increase/refresh attendance. |
| `allowed_cell_neighbor_rings` | 0 / 0 | `0`: claimed cell must equal destination cell. `1`: also allow the eight surrounding cells (within the service grid). Expands accepted approximate area, not a GPS accuracy proof. |

## Limits that affect observed matching

| Parameter under `limits` | Default | Effect / bounds |
| --- | --- | --- |
| `new_signals_per_network_window` | 60 | Creation request budget per network window, including retries; positive, no greater than request budget. Population runs declare multiple real loopback peers; a one-group run shares this budget. |
| `requests_per_network_window` | 3000 | Total authenticated request budget per network window; positive integer ≤1,000,000. |
| `network_window_seconds` | 60 | Abuse-counter window, 1–600 real seconds. Fake clock advancement does not advance it. |
| `global_writes_per_second` | 500 | Global write request budget; positive ≤1,000,000. Not a measured sustainable throughput. |
| `max_active_signals` | 150000 | Store admission cap and worker snapshot bound; positive ≤1,000,000 and ≥ activation count. Expired-index cleanup can affect admission. Not proof of capacity. |
| `max_declines_per_signal` | 32 | Maximum retained per-session declined invitations; 1–128. Once exhausted, further automatic offers stop. |
| `arrival_requests_per_signal_window` | 20 | Additional per-credential arrival request budget; 1–60 per configured real-time window. |
| `max_body_bytes` | 1024 | Maximum request body size, 1–4096 bytes. |
| `cleanup_batch_size` | 1000 | Maximum entries in an expiry cleanup batch and page size used by worker reads; 1–1000. |

## Other event counts and visibility

`notifications.foreground_poll_seconds` (30, allowed 1–60) is active: controls how
often the visible client checks own status, with jitter/backoff. It affects when a
person sees JEMI GATI/JEMI KËTU, not when the backend forms them.

Public snapshot settings below are active. Notification delivery/follow settings
and the daily-summary retention remain validated plans, not implemented behavior:

| Setting | Normal default | Intended meaning / current validation |
| --- | --- | --- |
| `public_activity.area_size_meters` | 1000 | Shared public willingness/gathering grid, 1000–5000 m, integer multiple of private cells. |
| `public_activity.capture_max_seconds` | 30 | Maximum capture interval, ≤30 s and below release interval; incomplete captures discarded. |
| `public_activity.max_snapshot_bytes` | 1000000 | Complete uncompressed release cap, 1024–1,000,000 bytes. |
| `public_activity.minimum_count` | 20 | Minimum to publish area activity; normal ≥20. |
| `public_activity.count_buckets` | `[20,50,100,250,500,1000]` | Ascending published lower bounds; first equals public minimum; 1–32 entries. |
| `public_activity.release_seconds` / `delay_epochs` | 300 / 1 | Fixed release interval and delayed epochs; normal interval ≥300, delay positive. |
| `public_activity.snapshot_retention_minutes` | 15 | Origin snapshot lifetime, ≤15; must exceed `release_seconds*(delay_epochs+1)` in seconds. |
| `public_activity.daily_summary_retention_days` | 30 | Planned protected summary lifetime, ≤30 days; summaries not implemented. |
| `notifications.nearby_gati_count` / `nearby_arrival_count` | 50 / 50 | Large-nearby alert thresholds, both must be public bucket boundaries. |
| `notifications.nearby_radius_km` | 5 | Proposed alert distance, 1–20 km; not participants' travel-radius setting. |
| `notifications.push_min_interval_seconds` | 300 | Proposed minimum push gap, at least queue TTL. |
| `notifications.push_max_per_hour` | 6 | Proposed maximum pushes per handle/hour, 1–60. |
| `notifications.queue_ttl_seconds` | 300 | Proposed queued update lifetime, 1–300 seconds. |
| `notifications.area_follow_max_hours` | 24 | Proposed temporary follow lifetime, 1–24 hours. |

All integer fields are positive and ≤1,000,000 unless tighter bounds or the explicit
zero-neighbor exception above apply. Simulation lowers some publication/activation
floors; isolated clock steps also drive public snapshots. It does not enable unfinished notifications. Avoid interpreting the nearby
50 threshold as today's activation threshold: the implemented JEMI GATI threshold
is `matching.activation_count` (30 normal/population, 3 small regression profile).

## Optional Web Push configuration (schema 8)

Delivery integration is in progress; the following typed settings are not yet a
claim of working provider delivery. `push_enabled` defaults to false. `push_contact`
is the operator's public HTTPS/mailto contact, never a participant email; replace
the localhost development placeholder before external use. `push_endpoint_hosts`
is an exact DNS allowlist (no wildcards, URLs or IP literals). Default providers are
Chrome/Firefox/Safari; network address validation is an additional transport gate.
`push_worker_batch_size` (1000, max 1000) bounds due records per pass;
`push_worker_seconds` (2, max 30) schedules passes; `push_concurrency` (8, max 16)
bounds simultaneous workers; `push_timeout_seconds` (5, max 10) bounds delivery;
`push_max_attempts` (3, max 5) and `push_retry_seconds` (10, max 60) bound exponential
retry attempts within `queue_ttl_seconds` (300 maximum). Existing
`push_min_interval_seconds` and `push_max_per_hour` jointly set the minimum gap
between newly queued notifications; opting out and back in cannot reset it.

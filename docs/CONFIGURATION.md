# Functional configuration

`config/gati.yaml` is the complete production-default document. `internal/config`
is its authoritative typed schema and validation policy (schema version 6).
Every field is required, including explicit zero/false values. Unknown or duplicate
keys, aliases, nulls, coercion from strings to numbers, extra documents and files
above 64 KiB are rejected. No field accepts credentials or external URLs.

`config/simulation.yaml` contains low-count local test defaults. Normal builds
reject it. Only explicitly compiled `-tags simulation` binaries can load it;
HTTP serving with a simulation build is restricted to literal loopback addresses.
Only the isolated simulation build includes authenticated clock controls.

Willingness, matching, admission, arrival and device-fix quality settings are active.
Aggregate and notification-subscription settings await those milestones; fields do
not establish that a feature or protection exists.

The public representation uses deterministic JSON field order and a SHA-256 hash
of those bytes. YAML comments/formatting do not change the hash. All consumers can
inspect the effective settings and schema version. Changes require restart. Gatherings pin their original configuration hash and
destination; pending reservations are revalidated.

Production floors: activation/arrival counts ≥10, public minimum ≥20, publication
interval ≥300 seconds and delay ≥1 epoch. Availability is 30..120 minutes. Retention
caps follow the plan; deadlines must leave enough time for matching stability and
delayed publication. Public alerts accept only released bucket boundaries. The
schema also bounds geography, lists and numeric values to reject overflow or
unbounded configuration. These are explicit review guardrails, not anonymity proofs.

Simulation can lower count/publication floors; all lifetime and shape validation
still applies. Changing a threshold does not change other independent thresholds.
Changing schema guardrails requires a visible code/doc review rather than an env
variable bypass. Geographic settings are checked against the imported intersection dataset at startup.

Schema 2 adds a required `limits` section: body size, network window/request/create
counts, global writes/second, active-signal capacity and cleanup batch size. Bounds
are enforced before serving; every limit remains public and auditable. The defaults
are starting controls, not evidence of effective Sybil resistance at production load.

The current bounded activation transaction accepts `matching.activation_count` up
to 500 and no greater than `limits.max_active_signals`. This is an implementation
work bound; larger crowds can use late admission without a larger founding write.

Schema 3 adds `matching.invitation_cooldown_seconds` (default 60; production floor
30; maximum 300) and `limits.max_declines_per_signal` (default 32, maximum 128).
These bound repeat invitations and temporary declined-gathering links. Once the
link budget is reached, automatic invitations stop for that willingness session.

Schema 4 adds `limits.arrival_requests_per_signal_window` (default 20, maximum 60),
using the configured network-window duration, in addition to network/global limits.
Arrival confirmation cohorts support thresholds up to 500. Nonce/freshness, coarse
neighbor rings and stability now actively control the arrival flow.

Schema 5 replaces `crossing_dataset`/`crossing_index_batch_size` with
`intersection_dataset`/`intersection_index_batch_size` and the destination rule with
`nearest_eligible_crossroad_to_coarse_group_center`. The derived map is
`tirana-intersections-v1`. Old crosswalk configuration is rejected, not aliased.
Private gathering JSON now calls its destination `intersection`.

`geography.location_max_accuracy_meters` defaults to 100 (positive, at most half
nominal cell size). `location_fix_max_age_seconds` defaults to 60 (5..120). Both
willingness and arrival ask the device for a fresh high-accuracy one-shot fix;
reported accuracy and age are checked locally. A willingness fix expires before
submission if the user waits too long. Denial, missing service, timeout, out-of-area
or poor/stale fixes leave participation disabled, with no manual fallback.
Neither coordinates nor reported accuracy are sent to GATI. These settings cannot
attest hardware, prevent spoofing or make the coarse server claim trustworthy.

For field-by-field defaults, accepted choices and current runtime effects, see
[MATCHING_PARAMETERS.md](MATCHING_PARAMETERS.md). In particular,
`matching.intersection_index_batch_size` currently has no matcher consumer.

Schema 6 accepts travel radii as finite numeric kilometres in 0.1 km increments
from 0.1 to 20, and cells from 100 to 5,000 m. Existing whole-kilometre configurations
remain valid after the schema version update. Store, API, matching and client
restoration preserve fractional values. The current requested gati.yaml settings
are 30 willing credentials, 20 arrivals and radii [0.1,0.5,1,3], with 1,000 m cells.
A short radius does not override the whole-cell conservative distance test.
Small cells increase location exposure; see decision 0005. Test fixtures no longer
read the developer-editable gati.yaml as immutable regression defaults.

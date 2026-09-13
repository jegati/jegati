# 0006 — Default to 100 m cells

Date: 2026-09-13. User-directed, following the recorded population comparison.

Set `geography.cell_size_meters` to 100 in the operating config and matching
population example. Set `location_max_accuracy_meters` to 50 to satisfy the
existing half-cell accuracy constraint. Keep the schema, travel choices, thresholds,
matching rule and deadlines unchanged. This supersedes the 1,000 m default retained
in decision 0005; historical experiment artifacts remain unchanged.

Smaller cells expose a more precise participant area and increase static index
memory/startup cost. Conservative reachability can still exclude 100 m journeys;
reported device accuracy remains an estimate, not proof of physical presence.

Restart the local development API with a fresh ephemeral store when applying the
grid change; do not reinterpret old cell IDs or gathering state. Fixed 1,000 m
regression fixtures remain explicit fixtures, separate from the operating default.
Validate configs and the actual browser flow against the new runtime settings.

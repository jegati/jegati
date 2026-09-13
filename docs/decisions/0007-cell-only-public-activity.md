# 0007 — Implement documented inference limits and cell-only public gatherings

Date: 2026-09-13. Status: user-directed, accepted for local implementation.

The user explicitly asked to document inferred information and implement the
features, place gatherings in the same cells on the statistics map rather than at
exact locations, and leave push notifications for the user to choose.

Use the planned fixed 1,000 m public grid for both willingness and gatherings;
retain 100 m private matching cells. Public API entries carry opaque gathering IDs,
public cell IDs, deadlines and permitted buckets/state. They never carry the
crossroad ID, name or coordinates. Render a shared cell polygon/list, not displaced
or exact destination pins. Keep the actionable destination in private invitation/
admission responses. Passive map viewing creates no willingness or admission.

Implement the documented minimum/bucket/delay policy despite known collusion,
differencing, nested-count and origin/destination inference. No claim of formal
anonymity or protection from inferred participation is made. This supersedes the
additional inference-blocking gate in ACTIVITY_IMPLEMENTATION_PLAN and broad
inference wording in REQUIREMENTS; it does not waive direct-data boundaries,
small-count suppression, retention, abuse controls, tests or audit documentation.

Push is off by default and requires explicit user choice. Foreground behavior is
fully available without permission. Optional resume credentials and provider
metadata have the bounded-access/physical-retention limitations described in the
activity plan; opt-out clears application delivery bindings and local resume data
when the client runs. Provider/OS copies are outside GATI's deletion guarantee.

Local implementation/tests/commits are authorized. No remote push or deployment is
included. Record exact implementation evidence in PROGRESS as each slice lands.

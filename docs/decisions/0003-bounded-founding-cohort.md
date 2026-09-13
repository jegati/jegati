# 0003 — Bounded founding cohort and stable reservations

Date: 2026-09-13. Status: implementation direction for milestone 07.

A dense compatible crowd must not require one unbounded transaction to activate.
Select exactly `activation_count` oldest eligible signals from the highest-ranked
compatible intersection candidate. Compute the nearest shared mapped intersection using
those founders' coarse centers. Other compatible people receive the same open
gathering through the late invitation/admission flow. They need not be counted in
the initial threshold. This preserves configurable activation and late joining.

Reserve that founding cohort while testing the configured stability interval.
Recheck all founders, remaining availability and the common intersection at activation.
A cancelled/expired/unavailable founder invalidates the reservation even if a new
signal has since replaced it. This conservatively delays some groups; it avoids
claiming uninterrupted eligibility based only on two snapshots. New willingness
must not reset a valid reservation or move its selected destination.

Reservations, private links and retry state are temporary Valkey records. No pending
counts, candidate identities or computed group center appear in participant/public
responses. A worker crash can postpone formation; recovery rechecks deadlines and
releases stale reservations. Actual activation keeps destination/deadline fixed.

The first milestone 07 slice is a pure planner with behavior tests. Reservation,
atomic transitions, worker scheduling and UI evidence are still required before
claiming continuous activation works. The snapshot/ranking heuristic can miss
alternative groups and must be measured in the simulation and scale milestones.

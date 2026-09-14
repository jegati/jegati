# 0014 — Bounded presence renewal and gathering expiry

Accepted 2026-09-14 after the lifecycle review. The approved priorities are expired
metadata cleanup, clear ended public presentation and explicit early renewal of
presence. Willingness extension, gathering extension and destination relocation
remain outside this change.

## Behavior

An already confirmed participant can press **JAM ENDE KËTU** during the final
`arrivals.renewal_window_seconds` of their current confirmation (default 120;
positive, at most 300, strictly shorter than freshness). Renewal requires a new
short-lived challenge and a fresh one-shot device fix, coarsened locally. It sets
presence expiry to the earliest of now plus configured freshness, the original
session expiry and the frozen gathering end. If no extension is possible, renewal
is unavailable. It never happens automatically. Existing admission cutoffs do not
prevent an already admitted participant from renewing before the gathering ends.

The challenge is bound to renewal mode and the current contribution, and cannot
outlive that contribution's previous deadline. Ordinary arrivals and renewals
cannot exchange challenges. Atomic confirmation updates the existing sorted-set
member's expiry; it neither adds another counted claim nor interrupts a valid
stability cohort. A used challenge only returns the existing confirmation on
retry. Retraction, cancellation, disappearance or expiry of the contribution
prevents renewal; a later ordinary arrival creates a new member and stability
interval. Existing global, network and per-session arrival limits cover both new
endpoints. A rejected location claim leaves the old confirmation until its expiry.

The public renderer labels ended cards using their already released deadline,
retains their historical buckets until snapshot expiry, and removes ended entries
from the gathering map layer. Another live gathering in the same cell keeps that
cell visible. No new queries or count releases are introduced.

Existing bounded cleanup prunes expired gathering index members as well as expired
signals. Existing matching reconciliation batches and own-session reads atomically
clear expired session/gathering associations and challenges, preserving newer
assignments and original session TTLs. No extra participant index or service is
introduced. Cleanup lag observes both expiry indexes.

## Privacy and verification

One current contribution reference can now persist across explicit renewals,
bounded by the original session and gathering. The session already links these
actions; no renewal/location history is stored and no new fields are public.
Compromise during that lifetime can still link activity. Device location remains
a claim, one credential is not one human, and thresholding is not formal anonymity.
Existing inference and administrator/host exposure limits remain.

Config schema 9 makes the new required setting visible and hashed. Old schema 8
files must be updated; no implicit environment default or silent migration exists.
Regression evidence is recorded in PROGRESS and the lifecycle implementation report.

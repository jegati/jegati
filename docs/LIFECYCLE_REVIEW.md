# Gathering and willingness lifecycle review — 2026-09-14

Historical review: the three approved follow-ups are now covered by
[decision 0014](decisions/0014-bounded-presence-renewal.md). See PROGRESS for current
implementation and test evidence. The findings below describe the pre-change code.

This review answers the user's lifecycle questions against the current source and
config. It does not authorize or implement a changed product lifecycle. Existing
regression evidence is in reports/auditability-followup-2026-09-14.md; no new runtime
tests were executed for this review.

## Implemented behavior

- Activation counts willingness, not promises to attend: 30 compatible signals
  stable for 10 seconds. Going is a separate explicit choice. Twenty fresh arrival
  claims stable for 10 seconds produce JEMI KËTU; credentials are not unique humans.
- Existing gatherings, including JEMI KËTU, accept eligible late participants.
  The current cutoff requires five minutes of remaining gathering/session time;
  the browser disables the public join button at that boundary. Public discovery
  depends on delayed, suppressed releases; willing users can get private offers.
- Activation freezes the deadline at the earlier of activation plus 60 minutes and
  the earliest founding signal expiry. Joining, going and arriving do not extend it.
  Destination geometry is also frozen. No user relocation endpoint exists.
- Willingness offers 30/60/90/120 minutes. No in-place extension endpoint or button
  exists. Identical creation retries preserve expiry; changed claims conflict.
  A person can explicitly end a session and start another, or start again after
  expiry, with fresh device location and a new random capability.
- Arrival freshness is the earliest of confirmation plus 15 minutes, session
  expiry and gathering end. The user can confirm again with a fresh location and
  challenge after freshness expires. While still confirmed, the JAM KËTU button
  is hidden and challenge issuance rejects a second active confirmation. There is
  no dedicated early renewal button. Retraction is supported.
- Falling below the fresh-arrival threshold resets JEMI KËTU to JEMI GATI; it does
  not close the gathering. Even an empty gathering retains its fixed lifetime.
  A still-live willingness may qualify for another gathering after this one ends;
  that is a new activation, not an extension or transfer of arrival confirmation.

Source: config/gati.yaml; internal/store/{gatherings,intent,arrivals,store}.go;
internal/httpapi/{invitations,arrivals}.go; internal/matching/planner.go;
web/src/{main,session-controller,activity}.ts.

## Disappearance: separate usability, publication and retention

Live admission and reads enforce deadlines. Gathering and destination-lock keys
have TTLs at the frozen end; arrival indexes cannot outlive the gathering. The
client clears its private invitation on expiry. Session records have their own
chosen expiry, so willingness can outlive an individual gathering.

Public releases are historical, not live occupancy: five-minute epochs, one extra
delayed epoch, and 15-minute retention from capture start. The current renderer
can retain an ended gathering in an unexpired snapshot with joining disabled. It
does not explicitly label the card as ended or remove its map cell on gathering
expiry alone. Later captures omit expired gatherings. External copies cannot be
revoked by server expiry.

Static review found two storage details requiring follow-up:

1. `activate` inserts gathering IDs/deadlines into `gati:gatherings` and extends
   the whole index's TTL to cover newer gatherings. `OpenGatherings` filters expired
   scores but does not remove them; the cleanup worker only prunes the willingness
   expiry index. No explicit expired-gathering-member pruning was found. Under
   continuous gathering creation, old IDs/deadlines can therefore accumulate in
   the shared index despite individual gathering records expiring. These entries
   contain no member list, but conflict with bounded metadata retention.
2. Expired `_gathering` / `_gathering_until` references can remain in a still-live
   session until reassignment or session expiry. HTTP responses ignore expired
   invitations, but that is not immediate removal of the stored association.

The existing short scenarios and logical-expiry tests do not establish cleanup of
these details during uninterrupted gathering churn. Add an integration test with
an expired gathering plus a still-live keeper, exercise the normal worker cleanup,
and assert pruning without touching the live gathering or its deadline. Also
check stale session associations and public ended-card behavior.

## Candidate additions, not implementation decisions

- Explicit short renewal of arrival freshness shortly before expiry can retain
  the current privacy boundaries if it requires a fresh one-shot device fix and
  nonce, atomically replaces rather than duplicates the current contribution,
  never extends session/gathering lifetime, stores no location history, and is
  rate limited. Review stability-cohort behavior during replacement.
- A clearer "start another willingness period" action can use a fresh capability
  after ending the old session; no permanent rollover link, double contribution
  or automatic transfer of arrival/push consent. Fresh tokens do not eliminate
  timing/coarse-location inference. Same-session extensions instead increase the
  association lifetime and require an explicit retention-policy decision; an
  absolute cap measured from original creation would be essential.
- Automatically extending gatherings merely because someone joins risks keeping
  a gathering alive through fake joins. Prefer existing finite activations or a
  separately designed continuation based on explicit fresh participation, with
  an absolute lifetime cap, bounded links and all expiration/cache/arrival
  invariants retested. The same ID living longer changes the frozen-deadline policy.
- Do not silently move an active destination: people may already be travelling
  or offline. A new automatically matched destination with explicit opt-in and
  fresh arrival checks is preferable to transferring commitments or arrivals.
- Label/hide ended public cards using their already released deadline; this needs
  no new participant query. Address index cleanup before longer-lived workflows.

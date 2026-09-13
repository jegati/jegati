# Selected usability improvements: implementation plan

Date: 2026-09-13. User selected suggestions 1, 2, 3 and 5 from the usability review:
optional background notifications, simpler JAM GATI, private destination preview
before commitment, and clearer waiting/recovery. Sharing links/QR (4) and new
purpose copy (6) are excluded. No native location attestation is added or claimed.
This plan refines the remaining activity plan; PROGRESS records actual completion.

## User journeys and boundaries

- **One action for willingness:** choose duration/radius, press JAM GATI, allow a
  fresh one-shot device fix, coarsen locally and submit. Do not acquire location on
  load. No manual location, automatic retry enrollment, or continuous tracking.
  Preserve the request/capability on uncertain submission; retries do not extend TTL.
- **Preview before commitment:** a map selection prepares a join; it never enrolls.
  An explicit preview action gets a fresh device fix and checks the published
  gathering, reachability and live cutoff. Show its private crossroad/map and time,
  then PO, PO SHKOJ / JO TANI. An existing session remains unchanged while previewing.
  A newcomer is enrolled only on confirmed atomic /api/join. Confirmation revalidates
  state/radius/cutoff; stale preview/location requires another explicit preview.
- **Waiting/recovery:** explain active willingness with suppressed/delayed public
  numbers; show offline/retry state and server-loss/expiry honestly. Distinguish
  waiting, going and fresh arrival; show arrival-confirmation expiry and offer an
  explicit fresh confirmation after expiry. No automatic extension or guilt.
- **Optional push:** off by default; an explicit action requests browser permission.
  Participating/matching works without it. Generic Albanian payload, no private
  coordinates/IDs/secrets in notification text or URLs. Open the first-party app,
  restore only an opted-in unexpired capability and fetch live state. Opt-out and
  cancellation remove delivery bindings. Provider retention is outside our control.

Private preview uses a bounded POST with an authorization-header random capability
and coarse cell/radius/duration/gathering ID. No preview record or participant link
is stored. Only a currently published gathering can be previewed; responses are
no-store and never enter public snapshots/service-worker caches. Its capability is
held only in page memory until an explicit confirmed join. Existing-session preview
uses that session's immutable coarse claim. Public maps remain 1 km cells; private
matching remains 100 m. Accepted inferred information stays documented (0007).

## Commit sequence and evidence

| Commit | Change | Acceptance evidence |
| --- | --- | --- |
| A | Record scope and private-preview decision | Requirements, plan and progress agree; excluded suggestions absent. |
| B | One-action willingness and waiting/recovery | Browser fresh/denied/stale/poor fix, keyboard, double-click, offline, lost response, cancellation and arrival-expiry checks. |
| C | Private preview API and confirmation UI | No preview enrollment or public destination; body/query/origin/rate limits; reachability, cutoff and stale preview; existing/new session confirm/cancel; real-store and browser checks. |
| D | Server-side offer discovery | Eligible background sessions are offered open gatherings without status polling, regardless of push preference. Bounded work, cooldown/decline/expiry and no changed counts from delivery. |
| E | Expiring notification registration/outbox | One binding per opted-in session, bounded indexes/queues/retries, dedup/coalescing and relevance revalidation; cancellation/expiry/restart tests using real Valkey. |
| F | Web Push transport | Reviewed pinned dependency, local VAPID keys, host allowlist and DNS/IP/redirect SSRF controls, bounded concurrency/time/body and failure cleanup; fake local transport tests, no real endpoints in fixtures/logs. |
| G | Opt-in browser delivery and resume | First-party service worker/manifest, explicit permission, generic notification, bounded IndexedDB capability resume, immediate logical opt-out and cleanup; denied/unsupported/closed/reopened/offline browser tests. |
| H | Evidence and local integration | Targeted suites, real store, synthetic flow; distinguish fake push from actual browser/provider/device interoperability. Do not claim release readiness or 100k capacity. |

Small changes may share a coherent commit. Do not expose a permission button that
claims working push until registration, delivery and cleanup are connected. The
push work can reuse the existing foreground poll without introducing chat/sockets.
Use existing session/gathering deadlines; typed configuration must expose any new
queue/work/transport/preview bounds. Secrets remain outside functional config.

No new credentials are needed for B–E or fake local delivery. Real push requires
local application keys and a valid contact/HTTPS origin configuration; never ask
for passwords in chat. Complete independently testable code before requesting an
external device/provider interoperability step. No remote push/deployment is authorized.
The broader area's follow notifications, daily summaries, production hardening,
100k benchmark and independently verifiable release remain tracked separately.

## Implementation status (2026-09-13)

A–G are implemented: one-action willingness, waiting/recovery, eligible private
preview, background offer discovery, bounded temporary subscription/outbox,
encrypted provider transport and explicit browser opt-in/resume. Step H local
checks passed (28 browser tests, real-store/race, native checks and fixed-clock
integration; see PROGRESS); actual provider/device delivery remains
unverified. Push is off by default while the public operator contact is unset.
Service keys are already generated locally and ignored by Git. The final
3,000-person regression and exact source/config provenance are in
[reports/usability-final.md](reports/usability-final.md).
Sharing links/QR and new purpose copy remain excluded. No stronger location proof
or public deployment readiness is claimed.

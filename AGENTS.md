# Working on GATI 🦩

These instructions apply to the whole repository. GATI is a small, anonymous,
temporary collective-activity tool. Keep it understandable and maintainable by
one developer. Product simplicity and privacy take precedence over extra features.

## Start and resume here

1. Read `docs/PROGRESS.md` for the current milestone, evidence and next action.
2. Read `docs/REQUIREMENTS.md` and the relevant sections of
   `docs/DEVELOPMENT_PLAN.md`. Read `docs/decisions/0001-mvp-boundaries.md` before
   changing matching, destination selection, admission or storage.
3. Inspect `git status --short`, recent commits and applicable nested instructions.
   Preserve existing work. Recheck recorded environment observations; they can age.
4. Run `bash scripts/doctor.sh` when toolchain readiness is relevant. It only
   inspects tools; it does not install them or certify runtime readiness.
5. Continue the next incomplete milestone. Commands described as planned are not
   evidence that a Make target, test, service or protection already exists.

The user's latest explicit instructions take precedence over these project files.
If documents conflict, reconcile them with those instructions; do not silently
implement an outdated plan. Record material decisions and keep the affected docs
consistent. `docs/REQUIREMENTS.md` captures agreed behavior; the plan supplies
proposed defaults and implementation order; `docs/PROGRESS.md` records reality.

## Autonomy and collaboration

- The user authorized implementation, installation of needed development/testing
  tools, running tests, fixes and incremental local commits. Continue through
  milestones without asking permission for each routine step.
- Prefer reversible, project-local or user-local setup. Pin and verify downloaded
  tools using publisher checksums/signatures where available. Record prerequisites
  and supported versions so another developer can reproduce the setup.
- Never ask the user to paste passwords or tokens into chat or tracked files. If
  credentials, interactive authentication or administrator access is necessary,
  explain the exact blocked operation. Complete independent work when practical.
- Use reasonable defaults within the agreed plan. Ask for a product decision only
  if a material behavior/privacy change cannot be resolved from existing direction.
- Current authorization includes local commits, but does not include remote
  pushes or public deployment. A question about whether pushing is possible was
  not permission to publish. Honor subsequent explicit authorization without
  asking repeatedly. Do not force-push or rewrite unrelated history.
- Do not spawn subagents unless the user explicitly authorizes delegation or
  another applicable instruction requires it. This file does not authorize it.
- Give brief progress updates during sustained work. Before stopping, record
  completed work, exact validation, limitations and the next action in PROGRESS.

## Product invariants

- Entire product UI is Albanian, including errors, accessible labels and
  notifications. Use the 🦩 identity and “Bëje vullnetin tënd të dukshëm.”
- Preserve: A JE GATI? → JAM GATI → JEMI GATI → PO, PO SHKOJ → JAM KËTU → JEMI KËTU.
- No accounts, identity fields, profiles, chat, feed, organizer roles, rankings,
  gamification, permanent membership or personal participation history.
- JAM GATI is willingness, not attendance. JO TANI declines that invitation and
  keeps willingness active. Cancellation must be immediate, neutral and easy.
- Availability starts at 30 minutes. Match continuously on changes/expiry and
  timer deadlines, without aligned appointment slots. Delayed public publication
  is a separate privacy mechanism, not a delay imposed on internal matching.
- Automatically choose the eligible mapped road intersection (crossroad) nearest the
  coarse group's center that all counted users can reach within their radii.
  Use imported map data, not a manually predefined meeting-area catalog. No shared
  mapped intersection means no invented destination. Describe pedestrian space beside
  the intersection; do not position a gathering in the roadway.
- Freeze an activated gathering's destination and deadline. Newly notified users
  can join before the admission cutoff, including after JEMI KËTU. Revalidate
  reachability, remaining availability and live gathering state on admission.
- Receiving a notification or viewing a map never counts as willingness or
  arrival. Founding and late participants follow the same arrival checks.
- Keep functionality easily configurable through typed `config/gati.yaml` and
  published nonsecret effective settings. Validate thresholds, timing, buckets,
  retention and profile rules. Do not hide behavior in environment-only overrides.
- Future availability, native apps, custom activities and route planning are
  outside the current MVP. Do not expand the scope merely because they are useful.

## Privacy and security invariants

- Require a fresh device-location fix for willingness and arrival; no manual
  participant-location selection. Reject stale/poor fixes using public settings.
  Device location is not authenticated proof of accuracy or presence.
- Convert one-shot device coordinates to coarse cells locally. Never transmit or
  persist exact participant coordinates. Public intersection geometry is map data.
- Use random expiring capabilities; store their hashes, transmit secrets only in
  authorization headers, and never introduce permanent participant/device IDs.
- Every participant record, link, index entry, nonce, queue item, subscription and
  client credential needs an explicit bounded lifecycle. Every read/transition
  checks server deadlines. Native key TTL does not expire set members for you.
- Keep live participant state in private Valkey with persistence disabled. No
  participant backups, disk snapshots, access logs, request-body logs, tracing
  identifiers or individual admin views. Enforce supported host/store controls
  and test actual settings; do not claim configuration comments enforce privacy.
- Public APIs expose fixed, delayed, suppressed/bucketed aggregate releases only.
  No individual markers, member lists, arbitrary count queries, exact counts or
  explicit complementary totals. Public gathering entries use the same aggregate
  cells as willingness, without intersection IDs/coordinates/labels. Private
  invitations/eligible preview/admission retain actionable destinations (0008). Decision 0007 accepts
  inferred information: document counterexamples without blocking implementation.
- Review all observable surfaces together: maps, private invitations, admission
  responses, arrival updates, history, area alerts, caches and configuration.
  Test differencing, colluding inputs and nearest-intersection inference.
- Never cache authenticated responses in shared caches. Public alerts use released
  buckets; they cannot consult suppressed raw counts. Push is optional, expiring,
  minimally revealing and separate from willingness. Use first-party map assets.
- Bound request sizes, work, admission, fanout and retries; use atomic transitions,
  replay protection, rate limits and idempotency. Do not address abuse with identity
  collection, fingerprinting or tracking vendors.
- Keep simulation data and endpoints separate. Production artifacts must lack
  fake-clock/test-control routes and reject simulation configuration. Synthetic
  population tools must refuse production targets by default.
- Do not overclaim: one credential is not one human; coarse location claims are
  not proof of presence; thresholding is not formal anonymity; TTL is not forensic
  erasure; reproducible artifacts do not prove a remote operator's honesty.

## Architecture and review discipline

- Follow the planned TypeScript/Vite PWA, MapLibre, one Go API/worker codebase,
  temporary Valkey and Compose/Caddy deployment. Avoid unnecessary services or
  hosted dependencies. Explain material architecture changes in a decision record.
- Inspect existing code before designing replacements. Prefer standard libraries
  and small explicit modules. Pin dependencies and retain licenses/attribution.
- Ship small coherent commits with relevant code, documentation and verification.
  Stage specific files, inspect the staged diff and use descriptive commit messages.
  Never commit credentials, real participant data, downloaded toolchains or caches.
- Tests should verify behavior and threats rather than mirror implementation.
  Cover configuration rejection, expiry, race/retry behavior, radius boundaries,
  continuous timing, automatic intersection selection, late admission, arrival replay,
  aggregate leakage and notification lifecycle as each feature lands.
- Use seeded synthetic Tirana scenarios and a controllable clock for functional
  tests. Use real time for load tests, report hardware and accepted/rejected traffic,
  and distinguish synthetic people from accepted credentials.
- Run focused tests and milestone acceptance checks. Use the race detector when
  concurrency is involved. Do not label skipped checks as passed or call 100k scale
  achieved before its benchmark passes. Investigate failures before committing.
- Keep source, schema, infrastructure, data-flow/threat documentation and release
  evidence auditable. Publish hashes/config versions and compare independent builds;
  distinguish outside artifact checks from privileged host inspection.

## Documentation to maintain

- `docs/PROGRESS.md`: implemented state, validation, blockers and next action.
- `docs/REQUIREMENTS.md`: agreed scope and product corrections.
- `docs/DEVELOPMENT_PLAN.md`: architecture, defaults, commit order and release gates.
- `docs/THREAT_MODEL.md`: actors, exposed data, controls, gaps and test evidence.
- `docs/decisions/`: decisions with rationale and consequences, not chat transcripts.
- `README.md`: working setup commands and honest status; no nonexistent features.

Documentation for contributors can be English. User-facing product content remains
Albanian. Keep these instructions concise as the codebase grows; place detailed
domain guidance near the implementation instead of duplicating the whole plan.

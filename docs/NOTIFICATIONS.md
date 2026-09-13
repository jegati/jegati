# Optional background notifications

GATI works without push. An active participant can explicitly choose **Aktivizo
njoftimet**; only that action asks the browser for permission and registers the
first-party worker. Foreground status continues without permission or provider
support. Area follows and alerts for nonparticipants are separate unfinished work.

## Local setup

The code is connected, but `config/gati.yaml` deliberately defaults to
`notifications.push_enabled: false`. Real provider/device interoperability has not
been tested. For a live test:

1. Set `notifications.push_contact` to the operator's real public project/security
   contact URL (or service contact mailbox), then set `push_enabled: true` in that
   same YAML. The `https://localhost` placeholder cannot enable production-profile
   delivery. No participant email or provider account is required by this setup.
2. Run `make config-check`, then `make dev-push`. This creates local service keys
   if absent and starts the usual Compose stack with the optional read-only key
   mount. Existing keys are retained. Do not commit or disclose `.runtime/vapid.json`.
3. Open `http://127.0.0.1:5173`, express willingness, and explicitly activate
   notifications. Localhost is the local testing origin; remote device access needs
   a trusted HTTPS origin. Keep the same hostname throughout the test.
4. On a supported device, close the app while eligible synthetic peers form an
   invitation. Check generic delivery, opening current state, JEMI KËTU updates,
   opt-out, revoked permission and expiry. Browser/OS background restrictions and
   installation requirements must be checked on each supported device. This is
   manual interoperability evidence, not supplied by the automated fixtures.

To turn transport off, restore `push_enabled: false` and recreate the API via the
usual Compose command. Participant-side **Çaktivizo njoftimet** removes local resume
material and requests server removal without cancelling willingness. A sent request
or notification already held by the provider/OS cannot be recalled. Keep the
production deployment/release gates in DEVELOPMENT_PLAN.md; this optional overlay
is a local setup, not a new public deployment process.

## Configuration and delivery behavior

[MATCHING_PARAMETERS.md](MATCHING_PARAMETERS.md) describes all notification knobs.
The default gap is `max(300, ceil(3600 / 6)) = 600` seconds between new notifications
per capability. Each queued update has at most three attempts within 300 seconds;
retry attempts are separately bounded and may produce duplicates after uncertain
provider replies. A stable changed state waits for the gap before starting that
queue lifetime. The worker reconciles latest state, so brief intermediate changes
can coalesce. Delivery is best-effort and cannot promise every transition.

The endpoint-host list is exact, with checked public DNS/IPs, verified TLS and no
redirect or proxy forwarding. Only delivery binding/deadline metadata is sent in
a padded encrypted payload. The app's worker checks its temporary local opt-in and
current authenticated server state before displaying generic Albanian text. Opening
always uses `/`; no capability, location or gathering ID goes into a notification
URL, and opening never enrolls or confirms arrival.

## Data lifecycle and verification

Normal participation uses sessionStorage. Explicit push opt-in adds one IndexedDB
record: capability, binding, opt-out revision, expiry and registration status. No
location, provider endpoint or participation history is copied into it. The browser
owns its native push subscription. The app removes its resume record on opt-out,
cancellation, observed server loss, revoked permission or expiry; a closed device
can retain expired bytes until next execution. Server deadlines still reject use.
The worker has no fetch handler and creates no response caches.

Valkey contains encrypted endpoint/key material, bounded latest-state/outbox fields
and a rate-gap deadline. An opt-out revision on the existing temporary signal rejects
a delayed earlier registration. Cancellation and expiry clean the scheduling index;
all state remains bounded by willingness. See [STORAGE.md](STORAGE.md) and
[decision 0009](decisions/0009-optional-push-lifecycle.md). Provider metadata retention,
API-host compromise and malicious operator behavior are not solved by encryption
at rest or open source.

`make test-store` exercises restricted Valkey plus fake-provider HTTP transport,
including encryption, replay/races, opt-out fences, DNS/redirect controls and expiry.
`make test-browser` uses synthetic location and provider-registration fixtures. Its
push suite uses full Chromium headless mode; CDP injects a push event into the actual
installed worker with the app page closed. IndexedDB, authenticated state fetches
and native notification display execute. This is not Google/Apple/Mozilla delivery
or a real OS notification-click test. No external provider is contacted by fixtures.

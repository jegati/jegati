# 0015 — Direct public alpha and device emulation

Accepted 2026-09-14. The first public release will be an **alpha available for
use**, without a separate volunteer pilot. Subsequent changes will incorporate
alpha feedback. This supersedes volunteer-stage sequencing in earlier plans; it
does not claim that real devices or the public host were tested.

## Validation and launch

Run Pixel, iPhone and iPad browser presets locally, refresh Firefox/WebKit
regressions, and refresh exact-release audit/reproduction/rollback evidence.
Presets emulate viewport, touch, device scale and user agent; WebKit on Linux is
not iOS Safari on an iPhone. GPS is synthetic. Permission prompts, radio changes,
OS process termination, battery management, installed-PWA behavior and external
push delivery cannot be certified by these checks. Emulated offline and resume
events verify client behavior only.

There is no required volunteer recruitment or separate user-testing stage. Actual
host/edge/TLS/cache/privacy controls, fresh exact-release audit, artifact comparison,
rollback and operational alerts remain pre-activation work. Independent review is
outstanding and must not be represented as completed. Real-device checks can use
operator devices when available and alpha issue reports after launch; explicitly
record untested behavior. Keep push disabled until real-provider/device evidence.
The realistic maximum-lifetime overnight scenario remains a separate testing gap;
it was not included in this device/two-refresh batch.

Prepare a visibly labeled alpha with concise known limitations and optional feedback
instructions before activation. Preserve matching/privacy thresholds; a small
population can legitimately produce no invitation or published count. Host capacity
must determine an honest operational cap; public availability is not a 100k or 1M
capacity promise. Hosting access is pending. This decision does not perform a remote
push, DNS change or public deployment.

## Learning from alpha use

Use existing bounded private health metrics: errors, resource pressure, worker lag,
release freshness and recovery. Configure a private operator alert destination
before unattended operation. No individual funnels, session replay, fingerprints,
capabilities, exact locations or participation histories.

Feedback is optional and does not gate participation. Prepare instructions for the
project-specific email and optional GitHub issues: browser/OS family, app version,
expected/observed behavior and steps reproduced without personal data. Do not
request real gathering coordinates, identity, tokens, raw network logs or screenshots
containing private invitations. Email/issue providers have their own identity and
retention boundaries; these are external contact channels, not anonymous in-app
reporting. Redact identifying details before creating a public issue, keep private
reports private and never commit original reports.

Triage reproducible failures; turn sanitized examples into synthetic regression
tests; implement small reviewed commits; rerun affected journeys and audit the exact
release before updating. Use compatible rollback for regressions. Pause the service
if a defect exposes private data or corrupts participation state. Do not silently
weaken geographic validation or publish smaller counts to make alpha activity appear.

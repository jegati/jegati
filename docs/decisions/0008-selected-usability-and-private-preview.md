# 0008 — One-action willingness, private preview and optional notifications

Date: 2026-09-13. Status: accepted user-selected implementation scope.

The user selected usability review suggestions 1, 2, 3 and 5 and excluded 4 and 6.
Implement optional background notifications, a JAM GATI action that obtains a fresh
one-shot fix before submitting, destination preview before commitment, and clearer
waiting/recovery. Do not add sharing links/QR or new purpose/activity copy.

A published cell-map entry may lead to a private, location-eligible destination
preview before PO, PO SHKOJ. This supersedes decision 0007's phrase that exact
crossroads are available only through invitation/admission: explicit eligible
preview may reveal the same public-map landmark without recording willingness or
going. Public statistics remain cell-only. Preview responses are no-store, their
coarse inputs are not stored/logged, and no preview membership/history is created.
Normal API limits bound probing; inferred information remains accepted, not solved.

Existing participants preview with their immutable active claim; new participants
supply device location and chosen duration/radius. The preview does not reserve a
place or bypass the later atomic live admission checks. Reject expired/cutoff or
unreachable targets with a neutral error. Keep the capability/preview in page memory
until confirmation, then preserve the existing bounded retry contract.

Push remains optional and separate from participation. Opted-in background resume
can store only a current capability/deadline/notification handle on the device;
server expiry is authoritative and byte deletion while closed is not guaranteed.
No push-provider account, installation or permission is required to express willingness.
Native location attestation, signed GPS claims, persistent identities and continuous
location collection remain excluded. Detailed sequence is in USABILITY_IMPLEMENTATION_PLAN.

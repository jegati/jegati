# Publication gate: colluding invitation inference

Date: 2026-09-13. **Status: inference failure retained as an explicit limitation;
user elected to retain nearest destinations (decision 0004).** All inputs in this report are synthetic. No real participant
or device location was used. This is a counterexample to the broad inference
protection requirement, not a failure of the documented exact-coordinate boundary.

The original evidence below used the crosswalk dataset. The user subsequently
corrected the landmark to a road intersection and required device-only location
input. The original report is retained as historical evidence. Local development
may continue under the bounded claim; this is not a privacy pass or deployment approval.

## Reproduction

Run `make privacy-probe`. It owns a disposable simulation store, uses the real
Tirana import and application API/planner/transactions, and exercises **production
thresholds and limits** with an isolated controllable clock. The command deliberately
exits 2 after reproducing the failure. Its underlying test passing means the known
counterexample reproduced, not that privacy passed. Evidence is written to
`reports/local/privacy-probe.json`; a reviewed synthetic copy is committed beside
this report as [09-inference-counterexample.json](09-inference-counterexample.json).

1. One synthetic observer submits 19 credentials from one network, within the
   production request limits. The observer knows its own coarse inputs.
2. No gathering invitation exists with these 19 signals.
3. Add one synthetic target signal and complete the ten-second stability interval.
4. The observer's ordinary authenticated own-session response now includes JEMI
   GATI and the automatically selected public crossing.
5. Enumerate the target's possible coarse cells and all allowed travel radii using
   the open selection algorithm and map. The crossing narrows **127 compatible
   cells to one**, `tirana-v1:1000:8:2`, which matches the synthetic target's input.

The observed crossing was `node/10199286492`. The report records map version and
effective configuration SHA-256. No target capability, exact GPS, civil identity,
member list, database read or public aggregate endpoint was observed by the attack.

The counterexample assumes exactly one additional eligible signal in this isolated
candidate neighborhood. It does not prove arbitrary people can always be identified
in a busy city. It does prove that the current architecture cannot claim protection
against this colluding-input location-inference case. Knowledge of a person's
identity would require additional outside information; the experiment reconstructs
a private coarse cell and participation, not a name or exact device coordinates.

## Why the current requirements need a decision

REQUIREMENTS.md requires protection against location inference and reconstruction;
AGENTS.md explicitly requires reviewing private invitations, colluding inputs and
nearest-crossing inference together. DEVELOPMENT_PLAN.md section 5 says to coarsen
or withhold affected surfaces if review finds the release rules inadequate.

The user also explicitly requested the closest shared mapped crossing selected
automatically. The implemented exact rule makes its output dependent on the unknown
signal's cell. Removing individual markers and hiding counts does not prevent this
case. Raising thresholds or adding rate limits raises the observer's cost but does
not establish one credential per human. Delaying/coarsening a public map alone
cannot repair a leak already present in the private invitation flow.

## Concrete choices for the product owner

**Retain the current exact rule for a bounded-anonymity MVP.** Keep no accounts,
client-side coarsening, short-lived state, anonymous capabilities and the implemented
anti-replay controls. Explicitly accept that colluding credentials can infer a lone
unknown participant's coarse area in cases like this. Continue public-map work with
its own conservative release safeguards. This would change the strength of the
privacy requirement; it must not be described as satisfying the absolute claim.
Public deployment would still need later abuse, load, host and independent audits.

**Require stronger protection and revise the product rules first.** Revisit how
activation and destinations depend on individual inputs, and which information
participants receive. Broader or randomized destination selection is a candidate
for analysis, not a demonstrated fix; it would change the exact nearest-crosswalk
requirement and still needs a collusion/contribution model. No identity collection
is proposed, and no homemade noise mechanism is being presented as a privacy proof.

## Decision and crossroad recheck

The user retained nearest destinations after the decision request, correcting
crosswalk to crossroad and removing manual location input. See
[decision 0004](../decisions/0004-crossroads-and-device-location.md). The selected
rule remains deterministic; the absolute inference requirement remains unmet.

`make privacy-probe` was rerun on `tirana-intersections-v1`, schema 5. It again
narrows 127 compatible cells to one, now `tirana-v1:1000:7:3`, from the invitation
landmark `node/10124813384`. It used 19 controlled credentials plus one synthetic
target under production limits and deliberately exited 2. The new independent
[synthetic evidence](09-crossroad-inference-counterexample.json) retains the new
config hash; the original crosswalk JSON was not overwritten.

Device-only input in the shipped UI does not authenticate a direct API caller's
coarse cell or prevent a modified client. No real geolocation was used in either
probe. Keep this known-failure diagnostic distinct from passing lifecycle tests.

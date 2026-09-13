# Isolated Tirana simulation

Run `make simulate` or `make simulate SCENARIO=tirana-sybil SEED=42` after bootstrap,
dependencies and Docker are available. Every invocation creates a fresh disposable
Valkey, starts a simulation-only Go binary on a free loopback port, drives the real
willingness API, writes a synthetic report, verifies functional-clock expiry and
removes its API/store. It does not touch the normal Compose stack.

Open `reports/local/simulation/index.html` in a browser for the Albanian synthetic
map/dashboard. Its road backdrop comes from the archived OSM extract. JSON beside
it records effective configuration hash, seed, synthetic people, generated/accepted
credentials, cancellations, successful retries and expired remaining signals.
Dashboard cell counts are cumulative accepted synthetic signals, not current
occupancy. No capability, real person or real device position is written to reports.

Edit `simulation/scenarios/*.yaml`: seed, population (1–1000), clustered/uniform/
sparse/single_hotspot distribution, offered availability/radii, cancellation,
duplicate-request fraction and Sybil fraction/credentials (up to 10 per person).
The simulation consumes thresholds and geography from `config/simulation.yaml`.
Unsupported scenario options are rejected; going/arrival/notification options will
be added with their functionality rather than silently ignored now.

## Isolation and time

- The CLI accepts only literal loopback HTTP origins with explicit ports, rejects
  redirects, verifies the simulation profile and authenticates the clock before
  writing willingness. No override allows a production target.
- `-tags simulation` is required for the simulation clock route and control-file
  flag. Production binaries lack them and reject simulation configuration.
- Simulation uses separate `.runtime/simulation/` service/control secrets and
  `gati-sim:*` ACL keys. Normal service credentials cannot initialize this namespace;
  simulation credentials cannot read/write `gati:*` records.
- The clock starts at a fixed epoch and is frozen until an authenticated loopback
  advance. Advances are forward-only, bounded per request and in total. The control
  rejects Origin-bearing browser requests and is absent from the product client.
- Functional deadlines use that clock. Native store TTLs and abuse windows use real
  time, retaining an independent two-hour maximum for this short-lived fixture.
  Advancing time verifies logical expiry/index cleanup, not accelerated Valkey
  forensic erasure. Cancel tombstones/native expired records can remain until their
  real TTL; the whole disposable store is removed at run completion.
- Native TTL behavior is tested separately by `make test-store`. Load benchmarks
  must use real time; these fast functional simulations are not scale evidence.

Current controls demonstrate bounded bursts, not privacy-preserving human identity
or robust Sybil resistance. A distributed attacker can exceed one network's budget.
Activation, admission, arrival and notifications are not simulated yet because
those application features have not landed.

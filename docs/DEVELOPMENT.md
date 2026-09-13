# Local development

## Pinned user-local toolchain

On Linux x86_64, from the repository root:

```sh
bash scripts/bootstrap.sh
source scripts/env.sh
bash scripts/doctor.sh
```

Bootstrap installs Go 1.27.1, Node 24.21.0 (npm 11.19.0) and the standalone Docker
Compose 5.5.1 client in `~/.local/share/gati/tools`. It does not modify shell startup
files or invoke sudo. Source `scripts/env.sh` in each terminal; Make targets will
also select these tools. Set `GATI_TOOLS_DIR` if a different install directory is
needed. Installation is idempotent; an existing executable is reused, so an
already-compromised local installation is outside its verification guarantee.

The checked-in `scripts/toolchain/versions.env` pins SHA-256 digests. Downloads are
verified before extraction/execution. Hashes were obtained from Go's release JSON,
Node's SHASUMS256.txt and the official Docker Compose GitHub release asset metadata.
This checks published artifact integrity; it is not an independent reproducible
build of these third-party toolchains. No downloaded binaries are tracked in Git.
Other platforms need equivalent supported tools installed manually.

## Container runtime prerequisite

Compose is a client; it does not install Docker Engine. On this Ubuntu PC, Docker
Engine is missing and administrator commands require interactive authentication.
To install the Ubuntu-packaged runtime and Compose plugin, run in your own terminal:

```sh
sudo apt-get update
sudo apt-get install docker.io docker-compose-v2
sudo systemctl enable --now docker
sudo docker info
```

Never paste your sudo password into chat. The system Compose plugin and the pinned
standalone client may have different versions; GATI will use the pinned client
when `scripts/env.sh` is sourced. Installing Docker can change host networking;
these instructions deliberately do not add your user to the privileged docker group.
For autonomous runtime testing, the agent's user also needs Docker socket access.
If you choose the standard Docker-group setup, run:

```sh
sudo usermod -aG docker "$USER"
```

Then log out/in and restart the Codex terminal so its process inherits the group;
`docker info` must succeed without sudo. Docker-group membership grants root-level
host access. Keep the socket restricted to that group; never make it world-writable.
If you prefer rootless Docker, use that setup instead of the group step.

Rootless Docker is another supported Engine option, but it requires `newuidmap`,
`newgidmap` and subordinate UID/GID ranges. Those helpers were absent during setup;
installing them also needs administrator access. Do not disable host protections or
make the Docker socket world-writable to get around that prerequisite.

## Current validation limits

Read [PROGRESS.md](PROGRESS.md) for implemented commands and checks. A successful
Compose configuration parse does not demonstrate a running stack, image build,
Valkey behavior or production readiness. Keep container startup validation pending
until a usable Docker Engine is available.

## Available scaffold commands

```sh
make deps                      # Download locked Go/npm dependencies
make config-check              # Validate production defaults
make config-show               # Print canonical public JSON
make config-check-simulation    # Check isolated simulation defaults
make verify-local              # Tests, builds, config/Compose parse and HTTP smoke
make dev-native                # API + Vite scaffold on loopback; Ctrl-C stops both
```

`make dev-native` opens the scaffold at http://127.0.0.1:5173, with an API on
127.0.0.1:8080. It has no participant data, willingness action or map yet and needs
no Valkey. This temporary scaffold command does not replace the planned Compose
workflow once participant storage is added. Node dependencies must first be installed
with `make deps`. The native smoke test uses temporary OS-selected loopback ports.

After Docker is usable:

```sh
make dev             # Build/start the development API, Vite and isolated Valkey
make down            # Stop the Compose stack
```

The Compose stack is only a development definition. The Valkey service has no
published port or persistent volume, and the API does not yet connect to it.
Restricted service ACLs/credentials must be added before participant writes.
The development frontend allows Vite's local asset cache; it is not the production
static frontend server. Caddy/production deployment belongs to milestone 13.

CI is configured to build/start the containers and check API/client/proxy/Valkey.
It has not run remotely because no push is authorized. Image references are pinned
to publisher index digests; schema parsing alone does not verify those image builds.

After starting the stack, `make check-containers` checks actual service health,
configuration proxying and Valkey persistence/network/mount settings. These checks
passed locally on 2026-09-13; remote CI and clean-host deployment remain separate.

## Willingness storage tests

`make secrets` creates local ignored service credentials without printing them;
`make dev` performs this prerequisite automatically. `make test-store` starts a
disposable synthetic-only Valkey with a random loopback port and removes it after
the integration suite. It never runs the suite against your application database.
The running Compose API now connects to private Valkey. `make dev-native` still
starts a read-only shell with participant writes unavailable unless an explicit
store address/password-file is supplied. Use Compose for the actual willingness API.

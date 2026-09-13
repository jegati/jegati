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
A later runtime check determines whether commands require `sudo` for this host.

Rootless Docker is another supported Engine option, but it requires `newuidmap`,
`newgidmap` and subordinate UID/GID ranges. Those helpers were absent during setup;
installing them also needs administrator access. Do not disable host protections or
make the Docker socket world-writable to get around that prerequisite.

## Current validation limits

Read [PROGRESS.md](PROGRESS.md) for implemented commands and checks. A successful
Compose configuration parse does not demonstrate a running stack, image build,
Valkey behavior or production readiness. Keep container startup validation pending
until a usable Docker Engine is available.

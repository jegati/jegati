#!/usr/bin/env bash
# Install pinned tools under ~/.local/share/gati/tools; never invokes sudo.
set -euo pipefail
repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
source "$repo_dir/scripts/toolchain/versions.env"
if [[ $(uname -s) != Linux || $(uname -m) != x86_64 ]]; then
  printf 'Automatic bootstrap supports Linux x86_64; see docs/DEVELOPMENT.md.\n' >&2
  exit 1
fi
gati_tools="${GATI_TOOLS_DIR:-$HOME/.local/share/gati/tools}"
mkdir -p "$gati_tools"
staging=$(mktemp -d "$gati_tools/.bootstrap.XXXXXXXX")
trap 'rm -rf -- "$staging"' EXIT
fetch() {
  curl --fail --location --silent --show-error --proto '=https' --tlsv1.2 \
    --retry 2 --connect-timeout 15 --max-time 300 "$1" -o "$staging/$2"
  printf '%s  %s\n' "$3" "$staging/$2" | sha256sum --check --status
}
if [[ ! -x "$gati_tools/go-$GATI_GO_VERSION/bin/go" ]]; then
  fetch "https://go.dev/dl/go$GATI_GO_VERSION.linux-amd64.tar.gz" go.tar.gz "$GATI_GO_SHA256"
  tar -xzf "$staging/go.tar.gz" -C "$staging"
  mv "$staging/go" "$gati_tools/go-$GATI_GO_VERSION"
fi
if [[ ! -x "$gati_tools/node-$GATI_NODE_VERSION/bin/node" ]]; then
  fetch "https://nodejs.org/dist/v$GATI_NODE_VERSION/node-v$GATI_NODE_VERSION-linux-x64.tar.xz" node.tar.xz "$GATI_NODE_SHA256"
  tar -xJf "$staging/node.tar.xz" -C "$staging"
  mv "$staging/node-v$GATI_NODE_VERSION-linux-x64" "$gati_tools/node-$GATI_NODE_VERSION"
fi
if [[ ! -x "$gati_tools/compose-$GATI_COMPOSE_VERSION/docker-compose" ]]; then
  fetch "https://github.com/docker/compose/releases/download/v$GATI_COMPOSE_VERSION/docker-compose-linux-x86_64" docker-compose "$GATI_COMPOSE_SHA256"
  mkdir -p "$gati_tools/compose-$GATI_COMPOSE_VERSION"
  install -m 755 "$staging/docker-compose" "$gati_tools/compose-$GATI_COMPOSE_VERSION/docker-compose"
fi
if [[ ! -x "$gati_tools/buildx-$GATI_BUILDX_VERSION/docker-buildx" ]]; then
  fetch "https://github.com/docker/buildx/releases/download/v$GATI_BUILDX_VERSION/buildx-v$GATI_BUILDX_VERSION.linux-amd64" docker-buildx "$GATI_BUILDX_SHA256"
  mkdir -p "$gati_tools/buildx-$GATI_BUILDX_VERSION"
  install -m 755 "$staging/docker-buildx" "$gati_tools/buildx-$GATI_BUILDX_VERSION/docker-buildx"
fi
# Docker discovers CLI plugins here. Preserve any separately managed installation.
gati_plugin_dir="${DOCKER_CONFIG:-$HOME/.docker}/cli-plugins"
mkdir -p "$gati_plugin_dir"
if [[ ! -e "$gati_plugin_dir/docker-buildx" && ! -L "$gati_plugin_dir/docker-buildx" ]]; then
  ln -s "$gati_tools/buildx-$GATI_BUILDX_VERSION/docker-buildx" "$gati_plugin_dir/docker-buildx"
fi
printf 'Installed pinned tools in %s\nUse: source scripts/env.sh\nDocker Engine must be installed separately.\n' "$gati_tools"

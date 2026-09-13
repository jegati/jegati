# Source from Bash to use GATI's pinned user-local toolchain.
gati_repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
source "$gati_repo_dir/scripts/toolchain/versions.env"
gati_tools_dir="${GATI_TOOLS_DIR:-$HOME/.local/share/gati/tools}"
export PATH="$gati_tools_dir/go-$GATI_GO_VERSION/bin:$gati_tools_dir/node-$GATI_NODE_VERSION/bin:$gati_tools_dir/compose-$GATI_COMPOSE_VERSION:$PATH"
unset gati_repo_dir gati_tools_dir

#!/usr/bin/env bash
# Black-box HTTP and production-build configuration checks; no container daemon.
set -euo pipefail
repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_dir"
source scripts/env.sh
if bin/gati -mode config-check -config config/simulation.yaml >/dev/null 2>&1; then
  printf 'Production binary accepted simulation configuration.\n' >&2
  exit 1
fi
node scripts/smoke.mjs

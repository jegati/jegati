#!/usr/bin/env bash
# Owned API/Valkey lab plus native Vite; synthetic local state is removed on exit.
set -euo pipefail
repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_dir"
source scripts/env.sh
exec python3 scripts/lab.py --output "reports/local/dev-native-$(date -u +%Y%m%dT%H%M%S)" -- node web/node_modules/vite/bin/vite.js web --host 127.0.0.1

#!/usr/bin/env bash
# Temporary scaffold-only workflow; the future stateful app needs Valkey.
set -euo pipefail
repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_dir"
source scripts/env.sh
mkdir -p bin
go build -o bin/gati ./cmd/gati
bin/gati -config config/gati.yaml &
api_pid=$!
# Run Vite directly so cleanup owns the actual process, not an npm wrapper.
node web/node_modules/vite/bin/vite.js web --host 127.0.0.1 &
web_pid=$!
cleanup() {
  kill "$api_pid" "$web_pid" 2>/dev/null || true
  wait "$api_pid" "$web_pid" 2>/dev/null || true
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
printf 'Scaffold: http://127.0.0.1:5173 (Ctrl-C stops both processes).\n'
wait -n "$api_pid" "$web_pid"

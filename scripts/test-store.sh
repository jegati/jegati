#!/usr/bin/env bash
# Disposable synthetic-only Valkey instance. Never connects tests to the app store.
set -euo pipefail
repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_dir"
source scripts/env.sh
source deploy/images.env
container=$(docker run -d --rm --read-only --user 999:999 --cap-drop ALL \
  --security-opt no-new-privileges --ulimit core=0 --memory 256m --memory-swap 256m \
  --tmpfs /data:size=128m -p 127.0.0.1::6379 \
  -v "$repo_dir/deploy/valkey.dev.conf:/etc/valkey/valkey.conf:ro" \
  -v "$repo_dir/.runtime/users.acl:/run/secrets/users.acl:ro" \
  -v "$repo_dir/.runtime/health-password:/run/secrets/health-password:ro" \
  "$GATI_VALKEY_IMAGE" valkey-server /etc/valkey/valkey.conf)
trap 'docker rm -f "$container" >/dev/null 2>&1 || true' EXIT
export GATI_TEST_ADDR
GATI_TEST_ADDR=$(docker port "$container" 6379/tcp)
export GATI_TEST_PASSWORD_FILE="$repo_dir/.runtime/app-password"
export GATI_INTEGRATION=1
# Go integration tests retry the initial connection with a short bounded deadline.
go test -race -p 1 -count=1 ./internal/store ./internal/httpapi

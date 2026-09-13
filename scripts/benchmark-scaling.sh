#!/usr/bin/env bash
# Never accepts a target address: every run creates and destroys its own store.
set -euo pipefail
repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_dir"
source scripts/env.sh
source deploy/images.env
output=${OUTPUT:?OUTPUT is required}
mkdir -p "$output"
[[ ! -e "$output/store.txt" && ! -e "$output/geography.txt" ]] || { echo 'Choose a new evidence directory.'; exit 1; }
python3 scripts/init-secrets.py >/dev/null
temporary=$(mktemp -d)
container=
cleanup() {
 if [[ -n "$container" ]]; then docker rm -f "$container" >/dev/null 2>&1 || true; fi
 rm -rf -- "$temporary"
}
trap cleanup EXIT
sed 's/maxmemory 128mb/maxmemory 1536mb/' deploy/valkey.dev.conf > "$temporary/valkey.conf"
chmod 644 "$temporary/valkey.conf"
container=$(docker run -d --rm --read-only --user 999:999 --cap-drop ALL \
 --security-opt no-new-privileges --ulimit core=0 --memory 1792m --memory-swap 1792m \
 --log-driver none --tmpfs /data:size=128m -p 127.0.0.1::6379 \
 -v "$temporary/valkey.conf:/etc/valkey/valkey.conf:ro" \
 -v "$repo_dir/.runtime/users.acl:/run/secrets/users.acl:ro" \
 "$GATI_VALKEY_IMAGE" valkey-server /etc/valkey/valkey.conf)
export GATI_TEST_ADDR GATI_TEST_PASSWORD_FILE="$repo_dir/.runtime/app-password" GATI_INTEGRATION=1 GATI_SCALING_BENCH=1
GATI_TEST_ADDR=$(docker port "$container" 6379/tcp)
git rev-parse HEAD > "$output/revision.txt"
git diff --binary > "$output/source.patch"
go version > "$output/toolchain.txt"
LC_ALL=C lscpu > "$output/hardware.txt"
go test -run '^$' -bench '^BenchmarkEligibleSnapshot$' -benchtime=3x -count=3 -timeout=30m ./internal/store | tee "$output/store.txt"
go test -run '^$' -bench '^BenchmarkCanReach$' -benchtime=1s -count=5 ./internal/geography | tee "$output/geography.txt"

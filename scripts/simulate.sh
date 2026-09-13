#!/usr/bin/env bash
# Own a fresh isolated store/API for every run. No production app store access.
set -euo pipefail
repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_dir"
source scripts/env.sh
source deploy/images.env
python3 scripts/init-secrets.py --simulation
mkdir -p bin reports/local
go build -tags simulation -o bin/gati-simulation ./cmd/gati
go build -o bin/gati-simulate ./cmd/simulate
container=$(docker run -d --rm --read-only --user 999:999 --cap-drop ALL \
  --security-opt no-new-privileges --ulimit core=0 --memory 256m --memory-swap 256m \
  --log-driver none --tmpfs /data:size=128m -p 127.0.0.1::6379 \
  -v "$repo_dir/deploy/valkey.dev.conf:/etc/valkey/valkey.conf:ro" \
  -v "$repo_dir/.runtime/simulation/users.acl:/run/secrets/users.acl:ro" \
  "$GATI_VALKEY_IMAGE" valkey-server /etc/valkey/valkey.conf)
api_pid=''
cleanup(){ if [[ -n "$api_pid" ]];then kill "$api_pid" 2>/dev/null || true;wait "$api_pid" 2>/dev/null || true;fi;docker rm -f "$container" >/dev/null 2>&1 || true; }
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
store_address=$(docker port "$container" 6379/tcp)
# Readiness checks use TCP only, never print the store credential.
python3 - "$store_address" <<'PY'
import socket,sys,time
host,port=sys.argv[1].split(':')
for _ in range(100):
    try:
        with socket.create_connection((host,int(port)),timeout=.2):break
    except OSError:time.sleep(.05)
else:raise SystemExit('Simulation store did not start')
PY
# A fresh free loopback port avoids touching a pre-existing API.
api_port=$(python3 -c 'import socket; s=socket.socket();s.bind(("127.0.0.1",0));print(s.getsockname()[1]);s.close()')
bin/gati-simulation -listen "127.0.0.1:$api_port" -config config/simulation.yaml \
  -store-address "$store_address" -store-password-file .runtime/simulation/app-password \
  -simulation-control-file .runtime/simulation/control-password > reports/local/simulation-startup.log 2>&1 &
api_pid=$!
for _ in {1..100};do
  if ! kill -0 "$api_pid" 2>/dev/null;then cat reports/local/simulation-startup.log;exit 1;fi
  if rg -q 'GATI API started.' reports/local/simulation-startup.log;then break;fi
  sleep .05
done
bin/gati-simulate -target "http://127.0.0.1:$api_port" -scenario "simulation/scenarios/${SCENARIO:-tirana-evening}.yaml" -seed "${SEED:-42}"

kill "$api_pid";wait "$api_pid";api_pid=""
GATI_SIM_INTEGRATION=1 GATI_TEST_ADDR="$store_address" GATI_TEST_PASSWORD_FILE="$repo_dir/.runtime/simulation/app-password" go test -race -p 1 -tags simulation -count=1 ./internal/store ./internal/httpapi -run "TestSimulationClockAndNamespace|TestSimulatedContinuousActivationAndLateAdmission|TestKnownColludingInvitationInference"

if [[ "${GATI_PRIVACY_PROBE:-0}" == 1 ]];then
  printf 'Known inference counterexample recorded in reports/local/privacy-probe.json. Privacy gate FAILED (expected exit 2).\n'
  exit 2
fi

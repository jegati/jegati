#!/usr/bin/env bash
# Own a fresh isolated store/API for every run. No production app store access.
set -euo pipefail
repo_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_dir"
source scripts/env.sh
source deploy/images.env
python3 scripts/init-secrets.py --simulation
mkdir -p bin reports/local
run_output="${OUTPUT:-reports/local/simulation}"
sim_config="${SIM_CONFIG:-config/simulation.yaml}"
if [[ -f "$run_output/report.json" && -n "${OUTPUT:-}" ]];then printf "Report already exists; choose a new OUTPUT directory.\n";exit 1;fi
mkdir -p "$run_output"
go build -tags simulation -o bin/gati-simulation ./cmd/gati
go build -o bin/gati-simulate ./cmd/simulate
# Work from a validated nonsecret snapshot. Only the profile is converted; the
# caller's file stays untouched. Outbound push must be disabled in simulations.
# JSON is valid strict YAML for the application.
bin/gati-simulation -mode config-show -config "$sim_config" > "$run_output/source-config.json"
python3 - "$run_output" <<'PYCONFIG'
import hashlib,json,pathlib,platform,subprocess,sys
out=pathlib.Path(sys.argv[1]);c=json.loads((out/'source-config.json').read_text());c['profile']='simulation';c['notifications']['push_enabled']=False
(out/'effective-config.json').write_text(json.dumps(c,separators=(',',':'))+'\n')
manifest={'source_revision':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'platform':platform.platform(),'machine':platform.machine(),'logical_cpus':__import__('os').cpu_count(),'hashes':{}}
for name in ['bin/gati-simulation','bin/gati-simulate','data/tirana/intersections.json']:
 manifest['hashes'][name]=hashlib.sha256(pathlib.Path(name).read_bytes()).hexdigest()
(out/'run-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
PYCONFIG
sim_config="$run_output/effective-config.json"
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
# Wait for the authenticated operation required by the API. Docker's published
# port can accept TCP connections briefly before Valkey has loaded its ACL.
python3 - "$store_address" .runtime/simulation/app-password <<'PY'
import pathlib,socket,sys,time
host,port=sys.argv[1].rsplit(':',1)
password=pathlib.Path(sys.argv[2]).read_text().strip()
def command(*parts):
    encoded=[part.encode() for part in parts]
    return b'*'+str(len(encoded)).encode()+b'\r\n'+b''.join(
        b'$'+str(len(part)).encode()+b'\r\n'+part+b'\r\n' for part in encoded)
for _ in range(300):
    try:
        with socket.create_connection((host,int(port)),timeout=.5) as connection:
            stream=connection.makefile('rb')
            connection.sendall(command('AUTH','app',password))
            if stream.readline()!=b'+OK\r\n':raise OSError
            connection.sendall(command('PING'))
            if stream.readline()==b'+PONG\r\n':break
    except OSError:pass
    time.sleep(.1)
else:raise SystemExit('Simulation store did not become authentication-ready')
PY
# A fresh free loopback port avoids touching a pre-existing API.
api_port=$(python3 -c 'import socket; s=socket.socket();s.bind(("127.0.0.1",0));print(s.getsockname()[1]);s.close()')
bin/gati-simulation -listen "127.0.0.1:$api_port" -config "$sim_config" \
  -store-address "$store_address" -store-password-file .runtime/simulation/app-password \
  -simulation-control-file .runtime/simulation/control-password > "$run_output/startup.log" 2>&1 &
api_pid=$!
ready=0
for _ in {1..2400};do
  if ! kill -0 "$api_pid" 2>/dev/null;then cat "$run_output/startup.log";exit 1;fi
  if grep -q 'GATI API started.' "$run_output/startup.log";then ready=1;break;fi
  sleep .05
done
if [[ "$ready" != 1 ]];then printf "Simulation API readiness timed out.\n";exit 1;fi
seed_args=()
if [[ -n "${SEED:-}" ]]; then seed_args=(-seed "$SEED"); fi
bin/gati-simulate -target "http://127.0.0.1:$api_port" -scenario "simulation/scenarios/${SCENARIO:-tirana-evening}.yaml" "${seed_args[@]}" -output "$run_output"

kill "$api_pid";wait "$api_pid";api_pid=""
if [[ "${INTEGRATION_CHECKS:-1}" == 1 ]]; then
GATI_SIM_INTEGRATION=1 GATI_TEST_ADDR="$store_address" GATI_TEST_PASSWORD_FILE="$repo_dir/.runtime/simulation/app-password" go test -race -p 1 -tags simulation -count=1 ./internal/store ./internal/httpapi -run "TestSimulatedArrivalRenewal|TestSimulationClockAndNamespace|TestSimulatedContinuousActivationAndLateAdmission|TestKnownColludingInvitationInference"

fi

if [[ "${GATI_PRIVACY_PROBE:-0}" == 1 ]];then
  printf 'Known inference counterexample recorded in reports/local/privacy-probe.json. Privacy gate FAILED (expected exit 2).\n'
  exit 2
fi

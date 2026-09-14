#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source scripts/env.sh
fuzz_seconds=${FUZZ_SECONDS:-30}
[[ "$fuzz_seconds" =~ ^[0-9]+$ && "$fuzz_seconds" -ge 1 && "$fuzz_seconds" -le 3600 ]] || { echo 'FUZZ_SECONDS must be 1..3600'; exit 1; }
for entry in 'config FuzzDecode' 'httpapi FuzzRequestParsers' 'httpapi FuzzStrictRequestRoundTrip' 'httpapi FuzzCapability' 'geography FuzzGrid'; do
 read -r package target <<< "$entry"
 go test "./internal/$package" -run '^$' -fuzz "^$target$" -fuzztime "${fuzz_seconds}s" -parallel 2
done

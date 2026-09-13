#!/usr/bin/env bash
# Read-only developer prerequisite check. No installation, daemon startup or auth.
set -uo pipefail

missing=0
check_tool() {
  local name="$1"
  local required="$2"
  shift 2
  local output
  if ! command -v "$name" >/dev/null 2>&1; then
    printf '%-12s missing (%s)\n' "$name" "$required"
    if [[ "$required" == required ]]; then missing=1; fi
    return
  fi
  if output=$("$name" "$@" 2>&1); then
    printf '%-12s %s\n' "$name" "${output%%$'\n'*}"
  else
    printf '%-12s installed but version check failed (%s)\n' "$name" "$required"
    if [[ "$required" == required ]]; then missing=1; fi
  fi
}

printf 'GATI development prerequisites (read-only check)\n'
check_tool git required --version
check_tool curl required --version
check_tool make required --version
check_tool go required version
check_tool node required --version
check_tool npm required --version
check_tool docker required --version
check_tool docker-compose optional version
check_tool gh optional --version

if command -v docker >/dev/null 2>&1; then
  if docker compose version >/dev/null 2>&1 || docker-compose version >/dev/null 2>&1; then
    printf '%-12s available\n' compose
  else
    printf '%-12s missing or unusable (required)\n' compose
    missing=1
  fi
fi

printf '\nThis does not check supported version ranges, Docker daemon access,\nGitHub authentication, application tests or deployment readiness.\n'
exit "$missing"

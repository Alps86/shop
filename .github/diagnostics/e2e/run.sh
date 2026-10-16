#!/usr/bin/env bash
set -Eeuo pipefail
set +x
# shellcheck source=.github/diagnostics/e2e/common.sh
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
[[ ! -e "$E2E_APPROVED" ]]
mkdir -m 0700 "$E2E_APPROVED"
python3 "$E2E_ASSETS/aggregate.py" --status "$E2E_APPROVED" guard none
E2E_PHASE=guard
trap cleanup_e2e EXIT
trap 'exit 143' TERM
trap 'exit 130' INT
[[ ! -e "$E2E_PRIVATE" ]]
mkdir -m 0700 "$E2E_PRIVATE"
exec > "$E2E_PRIVATE/runner-private.log" 2>&1
[[ "${GITHUB_WORKFLOW:-}" == Quality ]]
[[ "${GITHUB_RUN_NUMBER:-}" == 50 ]]
[[ "${GITHUB_RUN_ATTEMPT:-}" == 1 ]]
[[ "${GITHUB_JOB:-}" == race ]]
[[ "${GITHUB_EVENT_NAME:-}" == pull_request ]]
[[ "${GITHUB_HEAD_REF:-}" == codex/shop-performance ]]
jq -e '.action == "synchronize"' "$GITHUB_EVENT_PATH" > /dev/null
phase collectors
stop_original_events
# shellcheck source=.github/diagnostics/e2e/collect.sh
source "$E2E_ASSETS/collect.sh"
# shellcheck source=.github/diagnostics/e2e/prepare.sh
source "$E2E_ASSETS/prepare.sh"
phase build
prepare_e2e
phase recreate
migrate_e2e
E2E_COOKIE="$(jq -er '.OrderCookie' "$SHOP_PERFORMANCE_STATE_FILE")"
printf '%s\n' "$E2E_COOKIE" > "$E2E_PRIVATE/private-patterns.txt"
printf '%s\n' "${E2E_COOKIE#*=}" >> "$E2E_PRIVATE/private-patterns.txt"
printf '%s\n' Synthetischer >> "$E2E_PRIVATE/private-patterns.txt"
phase warmup
stamp > "$E2E_PRIVATE/warmup-start.utc"
"$SHOP_HEY" -n 100 -c 10 -o csv -disable-redirects -disable-compression -t 15 -H "Cookie: $E2E_COOKIE" "$SHOP_PERFORMANCE_PUBLIC_URL/de-DE/suche?q=Synthetischer" > "$E2E_PRIVATE/warmup.csv" 2> "$E2E_PRIVATE/hey-warmup.stderr"
stamp > "$E2E_PRIVATE/warmup-end.utc"
phase profile-start
start_collectors || true
start_jfr || true
control start
E2E_GO_STARTED=1
phase load
stamp > "$E2E_PRIVATE/load-start.utc"
"$SHOP_HEY" -n 1000 -c 10 -o csv -disable-redirects -disable-compression -t 15 -H "Cookie: $E2E_COOKIE" "$SHOP_PERFORMANCE_PUBLIC_URL/de-DE/suche?q=Synthetischer" > "$E2E_PRIVATE/load.csv" 2> "$E2E_PRIVATE/hey-load.stderr"
stamp > "$E2E_PRIVATE/load-end.utc"
phase profile-stop
stop_profiles
stop_collectors
identify_app after
cmp "$E2E_PRIVATE/app-before.json" "$E2E_PRIVATE/app-after.json"
checksum_state after
cmp "$E2E_PRIVATE/orders-before.sha256" "$E2E_PRIVATE/orders-after.sha256"
cmp "$E2E_PRIVATE/inventory-before.sha256" "$E2E_PRIVATE/inventory-after.sha256"
cmp "$E2E_PRIVATE/workflows-before.sha256" "$E2E_PRIVATE/workflows-after.sha256"
docker cp "$E2E_APP:/tmp/shop-e2e-private/." "$E2E_PRIVATE/profiles"
docker cp "$E2E_NGINX:/tmp/shop-e2e-nginx-times.log" "$E2E_PRIVATE/nginx-times.txt" || true
phase tools
bash "$E2E_ASSETS/approve.sh"
phase export
python3 "$E2E_ASSETS/aggregate.py" --verify-approved "$E2E_APPROVED"

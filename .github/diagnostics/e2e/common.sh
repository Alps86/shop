#!/usr/bin/env bash
set -Eeuo pipefail
set +x
umask 077
export PYTHONDONTWRITEBYTECODE=1
E2E_PRIVATE=/tmp/shop-e2e-private
export E2E_APPROVED=/tmp/shop-e2e-approved
export E2E_ASSETS
E2E_ASSETS="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
E2E_COLLECTORS=()
E2E_GO_STARTED=0
E2E_JFR_STARTED=0
benchmark() {
  docker compose -p shop-benchmark -f compose.yml -f compose.benchmark.yml "$@"
}
stamp() {
  date --utc --iso-8601=ns
}
control() {
  local started status=0
  started="$(date +%s%N)"
  # Shellumleitung bleibt bewusst beim unprivilegierten Runner.
  # shellcheck disable=SC2024
  sudo nsenter -t "$E2E_APP_PID" -n curl --noproxy '*' --fail --silent --show-error --max-time 5 -X POST --output "$E2E_PRIVATE/control-$1.txt" --write-out '%{http_code}' "http://127.0.0.1:6061/$1" > "$E2E_PRIVATE/control-$1.http" 2> "$E2E_PRIVATE/control-$1.stderr" || status=$?
  [[ "$(cat "$E2E_PRIVATE/control-$1.http")" == 204 ]] || status=1
  record_boundary "go-$1" "$started" "$status"
  return "$status"
}
record_boundary() {
  local confirmed=false
  [[ "$3" != 0 ]] || confirmed=true
  printf '{"startedUnixNanoseconds":%s,"finishedUnixNanoseconds":%s,"exitCode":%d,"confirmed":%s}\n' "$2" "$(date +%s%N)" "$3" "$confirmed" > "$E2E_PRIVATE/$1.json"
}

checksum_state() {
  local label="$1"
  docker exec "$E2E_PG" psql -XqAt -v ON_ERROR_STOP=1 -U shop -d shop -c "SELECT payload,fulfillment FROM \"$SHOP_PERFORMANCE_SCHEMA\".orders ORDER BY id" | sha256sum | cut -d' ' -f1 > "$E2E_PRIVATE/orders-$label.sha256"
  docker exec "$E2E_PG" psql -XqAt -v ON_ERROR_STOP=1 -U shop -d shop -c "SELECT payload FROM \"$SHOP_PERFORMANCE_SCHEMA\".records WHERE kind='inventory' ORDER BY id" | sha256sum | cut -d' ' -f1 > "$E2E_PRIVATE/inventory-$label.sha256"
  for workflow in "cart/$E2E_CART" "order/$E2E_ORDER"; do
    docker exec "$E2E_TEMPORAL" temporal workflow describe --address 127.0.0.1:7233 --namespace default --workflow-id "$workflow" --output json | jq -Sc . | sha256sum | cut -d' ' -f1
  done > "$E2E_PRIVATE/workflows-$label.sha256"
}
identify_app() {
  E2E_APP="$(benchmark ps -q shop)"
  E2E_APP_PID="$(docker inspect --format '{{.State.Pid}}' "$E2E_APP")"
  docker inspect --format '{"id":{{json .Id}},"pid":{{.State.Pid}},"started":{{json .State.StartedAt}},"restarts":{{.RestartCount}}}' "$E2E_APP" > "$E2E_PRIVATE/app-$1.json"
  docker top "$E2E_APP" -eo pid,ppid,comm > "$E2E_PRIVATE/app-process-$1.txt"
}
stop_profiles() {
  local failure=0
  if [[ "$E2E_GO_STARTED" == 1 ]]; then
    E2E_GO_STARTED=0
    control stop || failure=1
  fi
  if [[ "$E2E_JFR_STARTED" == 1 ]]; then
    E2E_JFR_STARTED=0
    stop_jfr || true
  fi
  return "$failure"
}
cleanup_e2e() {
  local code="$?" cleaned=0
  trap - EXIT
  stop_profiles || cleaned=1
  stop_collectors
  cleanup_containers || cleaned=1
  rm -rf -- "$E2E_PRIVATE" || cleaned=1
  python3 "$E2E_ASSETS/aggregate.py" --finish "$E2E_APPROVED" "$code" "${E2E_PHASE:-guard}" "$cleaned" || code=1
  [[ "$cleaned" == 0 ]] || code=1
  exit "$code"
}
cleanup_containers() {
  local failed=0
  if [[ -n "${E2E_OS:-}" ]]; then
    docker exec --user 0 "$E2E_OS" rm -f /tmp/shop-e2e.jfr /tmp/shop-e2e-safe.jfc > /dev/null 2>&1 || failed=1
  fi
  if [[ "${E2E_RECREATED:-0}" == 1 ]]; then
    remove_diagnostic_containers || failed=1
  fi
  return "$failed"
}

stop_collectors() {
  for collector in "${E2E_COLLECTORS[@]}"; do
    kill "$collector" 2>/dev/null || true
    wait "$collector" 2>/dev/null || true
  done
  E2E_COLLECTORS=()
}

stop_original_events() {
  local pid
  for file in /tmp/shop-demo-events.pid /tmp/shop-benchmark-events.pid; do
    read -r pid < "$file"
    [[ "$pid" =~ ^[1-9][0-9]*$ ]]
    kill -TERM -- "-$pid" 2>/dev/null || true
    sleep 1
    if kill -0 -- "-$pid" 2>/dev/null; then
      return 1
    fi
  done
  stamp > /tmp/shop-benchmark-results/resources/events-ended-before-e2e.utc
}

phase() {
  E2E_PHASE="$1"
  python3 "$E2E_ASSETS/aggregate.py" --phase "$E2E_APPROVED" "$1"
}

remove_diagnostic_containers() {
  local identities container failed=0
  identities="$(benchmark ps --all --quiet shop nginx varnish)" || return 1
  for container in $identities; do
    docker rm --force "$container" > /dev/null 2>&1 || failed=1
  done
  return "$failed"
}

#!/usr/bin/env bash
set -Eeuo pipefail
set +x
umask 077
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
    stop_jfr || failure=1
  fi
  return "$failure"
}
cleanup_e2e() {
  local status="$?"
  trap - EXIT
  stop_profiles || status=1
  stop_collectors
  if [[ -n "${E2E_APP:-}" ]]; then
    benchmark stop shop nginx varnish > "$E2E_PRIVATE/stop-private.log" 2>&1 || status=1
  fi
  printf '{"exitCode":%d,"utc":"%s"}\n' "$status" "$(stamp)" > "$E2E_PRIVATE/runner-status.json"
  exit "$status"
}

stop_collectors() {
  for collector in "${E2E_COLLECTORS[@]}"; do
    kill "$collector" 2>/dev/null || true
    wait "$collector" 2>/dev/null || true
  done
  E2E_COLLECTORS=()
}

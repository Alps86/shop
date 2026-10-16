#!/usr/bin/env bash
# shellcheck source=.github/diagnostics/e2e/common.sh
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
start_collectors() {
  mapfile -t E2E_CONTAINERS < <(benchmark ps --status running -q)
  [[ "${#E2E_CONTAINERS[@]}" -gt 0 ]] || return 1
  printf '%s\n' "${E2E_CONTAINERS[@]}" > "$E2E_PRIVATE/container-ids.txt"
  for container in "${E2E_CONTAINERS[@]}"; do
    printf 'container %s\n' "$container"
    docker top "$container" -eo pid,ppid,comm
  done > "$E2E_PRIVATE/container-processes.txt"
  container_stats > "$E2E_PRIVATE/container-stats.jsonl" 2> "$E2E_PRIVATE/container-stats.stderr" &
  E2E_COLLECTORS+=("$!")
  timeout 65s pidstat -H -h -r -u -p "$E2E_APP_PID" 1 > "$E2E_PRIVATE/app-pidstat.txt" 2> "$E2E_PRIVATE/app-pidstat.stderr" &
  E2E_COLLECTORS+=("$!")
  pg_waits > "$E2E_PRIVATE/pg-waits.jsonl" 2> "$E2E_PRIVATE/pg-waits.stderr" &
  E2E_COLLECTORS+=("$!")
  timeout 65s docker exec "$E2E_VARNISH" varnishlog -g request -i VCL_Log,Timestamp -q 'VCL_Log ~ "^diag:[0-9]+$"' > "$E2E_PRIVATE/varnish-times.txt" 2> "$E2E_PRIVATE/varnish-times.stderr" &
  E2E_COLLECTORS+=("$!")
}
pg_waits() {
  local query="SELECT json_build_object('utc',clock_timestamp(),'activity',COALESCE(json_agg(json_build_object('pid',pid,'backend_type',backend_type,'state',state,'wait_event_type',wait_event_type,'wait_event',wait_event,'query_start',query_start,'xact_start',xact_start,'blockers',pg_blocking_pids(pid))),'[]'::json)) FROM pg_stat_activity WHERE datname=current_database() AND pid<>pg_backend_pid()"
  local started="$SECONDS"
  for ((sample=0; sample<300; sample++)); do
    ((SECONDS - started < 60)) || break
    timeout 2s docker exec "$E2E_PG" psql -XqAt -v ON_ERROR_STOP=1 -U shop -d shop -c "$query"
    sleep 0.2
  done
}
start_jfr() {
  docker exec "$E2E_OS" sh -ec 'for path in /proc/[0-9]*; do test -r "$path/comm" || continue; read -r name < "$path/comm"; if test "$name" = java; then printf "%s " "${path##*/}"; awk "/^Uid:/{print \$2}" "$path/status"; fi; done' > "$E2E_PRIVATE/java-identity.txt"
  [[ "$(wc -l < "$E2E_PRIVATE/java-identity.txt")" == 1 ]] || return 1
  read -r E2E_JAVA_PID E2E_JAVA_UID < "$E2E_PRIVATE/java-identity.txt" || return 1
  docker cp "$E2E_ASSETS/safe.jfc" "$E2E_OS:/tmp/shop-e2e-safe.jfc" || return 1
  local started status=0
  started="$(date +%s%N)"
  timeout 10s docker exec --user "$E2E_JAVA_UID" "$E2E_OS" /usr/share/opensearch/jdk/bin/jcmd "$E2E_JAVA_PID" JFR.start name=shop_e2e settings=/tmp/shop-e2e-safe.jfc duration=60s maxsize=64m filename=/tmp/shop-e2e.jfr > "$E2E_PRIVATE/jfr-start.txt" || status=$?
  [[ "$status" != 0 ]] || E2E_JFR_STARTED=1
  grep -Eq '^Started recording [0-9]+\.' "$E2E_PRIVATE/jfr-start.txt" || status=1
  record_boundary jfr-start "$started" "$status"
  [[ "$status" == 0 ]] || return 1
  E2E_JFR_STARTED=1
}

container_stats() {
  timeout 65s docker stats --format '{{json .}}' "${E2E_CONTAINERS[@]}" | while IFS= read -r sample; do
    printf '{"utc":"%s","sample":%s}\n' "$(stamp)" "$sample"
  done
}

stop_jfr() {
  local started status=0
  started="$(date +%s%N)"
  timeout 10s docker exec --user "$E2E_JAVA_UID" "$E2E_OS" /usr/share/opensearch/jdk/bin/jcmd "$E2E_JAVA_PID" JFR.stop name=shop_e2e filename=/tmp/shop-e2e.jfr > "$E2E_PRIVATE/jfr-stop.txt" || status=$?
  grep -Eq '^Stopped recording "shop_e2e"' "$E2E_PRIVATE/jfr-stop.txt" || status=1
  record_boundary jfr-stop "$started" "$status"
  return "$status"
}

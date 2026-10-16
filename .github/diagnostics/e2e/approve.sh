#!/usr/bin/env bash
set -Eeuo pipefail
set +x
umask 077
private=/tmp/shop-e2e-private
approved=/tmp/shop-e2e-approved
[[ ! -e "$approved" ]]
jq -e '.State == "stopped"' "$private/profiles/state.json" > /dev/null
for sample in warmup:100 load:1000; do
  name="${sample%:*}"
  count="${sample#*:}"
  awk -F, -v expected="$count" 'NR==1 {if($0 != "response-time,DNS+dialup,DNS,Request-write,Response-delay,Response-read,status-code,offset") bad++} NR>1 {n++; if(NF != 8 || $7 != 200) bad++; for(i=1;i<=NF;i++) if($i !~ /^[0-9]+([.][0-9]+)?$/) bad++} END {if(n != expected || bad) exit 1}' "$private/$name.csv"
done
go tool pprof -top "$private/shop-e2e" "$private/profiles/cpu.pprof" > "$private/go-cpu-top.txt"
go tool pprof -top -cum "$private/shop-e2e" "$private/profiles/cpu.pprof" > "$private/go-cpu-cumulative.txt"
go tool pprof -raw "$private/shop-e2e" "$private/profiles/cpu.pprof" > "$private/go-cpu-samples.txt"
go tool trace -d=parsed "$private/profiles/execution.trace" > "$private/go-trace.txt"
for profile in net sync syscall sched; do
  go tool trace -pprof="$profile" "$private/profiles/execution.trace" > "$private/go-$profile.pprof"
  go tool pprof -top "$private/go-$profile.pprof" > "$private/go-$profile-top.txt"
done
os="$(docker compose -p shop-benchmark -f compose.yml -f compose.benchmark.yml ps -q opensearch)"
docker exec "$os" /usr/share/opensearch/jdk/bin/jfr summary /tmp/shop-e2e.jfr > "$private/jfr-summary.txt"
docker exec "$os" /usr/share/opensearch/jdk/bin/jfr view hot-methods /tmp/shop-e2e.jfr > "$private/jfr-hot-methods.txt"
docker exec "$os" /usr/share/opensearch/jdk/bin/jfr print --json /tmp/shop-e2e.jfr > "$private/jfr-events.json"
python3 "$(dirname "${BASH_SOURCE[0]}")/events.py" "$private/jfr-summary.txt"
for phase in go-start go-stop jfr-start jfr-stop; do
  jq -e '.confirmed == true and .exitCode == 0 and .startedUnixNanoseconds > 0 and .finishedUnixNanoseconds >= .startedUnixNanoseconds' "$private/$phase.json" > /dev/null
done
jq '{state:.State, stoppedUTC:.UTC, pid:.PID}' "$private/profiles/state.json" > "$private/go-profile-state.json"
jq 'map({source,sha256})' "$private/overlay/manifest.json" > "$private/overlay-manifest.json"
# Rohprofile bleiben privat. Nur feste Standardtoolausgaben und numerische Korrelation.
files=(warmup.csv load.csv go-start.json go-stop.json jfr-start.json jfr-stop.json go-profile-state.json binary.sha256 buildinfo.txt overlay-manifest.json bind-modes.txt container-ids.txt container-processes.txt java-identity.txt app-process-before.txt app-process-after.txt orders-before.sha256 orders-after.sha256 inventory-before.sha256 inventory-after.sha256 workflows-before.sha256 workflows-after.sha256 go-cpu-cumulative.txt go-cpu-samples.txt jfr-events.json go-cpu-top.txt go-trace.txt go-net-top.txt go-sync-top.txt go-syscall-top.txt go-sched-top.txt jfr-summary.txt jfr-hot-methods.txt nginx-times.txt varnish-times.txt pg-waits.jsonl container-stats.jsonl app-pidstat.txt app-before.json app-after.json load-start.utc load-end.utc warmup-start.utc warmup-end.utc)
for file in "${files[@]}"; do
  [[ -f "$private/$file" && ! -L "$private/$file" ]]
  if grep -F -f "$private/private-patterns.txt" "$private/$file" > /dev/null; then
    exit 1
  fi
  if grep -E 'https?://|Cookie:|Authorization:|Set-Cookie:|(^|[^0-9])[0-9]{1,3}(\.[0-9]{1,3}){3}([^0-9]|$)' "$private/$file" > /dev/null; then
    exit 1
  fi
done
mkdir -m 0700 "$approved"
for file in "${files[@]}"; do
  cp "$private/$file" "$approved/$file"
done

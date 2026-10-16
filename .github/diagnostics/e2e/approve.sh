#!/usr/bin/env bash
set -Eeuo pipefail
set +x
umask 077
private=/tmp/shop-e2e-private
approved=/tmp/shop-e2e-approved
assets="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
for profile in cpu net sync syscall sched; do
  source="$private/profiles/cpu.pprof"
  if [[ "$profile" != cpu ]]; then
    go tool trace -pprof="$profile" "$private/profiles/execution.trace" > "$private/go-$profile.pprof"
    source="$private/go-$profile.pprof"
  fi
  go tool pprof -top -unit=ns -nodecount=30 "$private/shop-e2e" "$source" > "$private/go-$profile-top.txt"
done
go tool pprof -tags -unit=ns "$private/shop-e2e" "$private/profiles/cpu.pprof" > "$private/go-cpu-tags.txt" || true
go run cmd/nm "$private/shop-e2e" > "$private/symbols.txt"
go tool trace -d=parsed "$private/profiles/execution.trace" > "$private/go-trace.txt"
os="$(docker compose -p shop-benchmark -f compose.yml -f compose.benchmark.yml ps -q opensearch)" || os=""
docker exec "$os" /usr/share/opensearch/jdk/bin/jfr summary /tmp/shop-e2e.jfr > "$private/jfr-summary.txt" || true
docker exec "$os" /usr/share/opensearch/jdk/bin/jfr print --json /tmp/shop-e2e.jfr > "$private/jfr-events.json" || true
python3 "$assets/aggregate.py" "$private" "$approved"

#!/usr/bin/env bash
prepare_e2e() {
  [[ "$SHOP_PERFORMANCE_SCHEMA" =~ ^[a-zA-Z_][a-zA-Z0-9_]*$ ]]
  [[ -s "$SHOP_PERFORMANCE_STATE_FILE" ]]
  E2E_PG="$(benchmark ps -q postgres)"
  export E2E_OS
  E2E_OS="$(benchmark ps -q opensearch)"
  export E2E_TEMPORAL
  E2E_TEMPORAL="$(benchmark ps -q temporal)"
  E2E_ORDER="$(jq -er '.Order' "$SHOP_PERFORMANCE_STATE_FILE")"
  [[ "$E2E_ORDER" =~ ^[a-zA-Z0-9_-]+$ ]]
  E2E_CART="$(docker exec "$E2E_PG" psql -XqAt -v ON_ERROR_STOP=1 -U shop -d shop -c "SELECT cart_id FROM \"$SHOP_PERFORMANCE_SCHEMA\".orders WHERE id='$E2E_ORDER'")"
  [[ "$E2E_CART" =~ ^[a-zA-Z0-9_-]+$ ]]
  checksum_state before
  python3 "$E2E_ASSETS/generate.py" "$E2E_PRIVATE/overlay"
  CGO_ENABLED=0 go build -trimpath -mod=readonly -overlay "$E2E_PRIVATE/overlay/overlay.json" -o "$E2E_PRIVATE/shop-e2e" ./cmd/shop
  chmod 0555 "$E2E_PRIVATE/shop-e2e"
  sha256sum "$E2E_PRIVATE/shop-e2e" | cut -d' ' -f1 > "$E2E_PRIVATE/binary.sha256"
  go version -m "$E2E_PRIVATE/shop-e2e" > "$E2E_PRIVATE/buildinfo.txt"
  python3 "$E2E_ASSETS/proxy.py" "$E2E_PRIVATE"
  chmod 0444 "$E2E_PRIVATE/nginx.conf" "$E2E_PRIVATE/default.vcl"
  stat -c '%a %u %g' "$E2E_PRIVATE/shop-e2e" "$E2E_PRIVATE/nginx.conf" "$E2E_PRIVATE/default.vcl" > "$E2E_PRIVATE/bind-modes.txt"
  write_override
}
write_override() {
  cat > "$E2E_PRIVATE/compose.yml" <<YAML
services:
  shop:
    entrypoint: [/shop-e2e]
    environment:
      SHOP_E2E_PROFILE: "1"
    volumes:
      - $E2E_PRIVATE/shop-e2e:/shop-e2e:ro
  nginx:
    volumes:
      - $E2E_PRIVATE/nginx.conf:/etc/nginx/shop.conf:ro
  varnish:
    volumes:
      - $E2E_PRIVATE/default.vcl:/etc/varnish/source.vcl:ro
YAML
}
migrate_e2e() {
  docker compose -p shop-benchmark -f compose.yml -f compose.benchmark.yml -f "$E2E_PRIVATE/compose.yml" up --detach --no-deps --wait --wait-timeout 120 shop nginx varnish > "$E2E_PRIVATE/recreate-private.log" 2>&1
  identify_app before
  export E2E_NGINX
  E2E_NGINX="$(benchmark ps -q nginx)"
  export E2E_VARNISH
  E2E_VARNISH="$(benchmark ps -q varnish)"
}

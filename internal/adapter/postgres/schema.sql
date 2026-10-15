CREATE TABLE IF NOT EXISTS records (
 kind text NOT NULL,
 id text NOT NULL,
 payload jsonb NOT NULL,
 PRIMARY KEY(kind,id)
);
CREATE TABLE IF NOT EXISTS users (
 id text PRIMARY KEY,
 profile text NOT NULL,
 payload jsonb NOT NULL
);
CREATE TABLE IF NOT EXISTS sessions (
 token_hash text PRIMARY KEY,
 user_id text NOT NULL REFERENCES users(id),
 issued_at timestamptz NOT NULL
);
CREATE TABLE IF NOT EXISTS carts (
 id text PRIMARY KEY,
 user_id text NOT NULL REFERENCES users(id),
 revision bigint NOT NULL,
 payload jsonb NOT NULL
);
CREATE TABLE IF NOT EXISTS orders (
 id text PRIMARY KEY,
 cart_id text NOT NULL UNIQUE REFERENCES carts(id),
 user_id text NOT NULL REFERENCES users(id),
 payload jsonb NOT NULL,
 fulfillment jsonb NOT NULL
);
CREATE TABLE IF NOT EXISTS projection_versions (
 product_id text PRIMARY KEY,
 revision bigint NOT NULL,
 deleted boolean NOT NULL DEFAULT false
);
CREATE TABLE IF NOT EXISTS sku_registry (
 sku text PRIMARY KEY,
 product_id text NOT NULL REFERENCES projection_versions(product_id)
);
CREATE TABLE IF NOT EXISTS commands (
 scope text NOT NULL,
 command_id text NOT NULL,
 payload jsonb NOT NULL,
 result jsonb NOT NULL,
 PRIMARY KEY(scope,command_id)
);
CREATE TABLE IF NOT EXISTS datasets (
 version text PRIMARY KEY,
 digest text NOT NULL,
 source_url text NOT NULL,
 retrieved_at timestamptz NOT NULL,
 license_note text NOT NULL,
 published_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS outbox (
 sequence bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 envelope jsonb NOT NULL
);
CREATE TABLE IF NOT EXISTS process_outbox (
 sequence bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 envelope jsonb NOT NULL,
 acknowledged boolean NOT NULL DEFAULT false
);
CREATE TABLE IF NOT EXISTS cdc_checkpoints (
 consumer text PRIMARY KEY,
 lsn pg_lsn NOT NULL
);
CREATE TABLE IF NOT EXISTS migrations (
 version integer PRIMARY KEY,
 applied_at timestamptz NOT NULL DEFAULT now()
);
INSERT INTO migrations(version) VALUES(1) ON CONFLICT DO NOTHING;
CREATE TABLE IF NOT EXISTS cdc_event_registry (
 event_id text PRIMARY KEY,
 envelope jsonb NOT NULL
);
CREATE TABLE IF NOT EXISTS cdc_events (
 consumer text NOT NULL,
 event_id text NOT NULL REFERENCES cdc_event_registry(event_id),
 PRIMARY KEY(consumer,event_id)
);

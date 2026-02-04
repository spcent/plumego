-- Idempotency keys schema (PostgreSQL)

CREATE TABLE IF NOT EXISTS idempotency_keys (
  key           VARCHAR(128) PRIMARY KEY,
  request_hash  VARCHAR(128) NOT NULL,
  status        VARCHAR(16) NOT NULL,
  response      BYTEA NULL,
  created_at    TIMESTAMP NOT NULL,
  updated_at    TIMESTAMP NOT NULL,
  expires_at    TIMESTAMP NULL
);

CREATE INDEX IF NOT EXISTS idx_idempotency_expires
  ON idempotency_keys (expires_at);

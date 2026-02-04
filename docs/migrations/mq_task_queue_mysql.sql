-- Durable task queue schema (MySQL 8+)

CREATE TABLE IF NOT EXISTS mq_tasks (
  id           VARCHAR(64) PRIMARY KEY,
  topic        VARCHAR(255) NOT NULL,
  tenant_id    VARCHAR(64) NOT NULL,
  payload      BLOB NOT NULL,
  meta         TEXT NULL,
  priority     INT NOT NULL DEFAULT 20,
  dedupe_key   VARCHAR(128) NULL,

  status       VARCHAR(16) NOT NULL,
  attempts     INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 5,

  available_at TIMESTAMP NOT NULL,
  lease_owner  VARCHAR(64) NULL,
  lease_until  TIMESTAMP NULL,
  expires_at   TIMESTAMP NULL,

  last_error   TEXT NULL,
  last_error_at TIMESTAMP NULL,

  created_at   TIMESTAMP NOT NULL,
  updated_at   TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS mq_task_dlq (
  id         BIGINT PRIMARY KEY AUTO_INCREMENT,
  task_id    VARCHAR(64) NOT NULL,
  topic      VARCHAR(255) NOT NULL,
  tenant_id  VARCHAR(64) NOT NULL,
  payload    BLOB NOT NULL,
  meta       TEXT NULL,
  reason     TEXT NOT NULL,
  failed_at  TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS mq_task_attempts (
  id          BIGINT PRIMARY KEY AUTO_INCREMENT,
  task_id     VARCHAR(64) NOT NULL,
  attempt_no  INT NOT NULL,
  reason      TEXT NULL,
  started_at  TIMESTAMP NOT NULL,
  finished_at TIMESTAMP NULL,
  success     BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_mq_tasks_ready
  ON mq_tasks (status, available_at, priority, created_at);

CREATE INDEX idx_mq_tasks_lease
  ON mq_tasks (lease_until, lease_owner);

CREATE INDEX idx_mq_tasks_tenant_topic
  ON mq_tasks (tenant_id, topic);

CREATE UNIQUE INDEX uq_mq_tasks_dedupe
  ON mq_tasks (tenant_id, dedupe_key);

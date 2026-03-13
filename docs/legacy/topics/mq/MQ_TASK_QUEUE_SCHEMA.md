# MQ Task Queue Schema

This document describes the recommended SQL schema for the durable task queue in `net/mq`.
It targets PostgreSQL and MySQL 8+ and is designed to work with `TaskStore` (SQL implementation).

**Tables**

Required:
- `mq_tasks`
- `mq_task_dlq`

Optional (recommended for audits):
- `mq_task_attempts`

**PostgreSQL Schema**

```sql
CREATE TABLE mq_tasks (
  id           VARCHAR(64) PRIMARY KEY,
  topic        VARCHAR(255) NOT NULL,
  tenant_id    VARCHAR(64) NOT NULL,
  payload      BYTEA NOT NULL,
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

CREATE TABLE mq_task_dlq (
  id         BIGSERIAL PRIMARY KEY,
  task_id    VARCHAR(64) NOT NULL,
  topic      VARCHAR(255) NOT NULL,
  tenant_id  VARCHAR(64) NOT NULL,
  payload    BYTEA NOT NULL,
  meta       TEXT NULL,
  reason     TEXT NOT NULL,
  failed_at  TIMESTAMP NOT NULL
);

CREATE TABLE mq_task_attempts (
  id          BIGSERIAL PRIMARY KEY,
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
```

**MySQL 8+ Schema**

```sql
CREATE TABLE mq_tasks (
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

CREATE TABLE mq_task_dlq (
  id         BIGINT PRIMARY KEY AUTO_INCREMENT,
  task_id    VARCHAR(64) NOT NULL,
  topic      VARCHAR(255) NOT NULL,
  tenant_id  VARCHAR(64) NOT NULL,
  payload    BLOB NOT NULL,
  meta       TEXT NULL,
  reason     TEXT NOT NULL,
  failed_at  TIMESTAMP NOT NULL
);

CREATE TABLE mq_task_attempts (
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
```

**Notes**
- `dedupe_key` allows `NULL`; unique index permits multiple `NULL` values.
- Reserve requires `FOR UPDATE SKIP LOCKED` for concurrent consumers.
- If your MySQL version does not support `SKIP LOCKED`, expect reduced concurrency safety.
- `meta` stores JSON as text; the SQL store expects JSON encoding.
- Use consistent timestamps in UTC to avoid lease and retry drift.
- To enable attempt logging, set `SQLConfig.EnableAttemptLog = true` and keep `SQLConfig.AttemptsTable` defaulted to `mq_task_attempts`.
- Use `SQLConfig.AttemptLogHook` to observe logging failures; attempt log errors do not block ack/release.

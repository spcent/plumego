-- SMS gateway message table (MySQL 8+)

CREATE TABLE IF NOT EXISTS sms_messages (
  id               VARCHAR(64) PRIMARY KEY,
  tenant_id        VARCHAR(64) NOT NULL,
  provider         VARCHAR(64) NULL,
  status           VARCHAR(32) NOT NULL,
  reason_code      VARCHAR(64) NULL,
  reason_detail    TEXT NULL,
  attempts         INT NOT NULL DEFAULT 0,
  max_attempts     INT NOT NULL DEFAULT 5,
  next_attempt_at  TIMESTAMP NULL,
  sent_at          TIMESTAMP NULL,
  provider_msg_id  VARCHAR(128) NULL,
  idempotency_key  VARCHAR(128) NULL,
  version          INT NOT NULL DEFAULT 0,
  created_at       TIMESTAMP NOT NULL,
  updated_at       TIMESTAMP NOT NULL
);

CREATE INDEX idx_sms_messages_tenant_status
  ON sms_messages (tenant_id, status, updated_at);

CREATE INDEX idx_sms_messages_next_attempt
  ON sms_messages (next_attempt_at);

CREATE UNIQUE INDEX uq_sms_messages_idem
  ON sms_messages (tenant_id, idempotency_key);

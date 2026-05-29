CREATE TABLE IF NOT EXISTS endpoints (
    endpoint_id    BIGINT AUTO_INCREMENT PRIMARY KEY,
    endpoint_key   VARCHAR(512) UNIQUE,
    endpoint_name  VARCHAR(512) NOT NULL,
    endpoint_group VARCHAR(512) NOT NULL,
    UNIQUE KEY uq_endpoint_name_group (endpoint_name, endpoint_group)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS endpoint_events (
    endpoint_event_id  BIGINT AUTO_INCREMENT PRIMARY KEY,
    endpoint_id        BIGINT    NOT NULL,
    event_type         VARCHAR(64) NOT NULL,
    event_timestamp    DATETIME(6) NOT NULL,
    CONSTRAINT fk_events_endpoint FOREIGN KEY (endpoint_id) REFERENCES endpoints(endpoint_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS endpoint_results (
    endpoint_result_id     BIGINT AUTO_INCREMENT PRIMARY KEY,
    endpoint_id            BIGINT      NOT NULL,
    success                TINYINT     NOT NULL,
    errors                 TEXT        NOT NULL,
    connected              TINYINT     NOT NULL,
    status                 INT         NOT NULL,
    dns_rcode              VARCHAR(64) NOT NULL,
    certificate_expiration BIGINT      NOT NULL,
    domain_expiration      BIGINT      NOT NULL,
    hostname               VARCHAR(512) NOT NULL,
    ip                     VARCHAR(128) NOT NULL,
    duration               BIGINT      NOT NULL,
    timestamp              DATETIME(6) NOT NULL,
    CONSTRAINT fk_results_endpoint FOREIGN KEY (endpoint_id) REFERENCES endpoints(endpoint_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS endpoint_result_conditions (
    endpoint_result_condition_id  BIGINT AUTO_INCREMENT PRIMARY KEY,
    endpoint_result_id            BIGINT      NOT NULL,
    condition                     TEXT        NOT NULL,
    success                       TINYINT     NOT NULL,
    CONSTRAINT fk_conditions_result FOREIGN KEY (endpoint_result_id) REFERENCES endpoint_results(endpoint_result_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS endpoint_uptimes (
    endpoint_uptime_id    BIGINT AUTO_INCREMENT PRIMARY KEY,
    endpoint_id           BIGINT  NOT NULL,
    hour_unix_timestamp   BIGINT  NOT NULL,
    total_executions      BIGINT  NOT NULL,
    successful_executions BIGINT  NOT NULL,
    total_response_time   BIGINT  NOT NULL,
    UNIQUE KEY uq_endpoint_hour (endpoint_id, hour_unix_timestamp),
    CONSTRAINT fk_uptimes_endpoint FOREIGN KEY (endpoint_id) REFERENCES endpoints(endpoint_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS endpoint_alerts_triggered (
    endpoint_alert_trigger_id     BIGINT AUTO_INCREMENT PRIMARY KEY,
    endpoint_id                   BIGINT       NOT NULL,
    configuration_checksum        VARCHAR(256) NOT NULL,
    resolve_key                   VARCHAR(512) NOT NULL,
    number_of_successes_in_a_row  INT          NOT NULL,
    UNIQUE KEY uq_endpoint_checksum (endpoint_id, configuration_checksum),
    CONSTRAINT fk_alerts_endpoint FOREIGN KEY (endpoint_id) REFERENCES endpoints(endpoint_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_endpoint_results_endpoint_id ON endpoint_results (endpoint_id);
CREATE INDEX idx_endpoint_uptimes_endpoint_id ON endpoint_uptimes (endpoint_id);
CREATE INDEX idx_endpoint_result_conditions_result_id ON endpoint_result_conditions (endpoint_result_id);

CREATE TABLE IF NOT EXISTS endpoint_configs (
    endpoint_config_id  BIGINT AUTO_INCREMENT PRIMARY KEY,
    endpoint_key        VARCHAR(512) NOT NULL UNIQUE,
    is_external         TINYINT      NOT NULL DEFAULT 0,
    payload_json        MEDIUMTEXT   NOT NULL,
    updated_at          DATETIME(6)  NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE INDEX idx_endpoint_configs_external ON endpoint_configs (is_external);

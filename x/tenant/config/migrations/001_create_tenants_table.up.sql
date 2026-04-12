-- Create tenants table for database-backed tenant configuration.
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(255) PRIMARY KEY,
    quota_requests_per_minute INT NOT NULL DEFAULT 0,
    quota_tokens_per_minute INT NOT NULL DEFAULT 0,
    quota_limits TEXT,
    allowed_models TEXT,
    allowed_tools TEXT,
    allowed_methods TEXT,
    allowed_paths TEXT,
    metadata TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for common list and update queries.
CREATE INDEX IF NOT EXISTS idx_tenants_updated_at ON tenants(updated_at);
CREATE INDEX IF NOT EXISTS idx_tenants_created_at ON tenants(created_at);

-- Comments for documentation.
COMMENT ON TABLE tenants IS 'Stores per-tenant configuration for multi-tenant applications';
COMMENT ON COLUMN tenants.id IS 'Unique tenant identifier';
COMMENT ON COLUMN tenants.quota_requests_per_minute IS 'Legacy per-minute request quota column (0 = unlimited)';
COMMENT ON COLUMN tenants.quota_tokens_per_minute IS 'Legacy per-minute token quota column (0 = unlimited)';
COMMENT ON COLUMN tenants.quota_limits IS 'JSON array of structured quota limits';
COMMENT ON COLUMN tenants.allowed_models IS 'JSON array of allowed model names (empty = allow all)';
COMMENT ON COLUMN tenants.allowed_tools IS 'JSON array of allowed tool names (empty = allow all)';
COMMENT ON COLUMN tenants.allowed_methods IS 'JSON array of allowed HTTP methods (empty = allow all)';
COMMENT ON COLUMN tenants.allowed_paths IS 'JSON array of allowed path patterns (empty = allow all)';
COMMENT ON COLUMN tenants.metadata IS 'JSON object for custom tenant metadata';

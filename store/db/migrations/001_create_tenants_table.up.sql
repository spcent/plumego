-- Create tenants table for multi-tenant configuration
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(255) PRIMARY KEY,
    quota_requests_per_minute INT NOT NULL DEFAULT 0,
    quota_tokens_per_minute INT NOT NULL DEFAULT 0,
    allowed_models TEXT,
    allowed_tools TEXT,
    metadata TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster queries by update time
CREATE INDEX IF NOT EXISTS idx_tenants_updated_at ON tenants(updated_at);

-- Index for common list operations
CREATE INDEX IF NOT EXISTS idx_tenants_created_at ON tenants(created_at);

-- Comments for documentation
COMMENT ON TABLE tenants IS 'Stores per-tenant configuration for multi-tenant applications';
COMMENT ON COLUMN tenants.id IS 'Unique tenant identifier';
COMMENT ON COLUMN tenants.quota_requests_per_minute IS 'Maximum requests per minute (0 = unlimited)';
COMMENT ON COLUMN tenants.quota_tokens_per_minute IS 'Maximum tokens per minute (0 = unlimited)';
COMMENT ON COLUMN tenants.allowed_models IS 'JSON array of allowed model names (empty = allow all)';
COMMENT ON COLUMN tenants.allowed_tools IS 'JSON array of allowed tool names (empty = allow all)';
COMMENT ON COLUMN tenants.metadata IS 'JSON object for custom tenant metadata';

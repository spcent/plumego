-- Create files table for file storage module
-- This table stores metadata for all uploaded files

CREATE TABLE IF NOT EXISTS files (
    -- Primary key
    id VARCHAR(64) PRIMARY KEY,

    -- Multi-tenancy
    tenant_id VARCHAR(64) NOT NULL,

    -- File information
    name VARCHAR(255) NOT NULL,
    path VARCHAR(1024) NOT NULL,
    size BIGINT NOT NULL CHECK (size >= 0),
    mime_type VARCHAR(128),
    extension VARCHAR(32),

    -- Deduplication
    hash VARCHAR(64) NOT NULL,

    -- Image metadata (NULL for non-images)
    width INTEGER CHECK (width IS NULL OR width > 0),
    height INTEGER CHECK (height IS NULL OR height > 0),
    thumbnail_path VARCHAR(1024),

    -- Storage backend
    storage_type VARCHAR(32) NOT NULL CHECK (storage_type IN ('local', 's3')),

    -- Flexible metadata as JSONB
    metadata JSONB,

    -- Audit fields
    uploaded_by VARCHAR(64),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Access tracking for LRU cleanup
    last_access_at TIMESTAMP,

    -- Soft delete
    deleted_at TIMESTAMP
);

-- Indexes for common queries

-- Index for tenant isolation
CREATE INDEX idx_files_tenant_id ON files(tenant_id) WHERE deleted_at IS NULL;

-- Index for deduplication lookup
CREATE INDEX idx_files_hash ON files(hash) WHERE deleted_at IS NULL;

-- Unique constraint: one file per hash per tenant (prevents duplicates)
CREATE UNIQUE INDEX idx_files_tenant_hash ON files(tenant_id, hash) WHERE deleted_at IS NULL;

-- Index for temporal queries (e.g., files created last 30 days)
CREATE INDEX idx_files_created_at ON files(created_at DESC) WHERE deleted_at IS NULL;

-- Index for soft delete queries
CREATE INDEX idx_files_deleted_at ON files(deleted_at) WHERE deleted_at IS NOT NULL;

-- Index for MIME type filtering (e.g., all images)
CREATE INDEX idx_files_mime_type ON files(mime_type) WHERE deleted_at IS NULL;

-- Index for uploader queries
CREATE INDEX idx_files_uploaded_by ON files(uploaded_by) WHERE deleted_at IS NULL;

-- Index for access-based cleanup (LRU)
CREATE INDEX idx_files_last_access_at ON files(last_access_at) WHERE deleted_at IS NULL;

-- Composite index for common query pattern: tenant + created_at
CREATE INDEX idx_files_tenant_created ON files(tenant_id, created_at DESC) WHERE deleted_at IS NULL;

-- JSONB index for metadata queries (GIN index)
CREATE INDEX idx_files_metadata ON files USING GIN (metadata) WHERE deleted_at IS NULL;

-- Comments for documentation
COMMENT ON TABLE files IS 'File storage metadata table with support for local and S3 storage';
COMMENT ON COLUMN files.id IS 'Unique file identifier (32-char hex)';
COMMENT ON COLUMN files.tenant_id IS 'Tenant identifier for multi-tenancy isolation';
COMMENT ON COLUMN files.hash IS 'SHA256 hash for deduplication';
COMMENT ON COLUMN files.storage_type IS 'Storage backend: local or s3';
COMMENT ON COLUMN files.metadata IS 'Flexible JSONB metadata for custom fields';
COMMENT ON COLUMN files.last_access_at IS 'Last access timestamp for LRU cleanup strategies';
COMMENT ON COLUMN files.deleted_at IS 'Soft delete timestamp, NULL means active';

-- Rollback migration: Drop files table and all associated indexes

-- Drop indexes first (PostgreSQL will drop them automatically with the table,
-- but explicitly dropping them makes the rollback clearer)
DROP INDEX IF EXISTS idx_files_metadata;
DROP INDEX IF EXISTS idx_files_tenant_created;
DROP INDEX IF EXISTS idx_files_last_access_at;
DROP INDEX IF EXISTS idx_files_uploaded_by;
DROP INDEX IF EXISTS idx_files_mime_type;
DROP INDEX IF EXISTS idx_files_deleted_at;
DROP INDEX IF EXISTS idx_files_created_at;
DROP INDEX IF EXISTS idx_files_tenant_hash;
DROP INDEX IF EXISTS idx_files_hash;
DROP INDEX IF EXISTS idx_files_tenant_id;

-- Drop the table
DROP TABLE IF EXISTS files;

-- Drop tenants table and indexes
DROP INDEX IF EXISTS idx_tenants_created_at;
DROP INDEX IF EXISTS idx_tenants_updated_at;
DROP TABLE IF EXISTS tenants;

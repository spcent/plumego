-- V0.1 schema fixture for migration upgrade tests.
-- Mirrors internal/database/migrate.go migrate001().

CREATE TABLE IF NOT EXISTS documents (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  slug TEXT,
  original_path TEXT,
  storage_key TEXT NOT NULL,
  current_version INTEGER NOT NULL DEFAULT 1,
  content_hash TEXT NOT NULL,
  size_bytes INTEGER NOT NULL DEFAULT 0,
  word_count INTEGER NOT NULL DEFAULT 0,
  line_count INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'active',
  sync_status TEXT NOT NULL DEFAULT 'synced',
  is_favorite INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  uploaded_at TEXT
);
CREATE INDEX IF NOT EXISTS idx_documents_updated_at  ON documents(updated_at);
CREATE INDEX IF NOT EXISTS idx_documents_status      ON documents(status);
CREATE INDEX IF NOT EXISTS idx_documents_sync_status ON documents(sync_status);
CREATE INDEX IF NOT EXISTS idx_documents_title       ON documents(title);

CREATE TABLE IF NOT EXISTS document_versions (
  id TEXT PRIMARY KEY,
  document_id TEXT NOT NULL,
  version INTEGER NOT NULL,
  storage_key TEXT NOT NULL,
  content_hash TEXT NOT NULL,
  size_bytes INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  note TEXT,
  FOREIGN KEY(document_id) REFERENCES documents(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_document_versions_doc_version
  ON document_versions(document_id, version);
CREATE INDEX IF NOT EXISTS idx_document_versions_document_id
  ON document_versions(document_id);

CREATE TABLE IF NOT EXISTS tags (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS document_tags (
  document_id TEXT NOT NULL,
  tag_id TEXT NOT NULL,
  PRIMARY KEY(document_id, tag_id),
  FOREIGN KEY(document_id) REFERENCES documents(id),
  FOREIGN KEY(tag_id) REFERENCES tags(id)
);

CREATE TABLE IF NOT EXISTS sync_jobs (
  id TEXT PRIMARY KEY,
  document_id TEXT,
  job_type TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  error_message TEXT,
  retry_count INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sync_jobs_status ON sync_jobs(status);

-- Mark v1 as applied so migrate() does not try to reapply migrate001.
CREATE TABLE IF NOT EXISTS schema_migrations (
  version    INTEGER PRIMARY KEY,
  applied_at TEXT    NOT NULL
);
INSERT INTO schema_migrations (version, applied_at) VALUES (1, '2026-01-01T00:00:00Z');

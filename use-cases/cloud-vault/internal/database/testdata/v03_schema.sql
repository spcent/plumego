-- V0.3 schema fixture: V0.1 + V0.2 import/management + V0.3 FTS.

-- V0.1 base
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
  uploaded_at TEXT,
  source_type   TEXT NOT NULL DEFAULT 'manual',
  import_job_id TEXT,
  imported_at   TEXT,
  summary       TEXT,
  heading_text  TEXT,
  review_status TEXT NOT NULL DEFAULT 'pending'
);
CREATE INDEX IF NOT EXISTS idx_documents_updated_at  ON documents(updated_at);
CREATE INDEX IF NOT EXISTS idx_documents_status      ON documents(status);
CREATE INDEX IF NOT EXISTS idx_documents_sync_status ON documents(sync_status);
CREATE INDEX IF NOT EXISTS idx_documents_title       ON documents(title);
CREATE INDEX IF NOT EXISTS idx_documents_source_type   ON documents(source_type);
CREATE INDEX IF NOT EXISTS idx_documents_import_job_id ON documents(import_job_id);
CREATE INDEX IF NOT EXISTS idx_documents_review_status ON documents(review_status);
CREATE INDEX IF NOT EXISTS idx_documents_is_favorite   ON documents(is_favorite);

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
  created_at TEXT NOT NULL,
  color TEXT,
  source TEXT
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

-- V0.2 import/management
CREATE TABLE IF NOT EXISTS import_jobs (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  source_path TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  total_count INTEGER NOT NULL DEFAULT 0,
  processed_count INTEGER NOT NULL DEFAULT 0,
  success_count INTEGER NOT NULL DEFAULT 0,
  failed_count INTEGER NOT NULL DEFAULT 0,
  skipped_count INTEGER NOT NULL DEFAULT 0,
  error_message TEXT,
  started_at TEXT,
  completed_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_import_jobs_status ON import_jobs(status);

CREATE TABLE IF NOT EXISTS import_job_items (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL,
  file_path TEXT NOT NULL,
  document_id TEXT,
  status TEXT NOT NULL DEFAULT 'pending',
  error_message TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(job_id) REFERENCES import_jobs(id)
);
CREATE INDEX IF NOT EXISTS idx_import_job_items_job_id ON import_job_items(job_id);
CREATE INDEX IF NOT EXISTS idx_import_job_items_status ON import_job_items(status);

CREATE TABLE IF NOT EXISTS document_metadata (
  id TEXT PRIMARY KEY,
  document_id TEXT NOT NULL UNIQUE,
  headings TEXT,
  code_languages TEXT,
  code_block_count INTEGER NOT NULL DEFAULT 0,
  link_count INTEGER NOT NULL DEFAULT 0,
  image_count INTEGER NOT NULL DEFAULT 0,
  extracted_at TEXT NOT NULL,
  FOREIGN KEY(document_id) REFERENCES documents(id)
);
CREATE INDEX IF NOT EXISTS idx_document_metadata_doc_id ON document_metadata(document_id);

-- V0.3 full-text search
CREATE VIRTUAL TABLE IF NOT EXISTS document_fts USING fts5(
  document_id UNINDEXED,
  title,
  original_path,
  summary,
  headings,
  content,
  tokenize = 'unicode61'
);

CREATE TABLE IF NOT EXISTS document_index_status (
  document_id     TEXT PRIMARY KEY,
  content_hash    TEXT NOT NULL,
  indexed_version INTEGER NOT NULL DEFAULT 0,
  status          TEXT NOT NULL DEFAULT 'pending',
  error_message   TEXT,
  indexed_at      TEXT,
  created_at      TEXT NOT NULL,
  updated_at      TEXT NOT NULL,
  FOREIGN KEY(document_id) REFERENCES documents(id)
);
CREATE INDEX IF NOT EXISTS idx_document_index_status_status
  ON document_index_status(status);

CREATE TABLE IF NOT EXISTS search_history (
  id           TEXT PRIMARY KEY,
  query        TEXT NOT NULL,
  filters_json TEXT,
  result_count INTEGER NOT NULL DEFAULT 0,
  created_at   TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_search_history_created_at
  ON search_history(created_at DESC);

-- Mark v1, v2, v3 as applied.
CREATE TABLE IF NOT EXISTS schema_migrations (
  version    INTEGER PRIMARY KEY,
  applied_at TEXT    NOT NULL
);
INSERT INTO schema_migrations (version, applied_at) VALUES (1, '2026-01-01T00:00:00Z');
INSERT INTO schema_migrations (version, applied_at) VALUES (2, '2026-01-01T00:00:00Z');
INSERT INTO schema_migrations (version, applied_at) VALUES (3, '2026-01-01T00:00:00Z');

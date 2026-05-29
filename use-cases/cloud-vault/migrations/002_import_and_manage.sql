-- V0.2 import and management schema changes
-- Applied programmatically via internal/database/migrate.go (migration version 2).
-- This file is a human-readable reference only.

-- New columns on documents table
ALTER TABLE documents ADD COLUMN source_type   TEXT NOT NULL DEFAULT 'manual';
ALTER TABLE documents ADD COLUMN import_job_id TEXT;
ALTER TABLE documents ADD COLUMN imported_at   TEXT;
ALTER TABLE documents ADD COLUMN summary       TEXT;
ALTER TABLE documents ADD COLUMN heading_text  TEXT;
ALTER TABLE documents ADD COLUMN review_status TEXT NOT NULL DEFAULT 'pending';

-- New columns on tags table
ALTER TABLE tags ADD COLUMN color  TEXT;
ALTER TABLE tags ADD COLUMN source TEXT;

-- Import jobs
CREATE TABLE IF NOT EXISTS import_jobs (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  source_path TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',  -- pending|running|paused|done|failed|cancelled
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

-- Import job items (one row per .md file scanned)
CREATE TABLE IF NOT EXISTS import_job_items (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL,
  file_path TEXT NOT NULL,
  document_id TEXT,
  status TEXT NOT NULL DEFAULT 'pending',  -- pending|success|skipped|failed
  error_message TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(job_id) REFERENCES import_jobs(id)
);
CREATE INDEX IF NOT EXISTS idx_import_job_items_job_id ON import_job_items(job_id);
CREATE INDEX IF NOT EXISTS idx_import_job_items_status ON import_job_items(status);

-- Rich extracted metadata for imported documents
CREATE TABLE IF NOT EXISTS document_metadata (
  id TEXT PRIMARY KEY,
  document_id TEXT NOT NULL UNIQUE,
  headings TEXT,        -- JSON: [{level, text}]
  code_languages TEXT,  -- JSON: ["go", "bash"]
  code_block_count INTEGER NOT NULL DEFAULT 0,
  link_count INTEGER NOT NULL DEFAULT 0,
  image_count INTEGER NOT NULL DEFAULT 0,
  extracted_at TEXT NOT NULL,
  FOREIGN KEY(document_id) REFERENCES documents(id)
);
CREATE INDEX IF NOT EXISTS idx_document_metadata_doc_id ON document_metadata(document_id);

-- New indexes on documents for new filter columns
CREATE INDEX IF NOT EXISTS idx_documents_source_type   ON documents(source_type);
CREATE INDEX IF NOT EXISTS idx_documents_import_job_id ON documents(import_job_id);
CREATE INDEX IF NOT EXISTS idx_documents_review_status ON documents(review_status);
CREATE INDEX IF NOT EXISTS idx_documents_is_favorite   ON documents(is_favorite);

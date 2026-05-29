-- V0.3 full-text search: FTS5 index, index status tracking, search history.

-- FTS5 virtual table for full-text search.
-- document_id is UNINDEXED (stored, not indexed).
CREATE VIRTUAL TABLE IF NOT EXISTS document_fts USING fts5(
  document_id UNINDEXED,
  title,
  original_path,
  summary,
  headings,
  content,
  tokenize = 'unicode61'
);

-- Tracks indexing state per document.
CREATE TABLE IF NOT EXISTS document_index_status (
  document_id      TEXT PRIMARY KEY,
  content_hash     TEXT NOT NULL,
  indexed_version  INTEGER NOT NULL DEFAULT 0,
  status           TEXT NOT NULL DEFAULT 'pending',
  error_message    TEXT,
  indexed_at       TEXT,
  created_at       TEXT NOT NULL,
  updated_at       TEXT NOT NULL,
  FOREIGN KEY(document_id) REFERENCES documents(id)
);

CREATE INDEX IF NOT EXISTS idx_document_index_status_status
  ON document_index_status(status);

-- Stores recent search queries for history display.
CREATE TABLE IF NOT EXISTS search_history (
  id           TEXT PRIMARY KEY,
  query        TEXT NOT NULL,
  filters_json TEXT,
  result_count INTEGER NOT NULL DEFAULT 0,
  created_at   TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_search_history_created_at
  ON search_history(created_at DESC);
